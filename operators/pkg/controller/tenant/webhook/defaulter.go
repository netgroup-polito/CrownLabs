package webhook

import "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

type TenantDefaulter struct {
	admission.CustomDefaulter
}
