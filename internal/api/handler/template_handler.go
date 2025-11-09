package handler

import (
	"encoding/json"
	"github.com/k8s/kube-app-operator/internal/approval/models"
	service "github.com/k8s/kube-app-operator/internal/approval/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"strconv"
)

// --- 通用响应方法 ---
func tempRespond(c *gin.Context, data interface{}, columns []map[string]string, err error, emptyMsg string) {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": err.Error(),
			"data":    []interface{}{},
			"total":   0,
			"columns": columns,
		})
		return
	}

	if data == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    20000,
			"message": emptyMsg,
			"data":    []interface{}{},
			"total":   0,
			"columns": columns,
		})
		return
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		if v.Len() == 0 {
			c.JSON(http.StatusOK, gin.H{
				"code":    20000,
				"message": emptyMsg,
				"data":    []interface{}{},
				"total":   0,
				"columns": columns,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    20000,
			"message": "success",
			"data":    data,
			"total":   v.Len(),
			"columns": columns,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "success",
		"data":    data,
		"total":   1,
		"columns": columns,
	})
}

// --- TemplateHandler ---

type TemplateHandler struct {
	svc *service.TemplateService
}

func NewTemplateHandler(svc *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{svc: svc}
}

var templateColumns = []map[string]string{
	{"label": "ID", "prop": "id"},
	{"label": "模版名称", "prop": "name"},
	{"label": "模版类型", "prop": "type"},
	{"label": "描述", "prop": "description"},
	{"label": "内容", "prop": "content"},
	{"label": "创建时间", "prop": "created_at"},
	{"label": "更新时间", "prop": "updated_at"},
}

// --- 创建模版 ---

func (h *TemplateHandler) Create(c *gin.Context) {
	var req models.Template
	body := make(map[string]interface{})

	if err := c.ShouldBindJSON(&body); err != nil {
		tempRespond(c, nil, templateColumns, err, "")
		return
	}

	// 基础字段
	if v, ok := body["name"].(string); ok {
		req.Name = v
	}
	if v, ok := body["type"].(string); ok {
		req.Type = v
	}
	if v, ok := body["description"].(string); ok {
		req.Description = v
	}

	// 自动处理 content 对象 → JSON
	if content, ok := body["content"]; ok {
		if b, err := json.Marshal(content); err == nil {
			req.Content = b
		} else {
			tempRespond(c, nil, templateColumns, err, "")
			return
		}
	}

	err := h.svc.CreateTemplate(&req)
	tempRespond(c, req, templateColumns, err, "")
}

// --- 查询模版列表 ---

func (h *TemplateHandler) List(c *gin.Context) {
	list, err := h.svc.ListTemplates()
	tempRespond(c, list, templateColumns, err, "当前没有模版")
}

// --- 查询单个模版 ---

func (h *TemplateHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	item, err := h.svc.GetTemplate(uint(id))
	tempRespond(c, item, templateColumns, err, "未找到该模版")
}

// --- 更新模版 ---

func (h *TemplateHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req models.Template
	body := make(map[string]interface{})

	if err := c.ShouldBindJSON(&body); err != nil {
		tempRespond(c, nil, templateColumns, err, "")
		return
	}

	req.ID = uint(id)
	if v, ok := body["name"].(string); ok {
		req.Name = v
	}
	if v, ok := body["type"].(string); ok {
		req.Type = v
	}
	if v, ok := body["description"].(string); ok {
		req.Description = v
	}
	//if content, ok := body["content"]; ok {
	//	if b, err := json.Marshal(content); err == nil {
	//		req.Content = b
	//	} else {
	//		tempRespond(c, nil, templateColumns, err, "")
	//		return
	//	}
	//}
	if content, ok := body["content"]; ok {
		switch v := content.(type) {
		case string:
			// 前端传字符串时尝试解析
			req.Content = json.RawMessage(v)
		default:
			b, err := json.Marshal(v)
			if err != nil {
				tempRespond(c, nil, templateColumns, err, "")
				return
			}
			req.Content = b
		}
	}
	err := h.svc.UpdateTemplate(&req)
	tempRespond(c, req, templateColumns, err, "")
}

// --- 删除模版 ---

func (h *TemplateHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := h.svc.DeleteTemplate(uint(id))
	tempRespond(c, gin.H{"message": "deleted"}, templateColumns, err, "")
}
