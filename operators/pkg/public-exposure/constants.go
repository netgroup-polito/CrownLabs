package publicexposure

const (
	// PublicExposureComponentLabelKey is the key for the label that identifies resources
	// managed by the public exposure feature.
	PublicExposureComponentLabelKey = "crownlabs.polito.it/component"

	// PublicExposureComponentLabelValue is the value for the label that identifies
	// LoadBalancer services created by this feature.
	PublicExposureComponentLabelValue = "public-exposure-loadbalancer"

	// MetallbAddressPoolAnnotation is the annotation key used by MetalLB to select an address pool.
	MetallbAddressPoolAnnotation = "metallb.universe.tf/address-pool"

	// MetallbAllowSharedIPAnnotation is the annotation key used by MetalLB to allow sharing an IP among services.
	MetallbAllowSharedIPAnnotation = "metallb.universe.tf/allow-shared-ip"

	// MetallbLoadBalancerIPsAnnotation is the annotation key used by MetalLB to specify a desired IP.
	MetallbLoadBalancerIPsAnnotation = "metallb.universe.tf/loadBalancerIPs"

	// DefaultAddressPool is the default MetalLB address pool to use if not specified elsewhere.
	DefaultAddressPool = "my-ip-pool"

	// AllowSharedIPValue is the string value to enable IP sharing in MetalLB.
	AllowSharedIPValue = "true"

	// BasePortForAutomaticAssignment is the starting port number for automatic port allocation.
	BasePortForAutomaticAssignment = 30000
)
