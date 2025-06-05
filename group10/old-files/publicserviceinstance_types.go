package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PublicServiceInstanceSpec defines the desired state of PublicServiceInstance
type PublicServiceInstanceSpec struct {
	ServiceName string `json:"serviceName"`
	Port        int32  `json:"port"`
	TargetPort  int32  `json:"targetPort"`
	Image       string `json:"image"` // Docker image for the Pod
}

// PublicServiceInstanceStatus defines the observed state of PublicServiceInstance
type PublicServiceInstanceStatus struct {
	ExternalIP string `json:"externalIP,omitempty"` // Assigned external IP
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PublicServiceInstance is the Schema for the publicserviceinstances API
type PublicServiceInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PublicServiceInstanceSpec   `json:"spec,omitempty"`
	Status PublicServiceInstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PublicServiceInstanceList contains a list of PublicServiceInstance
type PublicServiceInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PublicServiceInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PublicServiceInstance{}, &PublicServiceInstanceList{})
}
