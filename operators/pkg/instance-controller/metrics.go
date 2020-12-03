package instance_controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	bootTimes = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "vmi_boot_time_seconds",
		Help:    "The time required to boot for spawned VMs",
		Buckets: prometheus.LinearBuckets(30, 10, 20),
	})
	elaborationTimes = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "vmi_elaboration_time_seconds",
		Help:    "The time required to the operator logic to handle VMIs",
		Buckets: prometheus.LinearBuckets(0.5, 0.2, 20),
	})
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(bootTimes, elaborationTimes)
}
