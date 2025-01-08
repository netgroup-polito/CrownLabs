// Copyright 2020-2025 Politecnico di Torino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package instctrl

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	metricInitialReadyTimesLabelWorkspace   = "workspace"
	metricInitialReadyTimesLabelTemplate    = "template"
	metricInitialReadyTimesLabelEnvironment = "environment"
	metricInitialReadyTimesLabelType        = "type"
	metricInitialReadyTimesLabelPersistent  = "persistent"
)

var (
	metricInitialReadyTimes = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "instance_initial_ready_time_second",
		Help: "The number of seconds required by the Instances to become ready upon creation",
		// Buckets (upper bound in seconds): 5, 10, 15, .., 50, 55, 60, 90, 120, .. 570, 600
		Buckets: append(prometheus.LinearBuckets(5, 5, 11), prometheus.LinearBuckets(60, 30, 19)...),
	},
		[]string{
			metricInitialReadyTimesLabelWorkspace,
			metricInitialReadyTimesLabelTemplate,
			metricInitialReadyTimesLabelEnvironment,
			metricInitialReadyTimesLabelType,
			metricInitialReadyTimesLabelPersistent,
		},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(metricInitialReadyTimes)
}
