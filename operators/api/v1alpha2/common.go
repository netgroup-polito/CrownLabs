package v1alpha2

type GenericRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}
