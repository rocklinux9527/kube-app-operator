package define

import (
    "context"
    "fmt"
    appsv1alpha1 "github.com/k8s/kube-app-operator/api/v1alpha1"
    "github.com/k8s/kube-app-operator/internal/pkg/utils"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/util/intstr"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "time"

    // 添加日志依赖
    logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type ServiceInfo struct {
    Name       string            `json:"name"`
    Namespace  string            `json:"namespace"`
    ClusterIP  string            `json:"cluster_ip"`
    Type       string            `json:"type"`
    Ports      []string           `json:"ports"`
    Selector   string `json:"selector"`
    CreatedAt  string            `json:"created_at"`
    Age        string            `json:"age"`
}

// 创建日志记录器
var log_svc = logf.Log.WithName("service-creator")





// 新增：DeleteService 删除对应的 Service（当 enableService == false 时调用）

func DeleteService(ctx context.Context, cli client.Client, KubeApp *appsv1alpha1.KubeApp, namespace string) error {
    name := KubeApp.Spec.Service.Name
    if name == "" {
        name = KubeApp.Name
    }
    svc := &corev1.Service{}
    svc.SetName(name)
    svc.SetNamespace(namespace)
    return utils.DeleteIfExists(ctx, cli, svc)
}



// NewService 创建 Kubernetes Service
func NewService(KubeApp *appsv1alpha1.KubeApp, namespace string) (*corev1.Service, error) {
    // 记录开始创建 Service 的日志
    log_svc.Info("开始创建 Service","KubeApp名称", KubeApp.Name, "命名空间", namespace)

    // 1. 参数有效性检查
    if err := validateServiceParams(KubeApp); err != nil {
        log_svc.Error(err, "Service 参数验证失败", "KubeApp名称", KubeApp.Name)
        return nil, err
    }

    // 2. 服务规格检查
    if err := validateServiceSpec(KubeApp.Spec.Service); err != nil {
        log_svc.Error(err, "Service 规格验证失败", "KubeApp名称", KubeApp.Name)
        return nil, err
    }

    // 3. 创建 Service 对象
    service := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:      KubeApp.Spec.Service.Name,
            Namespace: namespace,
            // 添加额外的标签和注解
            Labels:      utils.MergeMaps(KubeApp.Labels, map[string]string{"managed-by": "KubeApp-operator"}),
            Annotations: KubeApp.Annotations,
        },
        Spec: corev1.ServiceSpec{
            // 选择器使用 Deployment 的标签
            Selector: map[string]string{"app": KubeApp.Spec.Deployment.Name},
            Ports: []corev1.ServicePort{
                {
                    // 端口配置
                    Port:       utils.NormalizePort(KubeApp.Spec.Service.Port),
                    TargetPort: intstr.FromInt(int(utils.NormalizePort(KubeApp.Spec.Service.TargetPort))),
                    // 可选：添加协议和名称
                    Protocol: corev1.ProtocolTCP,
                    Name:     fmt.Sprintf("%s-port", KubeApp.Spec.Service.Name),
                },
            },
            // 可选：服务类型
            Type: determineServiceType(KubeApp.Spec.Service),
        },
    }

    log_svc.Info("Service 创建成功", "名称", service.Name,"命名空间", service.Namespace)

    return service, nil
}

// validateServiceParams 验证 KubeApp 参数
func validateServiceParams(KubeApp *appsv1alpha1.KubeApp) error {
    if KubeApp == nil {
        return fmt.Errorf("KubeApp 对象不能为空")
    }

    if KubeApp.Spec.Service == nil {
        return fmt.Errorf("KubeApp 的 Service 规格不能为空")
    }

    if KubeApp.Spec.Deployment == nil {
        return fmt.Errorf("KubeApp 的 Deployment 规格不能为空")
    }

    return nil
}

// validateServiceSpec 验证服务规格
func validateServiceSpec(serviceSpec *appsv1alpha1.ServiceSpec) error {
    if serviceSpec == nil {
        return fmt.Errorf("Service 规格不能为空")
    }

    if serviceSpec.Name == "" {
        return fmt.Errorf("Service 名称不能为空")
    }

    // 端口范围检查
    if err := utils.ValidatePort(serviceSpec.Port); err != nil {
        return fmt.Errorf("Service 端口验证失败: %v", err)
    }

    if err := utils.ValidatePort(serviceSpec.TargetPort); err != nil {
        return fmt.Errorf("Service 目标端口验证失败: %v", err)
    }

    return nil
}


// determineServiceType 根据规格确定服务类型
func determineServiceType(serviceSpec *appsv1alpha1.ServiceSpec) corev1.ServiceType {
    // 默认使用 ClusterIP
    if serviceSpec.Type == "" {
        log_svc.V(1).Info("未指定服务类型，使用默认 ClusterIP")
        return corev1.ServiceTypeClusterIP
    }

    // 转换服务类型
    switch serviceSpec.Type {
    case "NodePort":
        return corev1.ServiceTypeNodePort
    case "LoadBalancer":
        return corev1.ServiceTypeLoadBalancer
    default:
        log_svc.V(1).Info("未识别的服务类型，使用默认 ClusterIP","指定类型", serviceSpec.Type)
        return corev1.ServiceTypeClusterIP
    }
}

// 使用 controller-runtime client 查询 service 查询接口

func ListServices(namespace string) ([]ServiceInfo, error) {
    if GlobalClient == nil{
        return nil, fmt.Errorf("k8s client 未初始化，请先调用 custom.Init()")
    }

    var svcList corev1.ServiceList
    if err := GlobalClient.List(context.Background(), &svcList, client.InNamespace(namespace)); err != nil {
        return nil, err
    }

    var result []ServiceInfo
    for _, svc := range svcList.Items {
        // +8 时区
        loc := time.FixedZone("CST", 8*3600)
        createdAt := svc.CreationTimestamp.Time.In(loc).Format("2006-01-02 15:04:05")

        // 计算 AGE
        age := utils.FormatSvcAge(time.Since(svc.CreationTimestamp.Time))

        // 拼接 Ports
        var ports []string
        for _, p := range svc.Spec.Ports {
            if p.NodePort > 0 {
                ports = append(ports, fmt.Sprintf("%d:%d/%s", p.Port, p.NodePort, p.Protocol))
            } else {
                ports = append(ports, fmt.Sprintf("%d/%s", p.Port, p.Protocol))
            }
        }
        labelStr := utils.FormatLabels(svc.Spec.Selector)
        result = append(result, ServiceInfo{
            Name:      svc.Name,
            Namespace: svc.Namespace,
            ClusterIP: svc.Spec.ClusterIP,
            Type:      string(svc.Spec.Type),
            Ports:     ports,
            Selector:  labelStr,
            CreatedAt: createdAt,
            Age:     age,
        })
    }
    return result, nil
}
