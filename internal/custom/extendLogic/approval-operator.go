package extendLogic

import (
    "context"
    "fmt"
    "github.com/k8s/kube-app-operator/internal/api/templates"
    repo "github.com/k8s/kube-app-operator/internal/approval/repositories"
    "gorm.io/gorm"
    "sync"

    kubev1alpha1 "github.com/k8s/kube-app-operator/api/v1alpha1"
    k8sresources "github.com/k8s/kube-app-operator/internal/api/resources"
    commontype "github.com/k8s/kube-app-operator/internal/api/types"
    "k8s.io/apimachinery/pkg/runtime"
    "reflect"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
    globalClient client.Client
    globalScheme *runtime.Scheme
    db *gorm.DB
    once         sync.Once
)

// Init 用于初始化全局 client 和 scheme
func Init(k8sClient client.Client, scheme *runtime.Scheme, database *gorm.DB) {
    globalClient = k8sClient
    globalScheme = scheme
    db = database
}


// InternalCreateKubeApp 供内部审批流调用
func InternalCreateKubeApp(req commontype.KubeAppRequest ) error {
    if globalClient == nil || globalScheme == nil || db == nil {
        return fmt.Errorf("k8s client 未初始化，和 数据库连接未初始化 请先调用 extendLogic.Init()")
    }

    // 初始化 TemplateService（注入数据库连接）
    tmplRepo := repo.NewTemplateRepo(db)

    var KubeApp *kubev1alpha1.KubeApp
    switch req.TemplateType {
    case "backend":
        KubeApp = templates.BuildOperatorAppFromDB(tmplRepo, req.TemplateName, req.Name, req.Namespace, req.Image, req.Replicas)
        if KubeApp == nil || reflect.DeepEqual(*KubeApp, kubev1alpha1.KubeApp{}) {
            return fmt.Errorf("模板生成 KubeApp 为空，请检查模板内容或模板名是否正确")
        }
    case "frontend":
        KubeApp = templates.BuildOperatorAppFromDB(tmplRepo, req.TemplateName, req.Name, req.Namespace, req.Image, req.Replicas)
    default:
        return fmt.Errorf("不支持的模板类型: %s", req.TemplateType)
    }
    if err := globalClient.Create(context.Background(), KubeApp); err != nil {
        return fmt.Errorf("创建 KubeApp 失败: %w", err)
    }
   return nil
}

// InternalUpdateKubeApp 供内部审批流调用 (只更新 image + replicas)

func InternalUpdateKubeApp(req commontype.KubeAppRequest) error {
    if globalClient == nil || globalScheme == nil {
        return fmt.Errorf("k8s client 未初始化，请先调用 extendLogic.Init()")
    }

    ctx := context.Background()
    var KubeApp kubev1alpha1.KubeApp

    // 先获取现有的 KubeApp
    if err := globalClient.Get(ctx,
        client.ObjectKey{Name: req.Name, Namespace: req.Namespace},
        &KubeApp,
    ); err != nil {
        return fmt.Errorf("获取 KubeApp 失败: %w", err)
    }

    // 构造 patch 前快照
    patch := client.MergeFrom(KubeApp.DeepCopy())

    // 只更新 Deployment 下的镜像和副本数
    if KubeApp.Spec.Deployment != nil {
        KubeApp.Spec.Deployment.Image = req.Image
        KubeApp.Spec.Deployment.Replicas = &req.Replicas
    } else {
        return fmt.Errorf("KubeApp %s/%s 没有 Deployment 配置，无法更新", req.Namespace, req.Name)
    }
    // Patch 更新
    if err := globalClient.Patch(ctx, &KubeApp, patch); err != nil {
        return fmt.Errorf("更新 KubeApp 失败: %w", err)
    }

    fmt.Printf("🔄 KubeApp %s/%s 已更新 Deployment (镜像=%s, 副本数=%d)\n",
        req.Namespace, req.Name, req.Image, req.Replicas)
    return nil
}



// InternalDeleteKubeApp 供内部审批流调用
func InternalDeleteKubeApp(req commontype.KubeDeleteAppRequest) error {
    if globalClient == nil || globalScheme == nil {
        return fmt.Errorf("k8s client 未初始化，请先调用 extendLogic.Init()")
    }

    result := k8sresources.DeleteKubeAppResources(context.Background(), globalClient, globalScheme, req)

    if result.Err != nil {
        return fmt.Errorf("删除 KubeApp 失败: %w", result.Err)
    }

    if len(result.Deleted) == 0 {
        return fmt.Errorf("没有找到需要删除的资源: %v", result.NotFound)
    }
    return nil
}






