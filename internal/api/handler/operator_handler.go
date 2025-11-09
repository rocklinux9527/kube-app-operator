package handler

import (
    "fmt"
    kubev1alpha1 "github.com/k8s/kube-app-operator/api/v1alpha1"
    k8sresources "github.com/k8s/kube-app-operator/internal/api/resources"
    "github.com/k8s/kube-app-operator/internal/api/templates"
    commontype "github.com/k8s/kube-app-operator/internal/api/types"
    "github.com/gin-gonic/gin"
    "k8s.io/apimachinery/pkg/runtime"
    "net/http"
    "sigs.k8s.io/controller-runtime/pkg/client"
)



// NewCreateKubeAppHandler returns a gin.HandlerFunc with injected client and scheme

func NewCreateKubeAppHandler(k8sClient client.Client, scheme *runtime.Scheme) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req commontype.KubeAppRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest,commontype.ErrorResponse{
               Code:    40001,
	       Message: "请求参数格式错误",
	       Detail:  err.Error(),
             })
            return
        }

        var KubeApp *kubev1alpha1.KubeApp
        switch req.TemplateType {
        case "backend":
            KubeApp = templates.BackendTemplate(req.Name, req.Namespace,req.Image,req.Replicas)
        case "frontend":
              KubeApp = templates.FrontendTemplate(req.Name, req.Namespace,req.Image,req.Replicas)
        default:
            c.JSON(http.StatusBadRequest, commontype.ErrorResponse{
		Code:    40002,
		Message: "不支持的模板类型",
		Detail:  "templateType 必须是 'backend' 或 'frontend'",
	    })
            return
        }

        if err := k8sClient.Create(c.Request.Context(), KubeApp); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": fmt.Sprintf("failed to create KubeApp: %v", err),
            })
            return
        }

        c.JSON(http.StatusCreated, gin.H{"message": "KubeApp created successfully"})
    }
}

// delete resource deployment service ingress pvc app

func NewDeleteKubeAppHandler(k8sClient client.Client, scheme *runtime.Scheme) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req commontype.KubeDeleteAppRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, commontype.ErrorResponse{
                Code:    40001,
                Message: "请求参数格式错误",
                Detail:  err.Error(),
            })
            return
        }

        // 调用资源删除逻辑
        result := k8sresources.DeleteKubeAppResources(c.Request.Context(), k8sClient, scheme, req)

        if result.Err != nil {
            c.JSON(http.StatusInternalServerError, commontype.ErrorResponse{
                Code:    result.Code,
                Message: result.ErrMsg,
                Detail:  result.Err.Error(),
            })
            return
        }

        if len(result.Deleted) == 0 {
            c.JSON(http.StatusNotFound, gin.H{
                "message":  "没有找到任何资源可以删除",
                "notFound": result.NotFound,
            })
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message":         "资源删除完成",
            "deleted":         result.Deleted,
            "notFoundSkipped": result.NotFound,
        })
    }
}



