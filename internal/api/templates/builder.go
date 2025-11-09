package templates

import (
	"encoding/json"
	"fmt"
	kubev1alpha1 "github.com/k8s/kube-app-operator/api/v1alpha1"
	repo "github.com/k8s/kube-app-operator/internal/approval/repositories"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)



// BuildOperatorAppFromDB 根据数据库模板生成 KubeApp

func BuildOperatorAppFromDB(repo *repo.TemplateRepo, templateName, name, namespace, image string, replicas int32) *kubev1alpha1.KubeApp {
	fmt.Println("⚙️ 调试信息: 进入 BuildOperatorAppFromDB")
	fmt.Printf("➡️ 参数: templateName=%s, name=%s, namespace=%s, image=%s, replicas=%d\n",
		templateName, name, namespace, image, replicas)

	// Step 1: 查询模板
	tmpl, err := repo.GetTemplateByName(templateName)
	if err != nil {
		fmt.Printf("❌ 查询模板失败: %v\n", err)
		return &kubev1alpha1.KubeApp{}
	}
	if tmpl == nil {
		fmt.Println("❌ tmpl 为 nil")
		return &kubev1alpha1.KubeApp{}
	}

	fmt.Printf("✅ 模板查询成功: 名称=%s, Content类型=%T, 长度=%d\n", tmpl.Name, tmpl.Content, len(tmpl.Content))

	// Step 2: 模板内容检查
	if len(tmpl.Content) == 0 {
		fmt.Println("❌ 模板内容为空 (tmpl.Content 长度为 0)")
		return &kubev1alpha1.KubeApp{}
	}

	// 打印模板前200字符预览
	fmt.Println("🧩 模板原始内容预览前200字符:")
	if len(tmpl.Content) > 200 {
		fmt.Println(string(tmpl.Content[:200]))
	} else {
		fmt.Println(string(tmpl.Content))
	}

	// Step 3: 解析 JSON
	config := make(map[string]interface{})
	if err := json.Unmarshal(tmpl.Content, &config); err != nil {
		fmt.Printf("❌ JSON解析失败: %v\n", err)
		if len(tmpl.Content) > 3588 {
			fmt.Println("❌ 原始内容前200字符:", string(tmpl.Content[:3588]))
		} else {
			fmt.Println("❌ 原始内容:", string(tmpl.Content))
		}
		return &kubev1alpha1.KubeApp{}
	}

	fmt.Printf("✅ JSON解析成功, 顶层字段数: %d\n", len(config))
	for k := range config {
		fmt.Println(" - 含字段:", k)
	}

	// Step 4: 构建 KubeApp

	app := BuildAppFromConfig(config, name, namespace, image, replicas)
	if app == nil {
		fmt.Printf("❌ 从配置构建 KubeApp 失败, 模板名: %s\n", templateName)
		return &kubev1alpha1.KubeApp{}
	}

	fmt.Printf("✅ 成功构建 KubeApp: %s/%s\n", app.Namespace, app.Name)
	fmt.Printf("🔍 构建摘要: enableDeployment=%v, enableService=%v, enableIngress=%v, enablePvc=%v\n",
		app.Spec.EnableDeployment, app.Spec.EnableService, app.Spec.EnableIngress, app.Spec.EnablePvc)
	if app.Spec.Deployment != nil && app.Spec.Deployment.Replicas != nil {
		fmt.Printf("🔧 Deployment 镜像=%s, 副本=%d\n", app.Spec.Deployment.Image, *app.Spec.Deployment.Replicas)
	}
	return app
}

