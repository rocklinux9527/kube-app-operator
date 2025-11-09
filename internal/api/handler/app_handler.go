package handler

import (
	"github.com/k8s/kube-app-operator/internal/approval/models"
	service "github.com/k8s/kube-app-operator/internal/approval/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"strconv"
)

type AppHandler struct {
	svc *service.AppService
}

func NewAppHandler(svc *service.AppService) *AppHandler {
	return &AppHandler{svc: svc}
}

// 表格列配置
var appColumns = []map[string]string{
	{"label": "ID", "prop": "id"},
	{"label": "应用名称", "prop": "name"},
	{"label": "命名空间", "prop": "namespace"},
	{"label": "关联模板ID", "prop": "template_id"},
	{"label": "创建时间", "prop": "created_at"},
	{"label": "更新时间", "prop": "updated_at"},
	{"label": "镜像", "prop": "image"},
}

// 通用响应
func appRespond(c *gin.Context, data interface{}, columns []map[string]string, err error, emptyMsg string) {
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
	if v.Kind() == reflect.Slice {
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

// 创建应用

func (h *AppHandler) Create(c *gin.Context) {
	var req models.App
	if err := c.ShouldBindJSON(&req); err != nil {
		appRespond(c, nil, appColumns, err, "")
		return
	}
	resp, err := h.svc.CreateApp(&req)
	appRespond(c, resp, appColumns, err, "")
}

// 更新应用

func (h *AppHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req models.App
	if err := c.ShouldBindJSON(&req); err != nil {
		appRespond(c, nil, appColumns, err, "")
		return
	}
	req.ID = uint(id)

	resp, err := h.svc.UpdateApp(&req)
	appRespond(c, resp, appColumns, err, "")
}

// 获取单个应用

func (h *AppHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	item, err := h.svc.GetApp(uint(id))
	appRespond(c, item, appColumns, err, "未找到该应用")
}

// 应用列表

func (h *AppHandler) List(c *gin.Context) {
	list, err := h.svc.ListApps()
	appRespond(c, list, appColumns, err, "当前没有已创建的应用")
}

// 删除应用

func (h *AppHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := h.svc.DeleteApp(uint(id))
	appRespond(c, gin.H{"message": "deleted"}, appColumns, err, "")
}







