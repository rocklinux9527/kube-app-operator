package define

import (
	"context"
	"fmt"
	"github.com/k8s/kube-app-operator/internal/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"time"
)

type PodContainer struct {
	Name         string   `json:"name"`
	Image        string   `json:"image"`
	Ports        []string `json:"ports"`
	Ready        bool     `json:"ready"`
	State        string   `json:"state"`
	RestartCount int32    `json:"restart_count"`
}

type PodInfo struct {
	Namespace      string        `json:"namespace"`
	PodName        string        `json:"pod_name"`
	HostIP         string        `json:"host_ip"`
	PodIP          string        `json:"pod_ip"`
	Ports          []string      `json:"ports"`
	QoS            string        `json:"qos"`
	PodStatus      string        `json:"pod_status"`
	ReadyCount     string        `json:"ready_count"`
	Restarts       int32         `json:"restarts"`
	PodErrorReason []string      `json:"pod_error_reasons"`
	StartTime      string        `json:"start_time"`
	Age            string        `json:"age"`
	Image          []string      `json:"image"`
	Containers     []PodContainer `json:"containers"`
	InitContainers []PodContainer `json:"init_containers"`
	ContainerNames []string       `json:"container_names"`
	InitNames      []string       `json:"init_names"`
}

// ListPods 查询指定命名空间下的 Pod
func ListPods(namespace string) ([]PodInfo, error) {
	if GlobalClient == nil {
		return nil, fmt.Errorf("k8s client 未初始化，请先调用 custom.Init()")
	}

	var podList corev1.PodList
	if err := GlobalClient.List(context.Background(), &podList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	loc := time.FixedZone("CST", 8*3600)
	var result []PodInfo

	for _, pod := range podList.Items {
		startTime := pod.Status.StartTime
		created := time.Now()
		if startTime != nil {
			created = startTime.Time
		}

		cstStart := created.In(loc)
		age := utils.FormatSvcAge(time.Since(created))

		qos := string(pod.Status.QOSClass)
		if qos == "BestEffort" {
			qos = "低"
		} else if qos == "Burstable" {
			qos = "中"
		} else if qos == "Guaranteed" {
			qos = "高"
		}

		// 容器
		var containers []PodContainer
		var containerNames []string
		var images []string
		var ports []string
		var readyCount int
		var restarts int32
		var errorReasons []string

		for _, c := range pod.Spec.Containers {
			cName := c.Name
			containerNames = append(containerNames, cName)
			img := c.Image
			images = append(images, img)

			var containerPorts []string
			for _, p := range c.Ports {
				containerPorts = append(containerPorts, strconv.Itoa(int(p.ContainerPort)))
				ports = append(ports, strconv.Itoa(int(p.ContainerPort)))
			}

			ready := false
			state := "Unknown"
			restartCount := int32(0)

			for _, cs := range pod.Status.ContainerStatuses {
				if cs.Name != c.Name {
					continue
				}
				restartCount = cs.RestartCount
				restarts += restartCount
				if cs.Ready {
					ready = true
					readyCount++
				}
				if cs.State.Waiting != nil {
					state = cs.State.Waiting.Reason
					if cs.State.Waiting.Message != "" {
						errorReasons = append(errorReasons, cs.State.Waiting.Message)
					}
				} else if cs.State.Terminated != nil {
					state = cs.State.Terminated.Reason
					if cs.State.Terminated.Message != "" {
						errorReasons = append(errorReasons, cs.State.Terminated.Message)
					}
				} else if cs.State.Running != nil {
					state = "Running"
				}
			}

			containers = append(containers, PodContainer{
				Name:         cName,
				Image:        img,
				Ports:        containerPorts,
				Ready:        ready,
				State:        state,
				RestartCount: restartCount,
			})
		}

		// InitContainers
		var initContainers []PodContainer
		var initNames []string
		for _, c := range pod.Spec.InitContainers {
			initNames = append(initNames, c.Name)
			initContainers = append(initContainers, PodContainer{
				Name:  c.Name,
				Image: c.Image,
			})
		}

		podInfo := PodInfo{
			Namespace:      pod.Namespace,
			PodName:        pod.Name,
			HostIP:         pod.Status.HostIP,
			PodIP:          pod.Status.PodIP,
			Ports:          ports,
			QoS:            qos,
			PodStatus:       string(pod.Status.Phase),
			ReadyCount:     fmt.Sprintf("%d/%d", readyCount, len(pod.Spec.Containers)),
			Restarts:       restarts,
			PodErrorReason: errorReasons,
			StartTime:      utils.TimeToCST(cstStart),
			Age:            age,
			Image:          images,
			Containers:     containers,
			InitContainers: initContainers,
			ContainerNames: containerNames,
			InitNames:      initNames,
		}

		result = append(result, podInfo)
	}

	return result, nil
}

// RestartPod 删除指定的 Pod，K8S 控制器会自动拉起新的 Pod，相当于重启
func RestartPod(namespace, podName string) error {
	if GlobalClient == nil {
		return fmt.Errorf("k8s client 未初始化，请先调用 custom.Init()")
	}

	var pod corev1.Pod
	if err := GlobalClient.Get(context.Background(), client.ObjectKey{
		Namespace: namespace,
		Name:      podName,
	}, &pod); err != nil {
		return fmt.Errorf("获取 Pod 失败: %v", err)
	}

	// 删除 Pod，Deployment/ReplicaSet 会自动重建

	if err := GlobalClient.Delete(context.Background(), &pod); err != nil {
		return fmt.Errorf("删除 Pod 失败: %v", err)
	}

	return nil
}