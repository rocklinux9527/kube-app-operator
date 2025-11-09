
package templates

import (
    kubev1alpha1 "github.com/k8s/kube-app-operator/api/v1alpha1"
    corev1 "k8s.io/api/core/v1"
    networkingv1 "k8s.io/api/networking/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/util/intstr"
    "k8s.io/utils/pointer"
)


func BackendTemplate(name, namespace, image string, replicas int32) *kubev1alpha1.KubeApp {
    hostPathType := corev1.HostPathDirectoryOrCreate
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
            EnableDeployment: true,
            EnableService:    true,
            EnableIngress:    true,
            EnablePvc:        true,
            Deployment: &kubev1alpha1.DeploymentSpec{
                Name:     name,
                Image:    image,
                Replicas: &replicas,
                Ports: []corev1.ContainerPort{
                    {
                        ContainerPort: 8080,
                        Name:          "http",
                        Protocol:      corev1.ProtocolTCP,
                    },
                },
                LivenessProbe: &corev1.Probe{
                    ProbeHandler: corev1.ProbeHandler{
                        HTTPGet: &corev1.HTTPGetAction{
                            Path: "/",
                            Port: intstr.FromInt(80),
                        },
                    },
                    InitialDelaySeconds: 5,
                    PeriodSeconds:       10,
                },
                ReadinessProbe: &corev1.Probe{
                    ProbeHandler: corev1.ProbeHandler{
                        HTTPGet: &corev1.HTTPGetAction{
                            Path: "/",
                            Port: intstr.FromInt(80),
                        },
                    },
                    InitialDelaySeconds: 5,
                    PeriodSeconds:       5,
                },
                Lifecycle: &corev1.Lifecycle{
                    PreStop: &corev1.LifecycleHandler{
                        Exec: &corev1.ExecAction{
                            Command: []string{"sleep", "20"},
                        },
                    },
                },
                NodeSelector: map[string]string{
                    "kubernetes.io/os": "linux",
                },
                Affinity: &corev1.Affinity{
                    PodAntiAffinity: &corev1.PodAntiAffinity{
                        PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
                            {
                                Weight: 100,
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
                },
                DNSConfig: &corev1.PodDNSConfig{
		    Nameservers: []string{"8.8.8.8"},
		},
                ImagePullSecrets: []corev1.LocalObjectReference{
		 {
		   Name: "private-registry",
		 },
		}, 
                Resources: &corev1.ResourceRequirements{
                    Limits: corev1.ResourceList{
                        corev1.ResourceCPU:    resource.MustParse("500m"),
                        corev1.ResourceMemory: resource.MustParse("512Mi"),
                    },
                    Requests: corev1.ResourceList{
                        corev1.ResourceCPU:    resource.MustParse("250m"),
                        corev1.ResourceMemory: resource.MustParse("256Mi"),
                    },
                },
                Env: []corev1.EnvVar{
                    {
                        Name:  "user-address",
                        Value: "nacos-headless.middleware.svc.cluster.local",
                    },
                    {
                        Name: "POD_IP",
                        ValueFrom: &corev1.EnvVarSource{
                            FieldRef: &corev1.ObjectFieldSelector{
                                FieldPath: "status.podIP",
                            },
                        },
                    },
                },
                TerminationGracePeriodSeconds: pointer.Int64(60),
                Volumes: []kubev1alpha1.VolumeConfig{
                    {
                        Name: "data-host",
                        HostPath: &corev1.HostPathVolumeSource{
                            Path: "/opt/k8s_host_path",
                            Type: &hostPathType,
                        },
                    },
                },
                VolumeMounts: []kubev1alpha1.VolumeMount{
                    {
                        Name:      "data-host",
                        MountPath: "/tmp/",
                    },
                },
            },
            Service: &kubev1alpha1.ServiceSpec{
                Name:       name,
                Port:       80,
                TargetPort: 80,
                Type:       corev1.ServiceTypeClusterIP,
            },
            Ingress: &kubev1alpha1.IngressSpec{
                Name:             name,
                Host:             "www.example.com",
                ServiceName:      name,
                ServicePort:      80,
                Path:             "/",
                PathType:         networkingv1.PathTypePrefix,
                IngressClassName: "alb",
            },
            Pvc: &kubev1alpha1.PvcSpec{
                Name:             name,
                Storage:          "100G",
                AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
                StorageClassName: pointer.String("alicloud-disk-ssd"),
                ForceDelete:      true,
            },
        },
    }
}


