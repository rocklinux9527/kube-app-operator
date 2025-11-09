package types



type KubeAppRequest struct {
    Name         string `json:"name"`
    Namespace    string `json:"namespace"`
    TemplateType string `json:"templateType"`
    Image        string `json:"image"`
    Replicas     int32 `json:"replicas"`
    TemplateName   string `json:"TemplateName"`
}

type KubeDeleteAppRequest struct {
    Name      string `json:"name" binding:"required"`
    Namespace string `json:"namespace" binding:"required"`

    DeleteDeployment bool `json:"deleteDeployment,omitempty"`
    DeleteService    bool `json:"deleteService,omitempty"`
    DeleteIngress    bool `json:"deleteIngress,omitempty"`
    DeletePVC        bool `json:"deletePvc,omitempty"`
    DeleteKubeApp    bool `json:"deleteKubeApp,omitempty"`
}

type ErrorResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Detail  string `json:"detail,omitempty"`
}


// DeleteResult 封装删除操作的结果
type DeleteResult struct {
    Deleted  []string
    NotFound []string
    Err      error
    ErrMsg   string
    Code     int
}

