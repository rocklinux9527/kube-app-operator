package define

import (
    "context"
    "fmt"
    appsv1alpha1 "github.com/k8s/kube-app-operator/api/v1alpha1"
    utils "github.com/k8s/kube-app-operator/internal/pkg/utils"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/utils/pointer"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "time"

    // 添加日志依赖
    logf "sigs.k8s.io/controller-runtime/pkg/log"

    v1 "k8s.io/api/core/v1"
)

// PVCInfo 定义返回结构
type PVCInfo struct {
    Name         string            `json:"name"`
    Namespace    string            `json:"namespace"`
    Status       v1.PersistentVolumeClaimPhase `json:"status"`
    Volume       string            `json:"volume"`
    Capacity     string            `json:"capacity"`
    ActualCapacity string            `json:"actual_capacity"`
    AccessModes  string            `json:"access_modes"`
    StorageClass string            `json:"storage_class"`
    CreatedAt    string            `json:"created_at"`
    Age          string            `json:"age"`
    Labels       string `json:"labels"`
}



// 创建PVC日志记录器
var log_pvc = logf.Log.WithName("pvc-creator")


// NewPvc 创建 PVC 对象结构

func NewPvc(ctx context.Context, KubeApp *appsv1alpha1.KubeApp, namespace string) (*unstructured.Unstructured, error) {
    log_pvc.Info("准备构建 PVC 对象", "KubeApp", KubeApp.Name)

    if err := validatePvcParams(KubeApp); err != nil {
        log_pvc.Error(err, "PVC 参数校验失败")
        return nil, err
    }

    pvcSpec := KubeApp.Spec.Pvc

    pvc := &unstructured.Unstructured{}
    pvc.SetGroupVersionKind(schema.GroupVersionKind{
        Group:   "",
        Version: "v1",
        Kind:    "PersistentVolumeClaim",
    })

    pvc.SetName(pvcSpec.Name)
    pvc.SetNamespace(namespace)
    pvc.SetLabels(utils.MergeMaps(KubeApp.Labels, map[string]string{"managed-by": "KubeApp-operator"}))
    pvc.SetAnnotations(KubeApp.Annotations)

    accessModes := make([]interface{}, len(pvcSpec.AccessModes))
    for i, mode := range pvcSpec.AccessModes {
        accessModes[i] = string(mode)
    }

    spec := map[string]interface{}{
        "accessModes": accessModes,
        "resources": map[string]interface{}{
            "requests": map[string]interface{}{
                "storage": pvcSpec.Storage,
            },
        },
    }
    if pvcSpec.StorageClassName != nil && *pvcSpec.StorageClassName != "" {
        spec["storageClassName"] = *pvcSpec.StorageClassName
    }
    pvc.Object["spec"] = spec
    return pvc, nil
}


// ApplyPvc 使用 controllerutil.CreateOrUpdate 方式 apply PVC

func ApplyPvc(ctx context.Context, cli client.Client, KubeApp *appsv1alpha1.KubeApp, namespace string) error {
    pvc, err := NewPvc(ctx, KubeApp, namespace)
    if err != nil {
        return fmt.Errorf("构建 PVC 对象失败: %w", err)
    }

    // 使用 Server-Side Apply
    err = cli.Patch(ctx, pvc, client.Apply, &client.PatchOptions{
        Force:        pointer.Bool(true),
        FieldManager: "KubeApp-operator",
    })

    if err != nil {
        log_pvc.Error(err, "PVC apply 失败", "PVC名称", pvc.GetName())
        return err
    }

    log_pvc.Info("PVC apply 成功", "名称", pvc.GetName(), "命名空间", namespace)
    return nil
}

// DeletePvc 删除 PVC
func DeletePvc(ctx context.Context, cli client.Client, KubeApp *appsv1alpha1.KubeApp, namespace string) error {
    name := KubeApp.Spec.Pvc.Name
    if name == "" {
        name = KubeApp.Name
    }

    pvc := &corev1.PersistentVolumeClaim{}
    pvc.SetName(name)
    pvc.SetNamespace(namespace)

    return utils.DeleteIfExists(ctx, cli, pvc)
}

// 参数校验
func validatePvcParams(KubeApp *appsv1alpha1.KubeApp) error {
    if KubeApp == nil {
        return fmt.Errorf("KubeApp 对象不能为空")
    }

    if KubeApp.Spec.Pvc == nil {
        return fmt.Errorf("KubeApp 的 Pvc 规格不能为空")
    }

    return validatePvcSpec(KubeApp.Spec.Pvc)
}

// PVC 规格校验
func validatePvcSpec(pvcSpec *appsv1alpha1.PvcSpec) error {
    if pvcSpec.Name == "" {
        return fmt.Errorf("PVC 名称不能为空")
    }

    if pvcSpec.Storage == "" {
        return fmt.Errorf("PVC Storage 不能为空")
    }

    if len(pvcSpec.AccessModes) == 0 {
        return fmt.Errorf("PVC AccessModes 不能为空")
    }

    return nil
}

// ListPVCs 获取指定命名空间下的 PVC

func ListPVCs(namespace string) ([]PVCInfo, error) {
    if GlobalClient == nil {
        return nil, fmt.Errorf("k8s client 未初始化，请先调用 custom.Init()")
    }

    var pvcList v1.PersistentVolumeClaimList
    if err := GlobalClient.List(context.Background(), &pvcList, client.InNamespace(namespace)); err != nil {
        return nil, err
    }

    var result []PVCInfo
    for _, pvc := range pvcList.Items {
        // 申请容量
        requested := pvc.Spec.Resources.Requests[v1.ResourceStorage]

        actual := pvc.Status.Capacity[v1.ResourceStorage]

        // 访问模式
        accessModes := ""
        for _, m := range pvc.Status.AccessModes {
            accessModes += string(m) + ","
        }
        if len(accessModes) > 0 {
            accessModes = accessModes[:len(accessModes)-1]
        }

        // 存储类
        storageClass := ""
        if pvc.Spec.StorageClassName != nil {
            storageClass = *pvc.Spec.StorageClassName
        }

        pvcLabels := utils.FormatLabels(pvc.Labels)
        // AGE
        age := time.Since(pvc.CreationTimestamp.Time).Round(time.Second).String()

        result = append(result, PVCInfo{
            Name:         pvc.Name,
            Namespace:    pvc.Namespace,
            Status:       pvc.Status.Phase,
            Volume:       pvc.Spec.VolumeName,
            Capacity:       requested.String(),
            ActualCapacity: actual.String(),
            AccessModes:  accessModes,
            StorageClass: storageClass,
            CreatedAt:    pvc.CreationTimestamp.Format(time.RFC3339),
            Age:          age,
            Labels:       pvcLabels,
        })
    }

    return result, nil
}

