package utils

import (
    "context"
    "fmt"
    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/util/intstr"
    "sigs.k8s.io/controller-runtime/pkg/client"
    logf "sigs.k8s.io/controller-runtime/pkg/log"
    "strings"
    "time"
)

var log = logf.Log.WithName("logutils-creator")
// normalizePort 规范化端口号

func NormalizePort(port int32) int32 {
    const (
        minPort = 1
        maxPort = 65535
        defaultPort = 80
    )

    if port < minPort || port > maxPort {log.V(1).Info("端口号超出范围，使用默认端口", "原始端口", port, "默认端口", defaultPort)
        return defaultPort
    }

    return port
}

// mergeMaps 合并多个 map

func MergeMaps(maps ...map[string]string) map[string]string {
    result := make(map[string]string)
    for _, m := range maps {
        for k, v := range m {
            result[k] = v
        }
    }
    return result
}
// validatePort 验证端口是否在有效范围内

func ValidatePort(port int32) error {
    const (
        minPort = 1
        maxPort = 65535
    )

    if port < minPort || port > maxPort {
        return fmt.Errorf("端口号必须在 %d 到 %d 之间", minPort, maxPort)
    }

    return nil
}


// DeleteIfExists 尝试删除指定资源，如果不存在则忽略

func DeleteIfExists(ctx context.Context, cli client.Client, obj client.Object) error {
    logger := logf.FromContext(ctx)

    err := cli.Get(ctx, client.ObjectKeyFromObject(obj), obj)
    if err != nil {
       if errors.IsNotFound(err) {
            logger.V(1).Info("资源不存在，跳过删除", "Kind", obj.GetObjectKind().GroupVersionKind().Kind, "Name", obj.GetName())
            return nil
        }
        return err
    }

    logger.Info("删除资源", "Kind", obj.GetObjectKind().GroupVersionKind().Kind, "Name", obj.GetName())
    return cli.Delete(ctx, obj)
}


func intstrFromInt(i int) intstr.IntOrString {
    return intstr.FromInt(i)
}


func FormatAge(t time.Time) string {
    if t.IsZero() {
        return ""
    }
    d := time.Since(t)

    if d.Hours() < 1 {
        return fmt.Sprintf("%dm", int(d.Minutes()))
    } else if d.Hours() < 24 {
        return fmt.Sprintf("%dh", int(d.Hours()))
    } else {
        return fmt.Sprintf("%dd", int(d.Hours()/24))
    }
}



func FormatSvcAge(d time.Duration) string {
    if d.Hours() >= 24 {
        return fmt.Sprintf("%dd", int(d.Hours()/24))
    } else if d.Hours() >= 1 {
        return fmt.Sprintf("%dh", int(d.Hours()))
    } else if d.Minutes() >= 1 {
        return fmt.Sprintf("%dm", int(d.Minutes()))
    }
    return fmt.Sprintf("%ds", int(d.Seconds()))
}


func DerefString(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}

func TimeToCST(t time.Time) string {
    loc := time.FixedZone("CST", 8*3600)
    return t.In(loc).Format("2006-01-02 15:04:05")
}


func FormatLabels(labels map[string]string) string {
    if len(labels) == 0 {
        return ""
    }
    pairs := make([]string, 0, len(labels))
    for k, v := range labels {
        pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
    }
    return strings.Join(pairs, ", ")
}