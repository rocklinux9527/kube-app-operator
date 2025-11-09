package define

import (
    "context"
    "errors"
    "fmt"
    appsv1alpha1 "github.com/k8s/kube-app-operator/api/v1alpha1"
    networkingv1 "k8s.io/api/networking/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "net/url"
    "regexp"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "strings"
    "time"

    "github.com/k8s/kube-app-operator/internal/pkg/utils"
    // 添加日志依赖
    logf "sigs.k8s.io/controller-runtime/pkg/log"
    //
)

// IngressRoute 单条路由信息
type IngressRoute struct {
    Host     string `json:"host"`
    Path     string `json:"path"`
    Service  string `json:"service"`
    Port     int32  `json:"port"`
    Protocol string `json:"protocol"`
}

// IngressInfo 查询定义返回的字段
type IngressInfo struct {
    Name             string         `json:"name"`
    Namespace        string         `json:"namespace"`
    IngressClassName string         `json:"ingress_class_name"`
    Annotations      string         `json:"annotations"`
    CreatedAt        string         `json:"created_at"`
    Age              string         `json:"age"`
    Routes           []IngressRoute `json:"routes"`
}



// 创建日志记录器
var log_ing = logf.Log.WithName("ingress-creator")


var (
    // 正则表达式参考自 Kubernetes 的 RFC1123 子域名规则
    rfc1123SubdomainRegex = regexp.MustCompile(`^([a-z0-9]([-a-z0-9]*[a-z0-9])?)(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
)



// DeleteIngress 删除 Ingress（当 EnableIngress == false 时调用）
func DeleteIngress(ctx context.Context, cli client.Client, KubeApp *appsv1alpha1.KubeApp, namespace string) error {
    name := KubeApp.Name
    if KubeApp.Spec.Ingress != nil && KubeApp.Spec.Ingress.Name != "" {
        name = KubeApp.Spec.Ingress.Name
    }
    ing := &networkingv1.Ingress{}
    ing.SetName(name)
    ing.SetNamespace(namespace)

    return utils.DeleteIfExists(ctx, cli, ing)
}



// NewIngress 创建 Kubernetes Ingress


/* ingress .spec.ingressClassName logic
1.校验合法性
2.设置到 spec.ingressClassName
3.同时更新或覆盖注解 kubernetes.io/ingress.class，确保一致
4.如果用户没有设置：不设置 spec.ingressClassName（保持为 nil）保留默认注解 kubernetes.io/ingress.class（兼容老的 Ingress Controller）
*/

func NewIngress(KubeApp *appsv1alpha1.KubeApp, namespace string) (*networkingv1.Ingress, error) {
    log_ing.Info("开始创建 Ingress", "KubeApp名称", KubeApp.Name, "命名空间", namespace)

    // 1. 参数有效性检查
    if err := validateIngressParams(KubeApp); err != nil {
        log_ing.Error(err, "Ingress 参数验证失败", "KubeApp名称", KubeApp.Name)
        return nil, err
    }

    // 2. Ingress 规格检查
    if err := validateIngressSpec(KubeApp.Spec.Ingress); err != nil {
        log_ing.Error(err, "Ingress 规格验证失败", "KubeApp名称", KubeApp.Name)
        return nil, err
    }

    // 3. 构建 IngressSpec（暂不设置 ingressClassName）
    ingressSpec := networkingv1.IngressSpec{
        Rules: []networkingv1.IngressRule{
            {
                Host: normalizeHost(KubeApp.Spec.Ingress.Host),
                IngressRuleValue: networkingv1.IngressRuleValue{
                    HTTP: &networkingv1.HTTPIngressRuleValue{
                        Paths: []networkingv1.HTTPIngressPath{
                            {
                                Path:     normalizePath(KubeApp.Spec.Ingress.Path),
                                PathType: getPathType(KubeApp.Spec.Ingress.PathType),
                                Backend: networkingv1.IngressBackend{
                                    Service: &networkingv1.IngressServiceBackend{
                                        Name: KubeApp.Spec.Ingress.ServiceName,
                                        Port: networkingv1.ServiceBackendPort{
                                            Number: utils.NormalizePort(KubeApp.Spec.Ingress.ServicePort),
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    // 4. 处理 ingressClassName（仅当显式指定时）
    cls := KubeApp.Spec.Ingress.IngressClassName
    if cls != "" {
        if err := ValidateIngressClassName(cls); err != nil {
            log_ing.Error(err, "无效的 ingressClassName", "值", cls)
            return nil, fmt.Errorf("invalid ingressClassName %q: %w", cls, err)
        }
        ingressSpec.IngressClassName = ptr(cls)
    }

    // 5. 注解处理：保持字段和注解一致性（清洗用户注解）
    cleanedAnnotations := sanitizeIngressAnnotations(KubeApp.Annotations, cls)

    // 6. 构建 Ingress 对象
    ingress := &networkingv1.Ingress{
        ObjectMeta: metav1.ObjectMeta{
            Name:        KubeApp.Name,
            Namespace:   namespace,
            Labels:      utils.MergeMaps(KubeApp.Labels, map[string]string{"managed-by": "KubeApp-operator"}),
            Annotations: cleanedAnnotations,
        },
        Spec: ingressSpec,
    }

    log_ing.Info("Ingress 创建成功", "名称", ingress.Name, "命名空间", ingress.Namespace)
    return ingress, nil
}


// 冲突处理逻辑
// sanitizeIngressAnnotations 处理注解与 ingressClassName 的一致性
func sanitizeIngressAnnotations(userAnnotations map[string]string, ingressClassName string) map[string]string {
    annotations := cloneMap(userAnnotations)

    // 用户未设置 ingressClassName，保留原始注解
    if ingressClassName == "" {
        return annotations
    }

    // 检查旧注解是否冲突
    if val, exists := annotations["kubernetes.io/ingress.class"]; exists && val != ingressClassName {
        log_ing.Info("注解 ingress.class 与 ingressClassName 不一致，已自动修正", "注解", val, "字段", ingressClassName)
        delete(annotations, "kubernetes.io/ingress.class")
    }

    // 设置统一注解（兼容旧 controller）
    annotations["kubernetes.io/ingress.class"] = ingressClassName

    return annotations
}

// cloneMap 浅拷贝 map[string]string
func cloneMap(m map[string]string) map[string]string {
    out := make(map[string]string, len(m))
    for k, v := range m {
        out[k] = v
    }
    return out
}

// validateIngressParams 验证 KubeApp 参数
func validateIngressParams(KubeApp *appsv1alpha1.KubeApp) error {
    if KubeApp == nil {
        return fmt.Errorf("KubeApp 对象不能为空")
    }

    if KubeApp.Spec.Ingress == nil {
        return fmt.Errorf("KubeApp 的 Ingress 规格不能为空")
    }

    return nil
}



func ValidateIngressClassName(name string) error {
    if name == "" {
        return errors.New("IngressClassName cannot be empty")
    }
    if len(name) > 253 {
        return fmt.Errorf("IngressClassName too long: %d characters (max 253)", len(name))
    }
    if !rfc1123SubdomainRegex.MatchString(name) {
        return fmt.Errorf("IngressClassName %q is invalid: must match RFC 1123 subdomain format (lowercase alphanumeric, '-', '.', and must start/end with alphanumeric)", name)
    }
    return nil
}



// validateIngressSpec 验证 Ingress 规格
func validateIngressSpec(ingressSpec *appsv1alpha1.IngressSpec) error {
    if ingressSpec == nil {
        return fmt.Errorf("Ingress 规格不能为空")
    }

    // 验证 Host
    if err := validateHost(ingressSpec.Host); err != nil {
        return fmt.Errorf("Ingress Host 验证失败: %v", err)
    }

    // 验证 ServiceName
    if ingressSpec.ServiceName == "" {
        return fmt.Errorf("Ingress ServiceName 不能为空")
    }

    // 验证 ServicePort
    if err := utils.ValidatePort(ingressSpec.ServicePort); err != nil {
        return fmt.Errorf("Ingress ServicePort 验证失败: %v", err)
    }

    return nil
}

// validateHost 验证主机名
func validateHost(host string) error {
    if host == "" {
        return fmt.Errorf("主机名不能为空")
    }

    // 使用 URL 解析验证主机名
    u, err := url.Parse("http://" + host)
    if err != nil {
        return fmt.Errorf("无效的主机名格式")
    }

    // 正则表达式验证主机名
    hostRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
    if !hostRegex.MatchString(u.Hostname()) {
        return fmt.Errorf("主机名格式不正确")
    }

    return nil
}



// 对任何类型（比如 string、int、bool 等）返回指针

func ptr[T any](v T) *T {
    return &v
}

// normalizeHost 规范化主机名
func normalizeHost(host string) string {
    if host == "" {
        log_ing.V(1).Info("未指定主机名，使用默认值loclhjosts")
        return "localhost"
    }
    return host
}

// normalizePath 规范化路径
func normalizePath(path string) string {
    if path == "" {
        log_ing.V(1).Info("未指定路径，使用默认根/路径")
        return "/"
    }
    return path
}


// getPathType 获取路径类型
func getPathType(pathType networkingv1.PathType) *networkingv1.PathType {
    switch pathType {
    case networkingv1.PathTypeExact:
         pt := networkingv1.PathTypeExact
         return &pt
    case networkingv1.PathTypeImplementationSpecific:
         pt := networkingv1.PathTypeImplementationSpecific
         return &pt
    case networkingv1.PathTypePrefix:
         pt := networkingv1.PathTypePrefix
         return &pt
    default:
         pt := networkingv1.PathTypePrefix
         return &pt
    }
}

// getDefaultIngressAnnotations 获取默认 Ingress 注解
func getDefaultIngressAnnotations() map[string]string {
    return map[string]string{
        "kubernetes.io/ingress.class": "nginx",
        // 可以添加更多默认注解
    }
}


// mergeAnnotations 合并注解，以后面的为准
func mergeAnnotations(annotationSets ...map[string]string) map[string]string {
    result := make(map[string]string)
    for _, annotations := range annotationSets {
        for k, v := range annotations {
            result[k] = v
        }
    }
    return result
}

// 使用 controller-runtime client 查询 Ingress 查询接口

func formatAnnotations(ann map[string]string) string {
    if len(ann) == 0 {
        return ""
    }
    pairs := make([]string, 0, len(ann))
    for k, v := range ann {
        pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
    }
    return strings.Join(pairs, ", ")
}



func ListIngress(namespace string) ([]IngressInfo, error) {
    if GlobalClient == nil {
        return nil, fmt.Errorf("k8s client 未初始化，请先调用 custom.Init()")
    }

    var ingList networkingv1.IngressList
    if err := GlobalClient.List(context.Background(), &ingList, client.InNamespace(namespace)); err != nil {
        return nil, fmt.Errorf("获取 Ingress 列表失败: %v", err)
    }

    loc := time.FixedZone("CST", 8*3600)
    var result []IngressInfo

    for _, ing := range ingList.Items {
        createdAt := ing.CreationTimestamp.Time.In(loc).Format("2006-01-02 15:04:05")
        age := formatAge(ing.CreationTimestamp.Time)
        annotations := formatAnnotations(ing.Annotations)

        // 记录哪些 Host 支持 TLS
        tlsHosts := make(map[string]bool)
        for _, tls := range ing.Spec.TLS {
            for _, h := range tls.Hosts {
                tlsHosts[h] = true
            }
        }

        var routes []IngressRoute

        // 遍历规则
        for _, rule := range ing.Spec.Rules {
            host := rule.Host
            protocol := "http"
            if tlsHosts[host] {
                protocol = "https"
            }

            if rule.HTTP != nil {
                for _, path := range rule.HTTP.Paths {
                    svcName := ""
                    port := int32(0)
                    if path.Backend.Service != nil {
                        svcName = path.Backend.Service.Name
                        port = path.Backend.Service.Port.Number
                    }

                    routes = append(routes, IngressRoute{
                        Host:     host,
                        Path:     path.Path,
                        Service:  svcName,
                        Port:     port,
                        Protocol: protocol,
                    })
                }
            }
        }

        // 没有规则时补空占位
        if len(routes) == 0 {
            routes = append(routes, IngressRoute{
                Host:     "(无路由规则)",
                Path:     "-",
                Service:  "-",
                Port:     0,
                Protocol: "-",
            })
        }

        result = append(result, IngressInfo{
            Name:             ing.Name,
            Namespace:        ing.Namespace,
            IngressClassName: safeString(ing.Spec.IngressClassName),
            Annotations:      annotations,
            CreatedAt:        createdAt,
            Age:              age,
            Routes:           routes,
        })
    }

    return result, nil
}

func safeString(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}
