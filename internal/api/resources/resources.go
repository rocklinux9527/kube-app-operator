package resources


import (
    "context"
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    networkingv1 "k8s.io/api/networking/v1"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    commontype "github.com/k8s/kube-app-operator/internal/api/types"
    kubev1alpha1 "github.com/k8s/kube-app-operator/api/v1alpha1"
    "sigs.k8s.io/controller-runtime/pkg/client"
)



// DeleteKubeAppResources 封装资源删除逻辑

func DeleteKubeAppResources(ctx context.Context, k8sClient client.Client, scheme *runtime.Scheme, req commontype.KubeDeleteAppRequest) commontype.DeleteResult {
    key := types.NamespacedName{Name: req.Name, Namespace: req.Namespace}
    deleted := []string{}
    notFound := []string{}

    // Deployment
    if req.DeleteDeployment {
        var dep appsv1.Deployment
        if err := k8sClient.Get(ctx, key, &dep); err == nil {
            if err := k8sClient.Delete(ctx, &dep); err != nil {
                return commontype.DeleteResult{Err: err, ErrMsg: "删除 Deployment 失败", Code: 50001}
            }
            deleted = append(deleted, "Deployment")
        } else if apierrors.IsNotFound(err) {
            notFound = append(notFound, "Deployment")
        } else {
            return commontype.DeleteResult{Err: err, ErrMsg: "查询 Deployment 失败", Code: 50002}
        }
    }

    // Service
    if req.DeleteService {
        var svc corev1.Service
        if err := k8sClient.Get(ctx, key, &svc); err == nil {
            if err := k8sClient.Delete(ctx, &svc); err != nil {
                return commontype.DeleteResult{Err: err, ErrMsg: "删除 Service 失败", Code: 50003}
            }
            deleted = append(deleted, "Service")
        } else if apierrors.IsNotFound(err) {
            notFound = append(notFound, "Service")
        } else {
            return commontype.DeleteResult{Err: err, ErrMsg: "查询 Service 失败", Code: 50004}
        }
    }

    // Ingress
    if req.DeleteIngress {
        var ing networkingv1.Ingress
        if err := k8sClient.Get(ctx, key, &ing); err == nil {
            if err := k8sClient.Delete(ctx, &ing); err != nil {
                return commontype.DeleteResult{Err: err, ErrMsg: "删除 Ingress 失败", Code: 50005}
            }
            deleted = append(deleted, "Ingress")
        } else if apierrors.IsNotFound(err) {
            notFound = append(notFound, "Ingress")
        } else {
            return commontype.DeleteResult{Err: err, ErrMsg: "查询 Ingress 失败", Code: 50006}
        }
    }

    // PVC
    if req.DeletePVC {
        var pvc corev1.PersistentVolumeClaim
        if err := k8sClient.Get(ctx, key, &pvc); err == nil {
            if err := k8sClient.Delete(ctx, &pvc); err != nil {
                return commontype.DeleteResult{Err: err, ErrMsg: "删除 PVC 失败", Code: 50007}
            }
            deleted = append(deleted, "PVC")
        } else if apierrors.IsNotFound(err) {
            notFound = append(notFound, "PVC")
        } else {
            return commontype.DeleteResult{Err: err, ErrMsg: "查询 PVC 失败", Code: 50008}
        }
    }

    // KubeApp
    if req.DeleteKubeApp {
        var app kubev1alpha1.KubeApp
        if err := k8sClient.Get(ctx, key, &app); err == nil {
            if err := k8sClient.Delete(ctx, &app); err != nil {
                return commontype.DeleteResult{Err: err, ErrMsg: "删除 KubeApp 失败", Code: 50009}
            }
            deleted = append(deleted, "KubeApp")
        } else if apierrors.IsNotFound(err) {
            notFound = append(notFound, "KubeApp")
        } else {
            return commontype.DeleteResult{Err: err, ErrMsg: "查询 KubeApp 失败", Code: 50010}
        }
    }

    return commontype.DeleteResult{
        Deleted:  deleted,
        NotFound: notFound,
    }
}

