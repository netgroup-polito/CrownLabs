package v1alpha1

// NameCreated contains info about the status of a resource
type NameCreated struct {
	Name    string `json:"name,omitempty"`
	Created bool   `json:"created"`
}

// SubscriptionStatus is an enum for the status of a subscription to a service
// +kubebuilder:validation:Enum=Ok;Failed
type SubscriptionStatus string

const (
	// SubscrOk -> the subscription was successful
	SubscrOk SubscriptionStatus = "Ok"
	// SubscrFailed -> the subscription has failed
	SubscrFailed SubscriptionStatus = "Failed"
)

// GenericRef stores generric data to point to a kubernetes resource
type GenericRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// WorkspaceLabelPrefix is the prefix of a label assigned to a tenant indicating it is subscribed to a workspace
const WorkspaceLabelPrefix = "crownlabs.polito.it/workspace-"

// TnOperatorFinalizerName is the name of the finalizer corresponding to the tenant operator
const TnOperatorFinalizerName = "crownlabs.polito.it/tenant-operator"
