package define

import (
    "context"
    "fmt"
    appsv1alpha1 "github.com/k8s/kube-app-operator/api/v1alpha1"
    "github.com/k8s/kube-app-operator/internal/pkg/utils"
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/util/intstr"
    "strings"
    "time"

    // "k8s.io/apimachinery/pkg/api/errors"
    "sigs.k8s.io/controller-runtime/pkg/client"
    // 添加日志依赖
    logf "sigs.k8s.io/controller-runtime/pkg/log"
)


type DeploymentInfo struct {
    AppName   string            `json:"app_name"`
    Namespace string            `json:"namespace"`
    Replicas  int32             `json:"replicas"`
    Ready     string            `json:"ready"`
    UpToDate  int32             `json:"up_to_date"`
    Available int32             `json:"available"`
    Image     []string          `json:"image"`
    DeployEnv string `json:"deploy_env"`
    Ports     []int32           `json:"ports"`
    CreatedAt string            `json:"created_at"`
    UpdatedAt string            `json:"updated_at"`
    Age       string            `json:"age"`
}
// 创建日志记录器
var log_dp = logf.Log.WithName("deployment-creator")


//  Deployment 根据 KubeApp 自定义资源false delete 删除deployment  

func DeleteDeployment(ctx context.Context, cli client.Client, KubeApp *appsv1alpha1.KubeApp, namespace string) error {
    name := KubeApp.Spec.Deployment.Name
    if name == "" {
        name = KubeApp.Name
    }
    dep := &appsv1.Deployment{}
    dep.SetName(name)
    dep.SetNamespace(namespace)
    return utils.DeleteIfExists(ctx, cli, dep)
}


// NewDeployment 根据 KubeApp 自定义资源创建 Kubernetes Deployment 优化版本

func NewDeployment(KubeApp *appsv1alpha1.KubeApp, namespace string) (*appsv1.Deployment, error) {
    // 0. 空指针保护必须在访问字段前做
    if KubeApp == nil {
        log_dp.Error(fmt.Errorf("KubeApp 对象为空"), "参数校验失败")
        return nil, fmt.Errorf("KubeApp 对象不能为空")
    }

    log_dp.Info("开始创建 Deployment", "KubeApp名称", KubeApp.Name, "命名空间", namespace)

    if KubeApp.Spec.Deployment == nil {
        log_dp.Error(fmt.Errorf("Deployment 规格为空"), "参数校验失败", "KubeApp名称", KubeApp.Name)
        return nil, fmt.Errorf("KubeApp 的 Deployment 规格不能为空")
    }

    if err := validateDeploymentSpec(KubeApp.Spec.Deployment); err != nil {
        log_dp.Error(err, "Deployment 规格验证失败", "KubeApp名称", KubeApp.Name)
        return nil, err
    }

    // 设置副本数
    replicas := int32(1)
    if KubeApp.Spec.Deployment.Replicas != nil && *KubeApp.Spec.Deployment.Replicas > 0 {
        replicas = *KubeApp.Spec.Deployment.Replicas
    }
    log_dp.V(1).Info("设置副本数", "副本数", replicas, "KubeApp名称", KubeApp.Name)

    // terminationGracePeriodSeconds
    var terminationGracePeriodSeconds *int64
    if KubeApp.Spec.Deployment.TerminationGracePeriodSeconds != nil {
        terminationGracePeriodSeconds = KubeApp.Spec.Deployment.TerminationGracePeriodSeconds
        log_dp.Info("设置 terminationGracePeriodSeconds", "值", *terminationGracePeriodSeconds)
    } else {
        log_dp.Info("未设置 terminationGracePeriodSeconds，使用 Kubernetes 默认值")
    }

    // imagePullSecrets
    var imagePullSecrets []corev1.LocalObjectReference
    if len(KubeApp.Spec.Deployment.ImagePullSecrets) > 0 {
        imagePullSecrets = convertImagePullSecrets(KubeApp.Spec.Deployment.ImagePullSecrets)
        log_dp.Info("配置 imagePullSecrets", "数量", len(imagePullSecrets))
    } else {
        log_dp.Info("未配置 imagePullSecrets，使用默认拉取策略")
    }

    // dnsConfig
    var dnsConfig *corev1.PodDNSConfig
    if KubeApp.Spec.Deployment.DNSConfig != nil {
        dnsConfig = KubeApp.Spec.Deployment.DNSConfig
        log_dp.Info("配置自定义 DNSConfig", "内容", dnsConfig)
    } else {
        log_dp.Info("未配置 DNSConfig，使用 Kubernetes 默认 DNS 策略")
    }

    // volumes
    var volumes []corev1.Volume
    if len(KubeApp.Spec.Deployment.Volumes) > 0 {
        volumes = convertVolumesToK8sVolumes(KubeApp.Spec.Deployment.Volumes)
        log_dp.Info("配置 Volumes", "数量", len(volumes))
    }

    // 构建 Deployment 对象
    deployment := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:        KubeApp.Spec.Deployment.Name,
            Namespace:   namespace,
            Labels:      utils.MergeMaps(KubeApp.Labels, map[string]string{"managed-by": "KubeApp-operator"}),
            Annotations: KubeApp.Annotations,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{
                    "app": KubeApp.Spec.Deployment.Name,
                },
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "app": KubeApp.Spec.Deployment.Name,
                    },
                },
                Spec: corev1.PodSpec{
                    Containers:                    []corev1.Container{prepareContainer(KubeApp.Spec.Deployment)},
                    Volumes:                       volumes,
                    NodeSelector:                  KubeApp.Spec.Deployment.NodeSelector,
                    TerminationGracePeriodSeconds: terminationGracePeriodSeconds,
                    ImagePullSecrets:              imagePullSecrets,
                    Affinity:                       prepareAffinity(KubeApp.Spec.Deployment),
                    DNSConfig:                     dnsConfig,
                },
            },
            Strategy: prepareDeploymentStrategy(),
        },
    }

    log_dp.Info("Deployment 创建成功", "名称", deployment.Name, "命名空间", deployment.Namespace)
    return deployment, nil
}

