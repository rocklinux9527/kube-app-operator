package handler

import (
	"fmt"
	"net/http"
	"reflect"

	clustom "github.com/k8s/kube-app-operator/internal/custom"
	"github.com/gin-gonic/gin"
)

// 通用响应方法（通用化：支持任意切片/数组或单个对象）

func respond(c *gin.Context, data interface{}, columns []map[string]string, err error, emptyMsg string) {
	// 错误优先
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

	// data == nil 当作空数据返回
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
	// 如果是切片或数组，检查长度并返回
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

	// 单个对象 -> total = 1
	c.JSON(http.StatusOK, gin.H{
		"code":    20000,
		"message": "success",
		"data":    data,
		"total":   1,
		"columns": columns,
	})
}

// ListNamespacesHandler 查询全部命名空间

func ListNamespaces(c *gin.Context) {
	columns := []map[string]string{
		{"label": "命名空间", "prop": "name"},
		{"label": "状态", "prop": "status"},
		{"label": "阶段", "prop": "phase"},
		{"label": "标签", "prop": "labels"},
		{"label": "创建时间", "prop": "created_at"},
		{"label": "AGE", "prop": "age"},
	}

	namespaces, err := clustom.ListAllNamespaces()

	respond(c, namespaces, columns, err, "当前集群中没有 Namespace")
}
// GetKubeDeployments

func GetKubeDeployments(c *gin.Context) {
	columns := []map[string]string{
		{"label": "服务名称", "prop": "app_name"},
		{"label": "命名空间", "prop": "namespace"},
		{"label": "副本", "prop": "replicas"},
		{"label": "镜像", "prop": "image"},
		{"label": "环境变量", "prop": "deploy_env"},
		{"label": "应用端口", "prop": "ports"},
		{"label": "创建时间", "prop": "created_at"},
		{"label": "更新时间", "prop": "updated_at"},
		{"label": "AGE", "prop": "age"},
		{"label": "READY", "prop": "ready"},
		{"label": "UP-TO-DATE", "prop": "up_to_date"},
		{"label": "AVAILABLE", "prop": "available"},
	}
	ns := c.DefaultQuery("namespace", "default")
	deployments, err := clustom.ListDeployments(ns)
	respond(c, deployments, columns, err, "该空间下没有 Deployments")
}

// GetKubeServices

func GetKubeServices(c *gin.Context) {
	columns := []map[string]string{
		{"label": "服务名称", "prop": "name"},
		{"label": "命名空间", "prop": "namespace"},
		{"label": "ClusterIP", "prop": "cluster_ip"},
		{"label": "类型", "prop": "type"},
		{"label": "端口", "prop": "ports"},
		{"label": "Selector", "prop": "selector"},
		{"label": "创建时间", "prop": "created_at"},
		{"label": "AGE", "prop": "age"},
	}
	ns := c.DefaultQuery("namespace", "default")
	services, err := clustom.ListServices(ns)
	respond(c, services, columns, err, "该命名空间下没有 Service")
}

// GetKubeIngress

func GetKubeIngress(c *gin.Context) {
	columns := []map[string]string{
		{"label": "Ingress 名称", "prop": "name"},
		{"label": "命名空间", "prop": "namespace"},
		// {"label": "路由关系 (Host → Path → Service)", "prop": "routes"},
		{"label": "IngressClass", "prop": "ingress_class_name"},
		{"label": "AGE", "prop": "age"},
		{"label": "创建时间", "prop": "created_at"},
	}

	ns := c.DefaultQuery("namespace", "")
	ingresses, err := clustom.ListIngress(ns)
	respond(c, ingresses, columns, err, "该命名空间下没有 Ingress")
}


func GetKubePVCS(c *gin.Context) {
	columns := []map[string]string{
		{"label": "名称", "prop": "name"},
		{"label": "命名空间", "prop": "namespace"},
		{"label": "状态", "prop": "status"},
		//{"label": "卷", "prop": "volume"},
		{"label": "申请值", "prop": "capacity"},
		{"label": "实际大小", "prop": "actual_capacity"},
		{"label": "访问模式", "prop": "access_modes"},
		{"label": "动态卷", "prop": "storage_class"},
		{"label": "创建时间", "prop": "created_at"},
		{"label": "AGE", "prop": "age"},
		{"label": "标签", "prop": "labels"},
	}

	ns := c.DefaultQuery("namespace", "default")
	pvcs, err := clustom.ListPVCs(ns)
	respond(c, pvcs, columns, err, "该命名空间下没有 PVC")
}

