package v1alpha1

// NameCreated contains info about the status of a resource
type NameCreated struct {
	Name    string `json:"name,omitempty"`
	Created bool   `json:"created"`
}

// SubscriptionStatus is an enum for the status of a subscription to a service
// +kubebuilder:validation:Enum=Ok;Pending;Failed
type SubscriptionStatus string

const (
	// Ok -> the subscription was successful
	Ok SubscriptionStatus = "Ok"
	// Pending -> the subscription is in process
	Pending SubscriptionStatus = "Pending"
	// Failed -> the subscription has failed
	Failed SubscriptionStatus = "Failed"
)
