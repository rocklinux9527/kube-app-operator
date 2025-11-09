package define

import (
	"context"
	"fmt"
	"github.com/k8s/kube-app-operator/internal/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// NamespaceInfo 定义返回的字段结构

type NamespaceInfo struct {
	Name       string            `json:"name"`
	Status     string            `json:"status"`
	Labels     string `json:"labels"`
	CreatedAt  string            `json:"created_at"`
	Age        string            `json:"age"`
	Phase      string            `json:"phase"`
}

// ListAllNamespaces 查询集群中所有命名空间

func ListAllNamespaces() ([]NamespaceInfo, error) {
	if GlobalClient == nil {
		return nil, fmt.Errorf("k8s client 未初始化，请先调用 custom.Init()")
	}

	var nsList corev1.NamespaceList
	ctx := context.Background()

	if err := GlobalClient.List(ctx, &nsList, &client.ListOptions{}); err != nil {
		return nil, fmt.Errorf("获取命名空间列表失败: %v", err)
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	var result []NamespaceInfo

	for _, ns := range nsList.Items {
		createdAt := ns.CreationTimestamp.In(loc).Format("2006-01-02 15:04:05")
		age := formatAge(ns.CreationTimestamp.Time)
		labelStr := utils.FormatLabels(ns.Labels)

		result = append(result, NamespaceInfo{
			Name:      ns.Name,
			Status:    string(ns.Status.Phase),
			Labels:    labelStr,
			CreatedAt: createdAt,
			Age:       age,
			Phase:     string(ns.Status.Phase),
		})
	}

	return result, nil
}


func formatAge(t time.Time) string {
	duration := time.Since(t)

	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

