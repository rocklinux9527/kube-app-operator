package define

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

// RolloutRestart 支持 Deployment / DaemonSet / StatefulSet 的滚动重启

func RolloutRestart(kind, namespace, name string) error {
	if GlobalClient == nil {
		return fmt.Errorf("k8s client 未初始化，请先调用 Init()")
	}

	ctx := context.Background()

	// 使用北京时间 (+8)
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc).Format("2006-01-02 15:04:05")

	// 统一转小写，兼容各种写法
	k := strings.ToLower(kind)

	switch k {
	case "deployment", "deploy":
		var deploy appsv1.Deployment
		if err := GlobalClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &deploy); err != nil {
			return fmt.Errorf("获取 Deployment 失败: %v", err)
		}
		updateRestartAnnotation(&deploy.Spec.Template.Annotations, now)
		if err := GlobalClient.Update(ctx, &deploy); err != nil {
			return fmt.Errorf("更新 Deployment 失败: %v", err)
		}

	case "daemonset", "ds":
		var ds appsv1.DaemonSet
		if err := GlobalClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &ds); err != nil {
			return fmt.Errorf("获取 DaemonSet 失败: %v", err)
		}
		updateRestartAnnotation(&ds.Spec.Template.Annotations, now)
		if err := GlobalClient.Update(ctx, &ds); err != nil {
			return fmt.Errorf("更新 DaemonSet 失败: %v", err)
		}

	case "statefulset", "sts":
		var sts appsv1.StatefulSet
		if err := GlobalClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &sts); err != nil {
			return fmt.Errorf("获取 StatefulSet 失败: %v", err)
		}
		updateRestartAnnotation(&sts.Spec.Template.Annotations, now)
		if err := GlobalClient.Update(ctx, &sts); err != nil {
			return fmt.Errorf("更新 StatefulSet 失败: %v", err)
		}

	default:
		return fmt.Errorf("不支持的资源类型: %s (仅支持 Deployment / DaemonSet / StatefulSet)", kind)
	}

	return nil
}

// updateRestartAnnotation 设置/更新重启时间

func updateRestartAnnotation(annotations *map[string]string, now string) {
	if *annotations == nil {
		*annotations = map[string]string{}
	}
	(*annotations)["kubectl.kubernetes.io/restartedAt"] = now
}