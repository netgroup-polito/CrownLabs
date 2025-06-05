package v1alpha2

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ServicePortMapping struct {
	Name         string `json:"name"`
	Port         int32  `json:"port"`
	TargetPort   int32  `json:"targetPort"`
	AssignedPort int32  `json:"assignedPort,omitempty"`
}

type InstanceRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type InstanceServiceExposureSpec struct {
	InstanceRef InstanceRef          `json:"instanceRef,omitempty"`
	Services    []ServicePortMapping `json:"services"`
}

type InstanceServiceExposureStatus struct {
	ExternalIP    string               `json:"externalIP,omitempty"`
	AssignedPorts []ServicePortMapping `json:"assignedPorts,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type InstanceServiceExposure struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceServiceExposureSpec   `json:"spec,omitempty"`
	Status InstanceServiceExposureStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type InstanceServiceExposureList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InstanceServiceExposure `json:"items"`
}
