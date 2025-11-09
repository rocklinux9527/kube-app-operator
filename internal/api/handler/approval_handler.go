package handler

import (
	"github.com/k8s/kube-app-operator/internal/approval/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ApprovalResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RequestHandler 处理请求相关的 HTTP 接口
type RequestHandler struct {
	svc *services.RequestService
}

// NewRequestHandler 构造函数
func NewRequestHandler(svc *services.RequestService) *RequestHandler {
	return &RequestHandler{svc: svc}
}

// -------------------- 创建请求 --------------------

func (h *RequestHandler) CreateRequest(c *gin.Context) {
	var body services.CreateRequestInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ApprovalResponse{
			Code:    40000,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	req, err := h.svc.CreateRequest(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ApprovalResponse{
			Code:    50000,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ApprovalResponse{
		Code:    20000,
		Message: "success",
		Data:    req,
	})
}


// -------------------- 创建请求 --------------------

func (h *RequestHandler) DeleteRequestList(c *gin.Context) {
	var body services.DeleteRequestInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ApprovalResponse{
			Code:    40000,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.svc.DeleteRequest(body); err != nil {
		c.JSON(http.StatusInternalServerError, ApprovalResponse{
			Code:    50000,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ApprovalResponse{
		Code:    20000,
		Message: "delete success",
	})
}

// -------------------- 审批请求（通过/拒绝） --------------------
func (h *RequestHandler) processApproval(c *gin.Context, decision string) {
	id := c.Param("id")
	var body services.ApprovalInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 20000,
			"message": "invalid request: " + err.Error(),
			"data": nil,
		})
		return
}
body.Decision = decision // 强制设置审批结果

req, err := h.svc.ApproveRequest(id, body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 20000,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// ⚡ 在这里附加 steps
	steps := []map[string]string{
		{"role": "ops", "description": "运维审批"},
		{"role": "sre", "description": "SRE 审批"},
		{"role": "k8s", "description": "K8S 审批"},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "success",
		"data": gin.H{
			"request": req,
			"steps":   steps,
		},
	})
}

func (h *RequestHandler) ApproveRequest(c *gin.Context) {
	h.processApproval(c, "APPROVE")
}

func (h *RequestHandler) RejectRequest(c *gin.Context) {
	h.processApproval(c, "REJECT")
}

// -------------------- 分页查询 --------------------

// GET /approvals/list?page=1&page_size=10

func (h *RequestHandler) ListRequests(c *gin.Context) {
	page, err1 := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, err2 := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if err1 != nil || page <= 0 {
		page = 1
	}
	if err2 != nil || pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 调用服务获取数据
	requests, total, err := h.svc.ListRequests(page, pageSize)

	// 定义 columns（表格列配置）——平级返回
	columns := []map[string]string{
		{"prop": "request_id", "label": "请求ID"},
		{"prop": "created_by", "label": "申请人"},
		{"prop": "business_line", "label": "业务线"},
		{"prop": "service_name", "label": "服务名"},
		{"prop": "image", "label": "镜像"},
		{"prop": "replicas", "label": "副本数"},
		{"prop": "template_name", "label": "模板名"},
		{"prop": "purpose", "label": "用途"},
		{"prop": "status", "label": "状态"},
		{"prop": "operation", "label": "操作类型"},
		{"prop": "last_updated", "label": "最后更新时间"},
		{"prop": "created_at", "label": "创建时间"},
		{"prop": "updated_at", "label": "更新时间"},
	}

	if err != nil {
		// 错误情况下也返回统一结构：data 作为空数组，columns 仍然返回，前端更健壮
		c.JSON(http.StatusOK, gin.H{
			"code":     50000,
			"message":  err.Error(),
			"data":     []interface{}{}, // data 直接是数组（不使用 items）
			"page":     page,
			"pageSize": pageSize,
			"total":    0,
			"columns":  columns,
		})
		return
	}

	// 成功返回：data 是记录数组（直接放到 data），分页信息平级返回
	c.JSON(http.StatusOK, gin.H{
		"code":     20000,
		"message":  "success",
		"data":     requests,
		"page":     page,
		"pageSize": pageSize,
		"total":    total,
		"columns":  columns,
	})
}


func (h *RequestHandler) BatchFindByIDs(c *gin.Context) {
	var body struct {
		RequestIDs []string `json:"request_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "invalid request: " + err.Error(),
		})
		return
	}
	if len(body.RequestIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "request_ids cannot be empty",
		})
		return
	}

	requests, err := h.svc.BatchFindByIDs(body.RequestIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "success",
		"data": gin.H{
			"items": requests,
		},
	})
}

func (h *RequestHandler) FindByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "id cannot be empty",
		})
		return
	}

	req, err := h.svc.FindByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": err.Error(),
		})
		return
	}

	if req == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    40400,
			"message": "request not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "success",
		"data": req,
		"steps": []gin.H{
			{"role": "OPS", "description": "运维审批"},
			{"role": "SRE", "description": "SRE 审批"},
			{"role": "K8S", "description": "K8S 审批"},
		},
	})
}




