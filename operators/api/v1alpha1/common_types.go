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
	// SubscrOk -> the subscription was successful
	SubscrOk SubscriptionStatus = "Ok"
	// SubscrPending -> the subscription is in process
	SubscrPending SubscriptionStatus = "Pending"
	// SubscrFailed -> the subscription has failed
	SubscrFailed SubscriptionStatus = "Failed"
)