// validateDeploymentSpec 验证 Deployment 规格的必填字段
func validateDeploymentSpec(spec *appsv1alpha1.DeploymentSpec) error {
    if spec == nil {
        log_dp.Error(fmt.Errorf("Deployment 规格为空"), "规格验证失败")
        return fmt.Errorf("Deployment 规格不能为空")
    }

    if spec.Name == "" {
        log_dp.Error(fmt.Errorf("Deployment 名称为空"), "名称验证失败")
        return fmt.Errorf("Deployment 名称不能为空")
    }

    if spec.Image == "" {
        log_dp.Error(fmt.Errorf("容器镜像为空"), "镜像验证失败", "Deployment名称", spec.Name)
        return fmt.Errorf("容器镜像不能为空")
    }

    return nil
}




// convertImagePullSecrets 指定 PullSecrets
func convertImagePullSecrets(secrets []corev1.LocalObjectReference) []corev1.LocalObjectReference {
    if len(secrets) == 0 {
        return nil
    }
    return secrets
}


// prepareContainer 容器常见功能配置
func prepareContainer(spec *appsv1alpha1.DeploymentSpec) corev1.Container {
    log_dp.V(1).Info("准备容器配置","容器名称", spec.Name, "镜像", spec.Image)

    container := corev1.Container{
        Name:  spec.Name,
        Image: spec.Image,
    }

    // 资源配置
    if spec.Resources != nil {
        container.Resources = *spec.Resources
        log_dp.V(1).Info("配置容器资源", "容器名称", spec.Name, "资源配置", spec.Resources)
   }

    // 端口配置
    if len(spec.Ports) > 0 {
        container.Ports = spec.Ports
        log_dp.V(1).Info("配置容器端口","容器名称", spec.Name,"端口数量", len(spec.Ports))
    }

    // 探针配置
    if spec.LivenessProbe != nil {
        container.LivenessProbe = spec.LivenessProbe
        log_dp.V(1).Info("配置存活探针", "容器名称", spec.Name)
    }

    if spec.ReadinessProbe != nil {
        container.ReadinessProbe = spec.ReadinessProbe
        log_dp.V(1).Info("配置就绪探针", "容器名称", spec.Name)
    }

    // 生命周期钩子
    if spec.Lifecycle != nil {
        container.Lifecycle = spec.Lifecycle
        log_dp.V(1).Info("配置生命周期钩子", "容器名称", spec.Name)
    }

    // 环境变量env
    container.Env = prepareEnv(spec)
    // 添加安全上下文
    // container.SecurityContext = prepareContainerSecurityContext()

    // 处理 VolumeMounts
    if len(spec.VolumeMounts) > 0 {
        var mounts []corev1.VolumeMount
        for _, vm := range spec.VolumeMounts {
            mounts = append(mounts, corev1.VolumeMount{
                Name:      vm.Name,
                MountPath: vm.MountPath,
                ReadOnly:  vm.ReadOnly,
            })
            log_dp.V(1).Info("配置卷挂载", "容器名称", spec.Name, "卷名", vm.Name, "挂载路径", vm.MountPath, "只读", vm.ReadOnly)
        }
        container.VolumeMounts = mounts
        log_dp.V(1).Info("完成所有卷挂载配置", "容器名称", spec.Name, "挂载数量", len(mounts))
    }
    return container
}


