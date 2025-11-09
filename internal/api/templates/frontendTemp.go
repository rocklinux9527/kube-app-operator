
package templates

import (
    kubev1alpha1 "github.com/k8s/kube-app-operator/api/v1alpha1"
    corev1 "k8s.io/api/core/v1"
    networkingv1 "k8s.io/api/networking/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func FrontendTemplate(name, namespace, image string, replicas int32 ) *kubev1alpha1.KubeApp {

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
            EnablePvc:        false,
            Deployment: &kubev1alpha1.DeploymentSpec{
                Name:     name,
                Image:    image,
                Replicas: &replicas,
                Ports: []corev1.ContainerPort{
                    {
                        ContainerPort: 80,
                        Name:          "http",
                        Protocol:      corev1.ProtocolTCP,
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
        },
    }
}
