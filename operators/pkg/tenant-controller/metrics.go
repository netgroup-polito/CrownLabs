package tenant_controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	tnOpinternalErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tenant_operator_internal_errors",
		Help: "The number of errors occurred internally during the reconcile of the tenant operator",
	},
		[]string{"controller", "reason"},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(tnOpinternalErrors)
}