// prepareAffinity 处理 Affinity 设置逻辑

func prepareAffinity(spec *appsv1alpha1.DeploymentSpec) *corev1.Affinity {
    if spec.Affinity != nil {
        log_dp.Info("设置 Affinity")
        return spec.Affinity
    }
    log_dp.Info("未设置 Affinity，跳过")
    return nil
}



// env 设置逻辑

func prepareEnv(spec *appsv1alpha1.DeploymentSpec) []corev1.EnvVar {
    if len(spec.Env) > 0 {
        log_dp.V(1).Info("配置环境变量", "容器名称", spec.Name, "变量数量", len(spec.Env))
        return spec.Env
    }
    log_dp.V(1).Info("未配置环境变量", "容器名称", spec.Name)
    return nil
}


// preparePodSecurityContext 设置 Pod 安全上下文
func _preparePodSecurityContext() *corev1.PodSecurityContext {
    log_dp.V(1).Info("配置 Pod 安全上下文")
    return &corev1.PodSecurityContext{
        RunAsNonRoot: boolPtr(true),
    }
}



// prepareContainerSecurityContext 设置容器安全上下文
func _prepareContainerSecurityContext() *corev1.SecurityContext {
    log_dp.V(1).Info("配置容器安全上下文")
    return &corev1.SecurityContext{
        AllowPrivilegeEscalation: boolPtr(false),
        ReadOnlyRootFilesystem:   boolPtr(true),
    }
}

// prepareDeploymentStrategy 配置部署策略
func prepareDeploymentStrategy() appsv1.DeploymentStrategy {
    log_dp.V(1).Info("配置部署策略")
    return appsv1.DeploymentStrategy{
        Type: appsv1.RollingUpdateDeploymentStrategyType,
        RollingUpdate: &appsv1.RollingUpdateDeployment{
            MaxUnavailable: intOrStrPtr(25),
            MaxSurge:       intOrStrPtr(25),
        },
    }
}


// 辅助函数：创建布尔指针
func boolPtr(b bool) *bool {
    return &b
}

// 辅助函数：创建整数或字符串指针
func intOrStrPtr(val int) *intstr.IntOrString {
    return &intstr.IntOrString{
        Type:   intstr.Int,
        IntVal: int32(val),
    }
}


// Volumes auth tyoe logic her 自动识别卷类型不需要用户指定