func BuildAppFromConfig(config map[string]interface{}, name, namespace, image string, replicas int32) *kubev1alpha1.KubeApp {
	deploymentConfig, _ := config["deployment"].(map[string]interface{})
	serviceConfig, _ := config["service"].(map[string]interface{})
	ingressConfig, _ := config["ingress"].(map[string]interface{})
	pvcConfig, _ := config["pvc"].(map[string]interface{})

	enableDeployment, _ := config["enableDeployment"].(bool)
	enableService, _ := config["enableService"].(bool)
	enableIngress, _ := config["enableIngress"].(bool)
	enablePvc, _ := config["enablePvc"].(bool)

	replicasInt32 := int32(3)
	if replicas > 0 {
		replicasInt32 = replicas
	}

	// -------------------------------
	// Deployment 基础构建
	// -------------------------------
	deployment := &kubev1alpha1.DeploymentSpec{
		Name:     name,
		Image:    image,
		Replicas: &replicasInt32,
	}

	// Ports
	if portsConfig, ok := deploymentConfig["ports"].([]interface{}); ok {
		for _, port := range portsConfig {
			portConfig := port.(map[string]interface{})
			deployment.Ports = append(deployment.Ports, corev1.ContainerPort{
				ContainerPort: int32(portConfig["containerPort"].(float64)),
				Name:          portConfig["name"].(string),
				Protocol:      corev1.Protocol(portConfig["protocol"].(string)),
			})
		}
	}

	// LivenessProbe
	if lp, ok := deploymentConfig["livenessProbe"].(map[string]interface{}); ok {
		deployment.LivenessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: lp["httpGet"].(map[string]interface{})["path"].(string),
					Port: intstr.FromInt(int(lp["httpGet"].(map[string]interface{})["port"].(float64))),
				},
			},
			InitialDelaySeconds: int32(lp["initialDelaySeconds"].(float64)),
			PeriodSeconds:       int32(lp["periodSeconds"].(float64)),
		}
	}

	// ReadinessProbe
	if rp, ok := deploymentConfig["readinessProbe"].(map[string]interface{}); ok {
		deployment.ReadinessProbe = &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: rp["httpGet"].(map[string]interface{})["path"].(string),
					Port: intstr.FromInt(int(rp["httpGet"].(map[string]interface{})["port"].(float64))),
				},
			},
			InitialDelaySeconds: int32(rp["initialDelaySeconds"].(float64)),
			PeriodSeconds:       int32(rp["periodSeconds"].(float64)),
		}
	}

	// Lifecycle
	if lifecycleConfig, ok := deploymentConfig["lifecycle"].(map[string]interface{}); ok {
		if preStopConfig, ok := lifecycleConfig["preStop"].(map[string]interface{}); ok {
			cmds := []string{}
			for _, c := range preStopConfig["exec"].(map[string]interface{})["command"].([]interface{}) {
				cmds = append(cmds, c.(string))
			}
			deployment.Lifecycle = &corev1.Lifecycle{
				PreStop: &corev1.LifecycleHandler{
					Exec: &corev1.ExecAction{Command: cmds},
				},
			}
		}
	}

	// NodeSelector
	if nodeSelector, ok := deploymentConfig["nodeSelector"].(map[string]interface{}); ok {
		deployment.NodeSelector = map[string]string{}
		for k, v := range nodeSelector {
			deployment.NodeSelector[k] = v.(string)
		}
	}

	// Affinity
	if affinityConfig, ok := deploymentConfig["affinity"].(map[string]interface{}); ok {
		if pa, ok := affinityConfig["podAntiAffinity"].(map[string]interface{}); ok {
			if preferredList, ok := pa["preferredDuringSchedulingIgnoredDuringExecution"].([]interface{}); ok && len(preferredList) > 0 {
				first := preferredList[0].(map[string]interface{})
				weight := int32(first["weight"].(float64))
				deployment.Affinity = &corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
							{
								Weight: weight,
								PodAffinityTerm: corev1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{name},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				}
			}
		}
	}

	// DNSConfig
	if dnsConfig, ok := deploymentConfig["dnsConfig"].(map[string]interface{}); ok && len(dnsConfig) > 0 {
		nameservers := []string{}
		for _, ns := range dnsConfig["nameservers"].([]interface{}) {
			nameservers = append(nameservers, ns.(string))
		}
		deployment.DNSConfig = &corev1.PodDNSConfig{Nameservers: nameservers}
	}

	// ImagePullSecrets
	if imageSecrets, ok := deploymentConfig["imagePullSecrets"].([]interface{}); ok {
		for _, item := range imageSecrets {
			m := item.(map[string]interface{})
			deployment.ImagePullSecrets = append(deployment.ImagePullSecrets, corev1.LocalObjectReference{Name: m["name"].(string)})
		}
	}

	// Resources
	if resources, ok := deploymentConfig["resources"].(map[string]interface{}); ok {
		deployment.Resources = &corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(resources["limits"].(map[string]interface{})["cpu"].(string)),
				corev1.ResourceMemory: resource.MustParse(resources["limits"].(map[string]interface{})["memory"].(string)),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(resources["requests"].(map[string]interface{})["cpu"].(string)),
				corev1.ResourceMemory: resource.MustParse(resources["requests"].(map[string]interface{})["memory"].(string)),
			},
		}
	}

	// Env
	if envs, ok := deploymentConfig["env"].([]interface{}); ok {
		for _, e := range envs {
			envMap := e.(map[string]interface{})
			env := corev1.EnvVar{Name: envMap["name"].(string)}
			if v, ok := envMap["value"].(string); ok {
				env.Value = v
			}
			deployment.Env = append(deployment.Env, env)
		}
	}

	// Volumes
	hostPathType := corev1.HostPathDirectoryOrCreate
	if vols, ok := deploymentConfig["volumes"].([]interface{}); ok {
		for _, v := range vols {
			vm := v.(map[string]interface{})
			deployment.Volumes = append(deployment.Volumes, kubev1alpha1.VolumeConfig{
				Name: vm["name"].(string),
				HostPath: &corev1.HostPathVolumeSource{
					Path: vm["hostPath"].(map[string]interface{})["path"].(string),
					Type: &hostPathType,
				},
			})
		}
	}

	// VolumeMounts
	if mounts, ok := deploymentConfig["volumeMounts"].([]interface{}); ok {
		for _, m := range mounts {
			mm := m.(map[string]interface{})
			deployment.VolumeMounts = append(deployment.VolumeMounts, kubev1alpha1.VolumeMount{
				Name:      mm["name"].(string),
				MountPath: mm["mountPath"].(string),
			})
		}
	}

	// -------------------------------
	// Service
	// -------------------------------
	service := &kubev1alpha1.ServiceSpec{}
	if serviceConfig != nil && len(serviceConfig) > 0 {
		service.Name = getString(serviceConfig, "name")
		service.Port = int32(getFloat(serviceConfig, "port"))
		service.TargetPort = int32(getFloat(serviceConfig, "targetPort"))
		t := getString(serviceConfig, "type")
		if t != "" {
			service.Type = corev1.ServiceType(t)
		}
	}

	// -------------------------------
	// Ingress
	// -------------------------------
	ingress := &kubev1alpha1.IngressSpec{}
	if ingressConfig != nil && len(ingressConfig) > 0 {
		ingress.Name = getString(ingressConfig, "name")
		ingress.Host = getString(ingressConfig, "host")
		ingress.ServiceName = getString(ingressConfig, "serviceName")
		ingress.ServicePort = int32(getFloat(ingressConfig, "servicePort"))
		ingress.Path = getString(ingressConfig, "path")
		pt := getString(ingressConfig, "pathType")
		if pt != "" {
			ingress.PathType = networkingv1.PathType(pt)
		}
		ingress.IngressClassName = getString(ingressConfig, "ingressClassName")
	}

	// -------------------------------
	// PVC 安全构建
	// -------------------------------
	pvc := &kubev1alpha1.PvcSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
	}
	if pvcConfig != nil && len(pvcConfig) > 0 {
		pvc.Name = getString(pvcConfig, "name")
		pvc.Storage = getString(pvcConfig, "storage")
		sc := getString(pvcConfig, "storageClassName")
		if sc != "" {
			pvc.StorageClassName = pointer.String(sc)
		}
		if fd, ok := pvcConfig["forceDelete"].(bool); ok {
			pvc.ForceDelete = fd
		}
		if rawModes, ok := pvcConfig["accessModes"]; ok {
			if arr, ok := rawModes.([]interface{}); ok && len(arr) > 0 {
				modes := []corev1.PersistentVolumeAccessMode{}
				for _, m := range arr {
					if s, ok := m.(string); ok {
						switch s {
						case "ReadWriteOnce":
							modes = append(modes, corev1.ReadWriteOnce)
						case "ReadOnlyMany":
							modes = append(modes, corev1.ReadOnlyMany)
						case "ReadWriteMany":
							modes = append(modes, corev1.ReadWriteMany)
						}
					}
				}
				if len(modes) > 0 {
					pvc.AccessModes = modes
				}
			}
		}
	}

	return &kubev1alpha1.KubeApp{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.kube.com/v1alpha1",
			Kind:       "KubeApp",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kubev1alpha1.KubeAppSpec{
			EnableDeployment: enableDeployment,
			EnableService:    enableService,
			EnableIngress:    enableIngress,
			EnablePvc:        enablePvc,
			Deployment:       deployment,
			Service:          service,
			Ingress:          ingress,
			Pvc:              pvc,
		},
	}
}


// ---- 辅助函数 ----
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}
