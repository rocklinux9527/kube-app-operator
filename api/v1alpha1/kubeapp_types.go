/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubeAppSpec defines the desired state of KubeApp.
type KubeAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	EnableDeployment bool            `json:"enableDeployment,omitempty"` // if deployments  json defines
	EnableService    bool            `json:"enableService,omitempty"`    // if service json defines
	EnableIngress    bool            `json:"enableIngress,omitempty"`    // if ingress json defines
	EnablePvc        bool            `json:"enablePvc"`                  // if pvc json define
	Pvc              *PvcSpec        `json:"pvc,omitempty"`
	Deployment       *DeploymentSpec `json:"deployment,omitempty"` // Citation deployments struct
	Service          *ServiceSpec    `json:"service,omitempty"`    // Citation service  struct
	Ingress          *IngressSpec    `json:"ingress,omitempty"`    // Citation ingress struct
}

type DeploymentSpec struct {
	Name      string                     `json:"name"`    // defines deployment property
	Image     string                     `json:"image"`
	Replicas  *int32                      `json:"replicas"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	Ports           []corev1.ContainerPort  `json:"ports,omitempty"`

	LivenessProbe   *corev1.Probe        `json:"livenessProbe,omitempty"`
	ReadinessProbe  *corev1.Probe        `json:"readinessProbe,omitempty"`
	Lifecycle       *corev1.Lifecycle    `json:"lifecycle,omitempty"`

	NodeSelector    map[string]string    `json:"nodeSelector,omitempty"`
	Volumes         []VolumeConfig       `json:"volumes,omitempty"`
	VolumeMounts    []VolumeMount  `json:"volumeMounts,omitempty"`

	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	DNSConfig    *corev1.PodDNSConfig  `json:"dnsConfig,omitempty"`

	Affinity    *corev1.Affinity  `json:"affinity,omitempty"`
	Env []corev1.EnvVar `json:"env,omitempty"`
}

// defines service spec field object

type ServiceSpec struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"targetPort"`
	Type       corev1.ServiceType `json:"type"`
}

// defines ingress spec field object

type IngressSpec struct {
	Name        string `json:"name,omitempty"`
	Host        string `json:"host,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	ServicePort int32  `json:"service_port,omitempty"`
	Path        string  `json:"path,omitempty"`
	PathType    networkingv1.PathType `json:"path_type,omitempty"`
	IngressClassName string                  `json:"ingressClassName,omitempty"`

}

// volumeMount define

type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}


type VolumeConfig struct {
	// 卷名称
	Name string `json:"name"`
	// 卷类型
	// Type string `json:"type,omitempty"` //auto type
	// PersistentVolumeClaim 配置
	PersistentVolumeClaim *corev1.PersistentVolumeClaimVolumeSource `json:"persistentVolumeClaim,omitempty"`
	// ConfigMap 配置
	ConfigMap *corev1.ConfigMapVolumeSource `json:"configMap,omitempty"`
	// Secret 配置
	Secret *corev1.SecretVolumeSource `json:"secret,omitempty"`
	// EmptyDir 配置
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
	// HostPath 配置
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`
	// NFS 配置
	NFS    *corev1.NFSVolumeSource  `json:"nfs,omitempty"`
}


type PvcSpec struct {
	Name             string                              `json:"name"`
	Storage          string                              `json:"storage"` // e.g. "1Gi"
	AccessModes      []corev1.PersistentVolumeAccessMode `json:"accessModes"`
	StorageClassName *string                             `json:"storageClassName,omitempty"`
	ForceDelete bool `json:"forceDelete,omitempty"`  // 是否强制删除 PVC
}



// KubeAppStatus defines the observed state of KubeApp.
type KubeAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KubeApp is the Schema for the kubeapps API.
type KubeApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeAppSpec   `json:"spec,omitempty"`
	Status KubeAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KubeAppList contains a list of KubeApp.
type KubeAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeApp{}, &KubeAppList{})
}