func convertVolumesToK8sVolumes(volumeConfigs []appsv1alpha1.VolumeConfig) []corev1.Volume {
    var volumes []corev1.Volume

    for _, volConfig := range volumeConfigs {
        volume := corev1.Volume{
            Name: volConfig.Name,
        }
       log_dp.V(1).Info("开始Deployment处理卷", "name", volConfig.Name)
        switch {
        case volConfig.PersistentVolumeClaim != nil:
            volume.VolumeSource.PersistentVolumeClaim = volConfig.PersistentVolumeClaim
            log_dp.V(1).Info("使用 pvc类型卷", "name", volConfig.Name)
        case volConfig.ConfigMap != nil:
            volume.VolumeSource.ConfigMap = volConfig.ConfigMap
            log_dp.V(1).Info("使用 ConfigMap 类型卷", "name", volConfig.Name)
        case volConfig.Secret != nil:
            volume.VolumeSource.Secret = volConfig.Secret
            log_dp.V(1).Info("使用 Secret 类型卷", "name", volConfig.Name)
        case volConfig.EmptyDir != nil:
            volume.VolumeSource.EmptyDir = volConfig.EmptyDir
            log_dp.V(1).Info("使用 EmptyDir 类型卷", "name", volConfig.Name)
        case volConfig.HostPath != nil:
            volume.VolumeSource.HostPath = volConfig.HostPath
            log_dp.V(1).Info("使用 HostPath 类型卷", "name", volConfig.Name)
        case volConfig.NFS != nil:
            volume.VolumeSource.NFS = volConfig.NFS
            log_dp.V(1).Info("使用 NFS 类型卷", "name", volConfig.Name, "server", volConfig.NFS.Server, "path", volConfig.NFS.Path)
        default:
            log_dp.Info("未识别的卷类型，跳过", "卷名称", volConfig.Name)
            continue
        }

        volumes = append(volumes, volume)
    }
    log_dp.V(1).Info("所有卷处理完成", "总数量", len(volumes))
    return volumes
}


// 使用 controller-runtime client 查询 Deployment 查询接口

func ListDeployments(namespace string) ([]DeploymentInfo, error) {
    if GlobalClient == nil {
        return nil, fmt.Errorf("k8s client 未初始化")
    }

    var deployList appsv1.DeploymentList
    if err := GlobalClient.List(context.Background(), &deployList, client.InNamespace(namespace)); err != nil {
        return nil, err
    }

    loc, _ := time.LoadLocation("Asia/Shanghai")
    var result []DeploymentInfo

    for _, d := range deployList.Items {
        replicas := int32(0)
        if d.Spec.Replicas != nil {
            replicas = *d.Spec.Replicas
        }
        ready := fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, replicas)

        // images
        var images []string
        for _, c := range d.Spec.Template.Spec.Containers {
            images = append(images, c.Image)
        }

        // env
        var envList []string
        for _, c := range d.Spec.Template.Spec.Containers {
            for _, e := range c.Env {
                envList = append(envList, fmt.Sprintf("%s=%s", e.Name, e.Value))
            }
        }
        deployEnv := strings.Join(envList, ",")

        // ports
        var ports []int32
        for _, c := range d.Spec.Template.Spec.Containers {
            for _, p := range c.Ports {
                ports = append(ports, p.ContainerPort)
            }
        }

        // created_at / updated_at
        createdAt := d.CreationTimestamp.Time
        updatedAt := createdAt
        if len(d.ManagedFields) > 0 {
            last := d.ManagedFields[len(d.ManagedFields)-1]
            if last.Time != nil {
                updatedAt = last.Time.Time
            }
        }

        result = append(result, DeploymentInfo{
            AppName:   d.Name,
            Namespace: d.Namespace,
            Replicas:  replicas,
            Ready:     ready,
            UpToDate:  d.Status.UpdatedReplicas,
            Available: d.Status.AvailableReplicas,
            Image:     images,
            DeployEnv: deployEnv,
            Ports:     ports,
            CreatedAt: createdAt.In(loc).Format("2006-01-02 15:04:05"),
            UpdatedAt: updatedAt.In(loc).Format("2006-01-02 15:04:05"),
            Age:       utils.FormatAge(createdAt),
        })
    }
    return result, nil
}