// GetKubePods

func GetKubePods(c *gin.Context) {
	columns := []map[string]string{
		{"label": "命名空间", "prop": "namespace"},
		{"label": "Pod 名称", "prop": "pod_name"},
		{"label": "主机 IP", "prop": "host_ip"},
		{"label": "Pod IP", "prop": "pod_ip"},
		{"label": "端口", "prop": "ports"},
		{"label": "QoS", "prop": "qos"},
		{"label": "状态", "prop": "pod_status"},
		{"label": "Ready", "prop": "ready_count"},
		{"label": "重启次数", "prop": "restarts"},
		{"label": "错误原因", "prop": "pod_error_reasons"},
		{"label": "启动时间", "prop": "start_time"},
		{"label": "AGE", "prop": "age"},
		{"label": "镜像", "prop": "image"},
		{"label": "容器名称", "prop": "container_names"},
		{"label": "Init 容器名称", "prop": "init_names"},
	}
	ns := c.DefaultQuery("namespace", "default")
	pods, err := clustom.ListPods(ns)
	respond(c, pods, columns, err, "该命名空间下没有 Pod")
}

// RestartKubePodHandler 重启指定的 Pod（通过删除 Pod，让控制器重新创建）

func RestartKubePod(c *gin.Context) {
	var req struct {
		Namespace string `json:"namespace" binding:"required"`
		PodName   string `json:"pod_name" binding:"required"`
	}

	columns := []map[string]string{
		{"label": "命名空间", "prop": "namespace"},
		{"label": "Pod 名称", "prop": "pod_name"},
		{"label": "主机 IP", "prop": "host_ip"},
		{"label": "Pod IP", "prop": "pod_ip"},
		{"label": "端口", "prop": "ports"},
		{"label": "QoS", "prop": "qos"},
		{"label": "状态", "prop": "pod_status"},
		{"label": "Ready", "prop": "ready_count"},
		{"label": "重启次数", "prop": "restarts"},
		{"label": "错误原因", "prop": "pod_error_reasons"},
		{"label": "启动时间", "prop": "start_time"},
		{"label": "AGE", "prop": "age"},
		{"label": "镜像", "prop": "image"},
		{"label": "容器名称", "prop": "container_names"},
		{"label": "Init 容器名称", "prop": "init_names"},
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, []interface{}{}, columns, fmt.Errorf("参数错误: %v", err), "参数错误")
		return
	}

	// 调用自定义逻辑重启 Pod（实现为删除 Pod）
	if err := clustom.RestartPod(req.Namespace, req.PodName); err != nil {
		respond(c, []interface{}{}, columns, fmt.Errorf("重启 Pod 失败: %v", err), "重启 Pod 失败")
		return
	}

	// 返回被操作的 pod 基本信息（用 slice 以保持与其他接口一致）
	data := []map[string]string{
		{
			"namespace": req.Namespace,
			"pod_name":  req.PodName,
		},
	}
	respond(c, data, columns, nil, "")
}


func RolloutRestart(c *gin.Context) {
	var req struct {
		Kind      string `json:"kind" binding:"required"`      // Deployment / DaemonSet / StatefulSet
		Namespace string `json:"namespace" binding:"required"` // 命名空间
		Name      string `json:"name" binding:"required"`      // 资源名
	}

	columns := []map[string]string{
		{"label": "类型", "prop": "kind"},
		{"label": "命名空间", "prop": "namespace"},
		{"label": "名称", "prop": "name"},
		{"label": "结果", "prop": "result"},
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, []interface{}{}, columns, fmt.Errorf("参数错误: "+err.Error()), "")
		return
	}

	if err := clustom.RolloutRestart(req.Kind, req.Namespace, req.Name); err != nil {
		respond(c, []interface{}{}, columns, fmt.Errorf("重启失败: "+err.Error()), "")
		return
	}

	data := []map[string]string{
		{
			"kind":      req.Kind,
			"namespace": req.Namespace,
			"name":      req.Name,
			"result":    "已触发滚动重启",
		},
	}

	respond(c, data, columns, nil,
		fmt.Sprintf("%s/%s/%s 已触发滚动重启", req.Kind, req.Namespace, req.Name))
}

