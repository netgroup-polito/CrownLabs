// Copyright 2020-2026 Politecnico di Torino
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

package main

import (
	"flag"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	ctrlcommon "github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/crownjob"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

var (
	crownJobRuntimeClasses          string
	crownJobDefaultMaxRunTime       string
	crownJobLogHelperImage          string
	crownJobFinishedJobTTL          string
	maxConcurrentCrownJobReconciles int
)

const (
	crownJobCtrl = "CrownJob"
)

func init() {
	flag.StringVar(&crownJobRuntimeClasses, "crownjob-runtime-classes", "",
		"The RuntimeClass of the jobs that request a given resource, as a comma-separated list of resource=runtimeClass (e.g., nvidia.com/gpu=nvidia)")
	flag.StringVar(&crownJobDefaultMaxRunTime, "crownjob-default-max-run-time", "2h",
		"The maximum run time of a job when neither its Template nor its Workspace set one, as a number followed by m, h or d")
	flag.StringVar(&crownJobLogHelperImage, "crownjob-log-helper-image", "registry.k8s.io/kubectl:v1.34.1",
		"The image of the Jobs that save the log of each job pod into its run folder (it must contain /go-runner and /bin/kubectl)")
	flag.StringVar(&crownJobFinishedJobTTL, "crownjob-finished-job-ttl", "10m",
		"How long the Job of a CrownJob, and each of its log helpers, is kept after it ends, with its pods, as a number followed by m, h or d (the CrownJob is kept)")
	flag.IntVar(&maxConcurrentCrownJobReconciles, "crownjob-max-concurrent-reconciles", 1,
		"The maximum number of concurrent Reconciles which can be run for the CrownJob controllers")
}

// setupCrownJob configures the CrownJob controller, the controller keeping the Instances while their jobs run,
// and the controller of the job queues of the Workspaces.
func setupCrownJob(mgr manager.Manager, targetLabel ctrlcommon.KVLabel) error {
	runtimeClasses, err := forge.MapFromKVString(crownJobRuntimeClasses)
	if err != nil {
		return fmt.Errorf("invalid value of --crownjob-runtime-classes: %w", err)
	}
	defaultMaxRunTime, err := utils.ParseDurationWithDays(crownJobDefaultMaxRunTime)
	if err != nil {
		return fmt.Errorf("invalid value of --crownjob-default-max-run-time: %w", err)
	}
	finishedJobTTL, err := utils.ParseDurationWithDays(crownJobFinishedJobTTL)
	if err != nil {
		return fmt.Errorf("invalid value of --crownjob-finished-job-ttl: %w", err)
	}
	if err := checkKueueKinds(mgr); err != nil {
		return err
	}

	if err := (&crownjob.Reconciler{
		Client:            mgr.GetClient(),
		APIReader:         mgr.GetAPIReader(),
		TargetLabel:       targetLabel,
		EventsRecorder:    mgr.GetEventRecorderFor(crownJobCtrl),
		RuntimeClasses:    runtimeClasses,
		DefaultMaxRunTime: defaultMaxRunTime,
		LogHelperImage:    crownJobLogHelperImage,
		FinishedJobTTL:    finishedJobTTL,
	}).SetupWithManager(mgr, maxConcurrentCrownJobReconciles); err != nil {
		return err
	}

	if err := (&crownjob.InstanceReconciler{
		Client:         mgr.GetClient(),
		APIReader:      mgr.GetAPIReader(),
		TargetLabel:    targetLabel,
		EventsRecorder: mgr.GetEventRecorderFor(crownJobCtrl),
	}).SetupWithManager(mgr, maxConcurrentCrownJobReconciles); err != nil {
		return err
	}

	return (&crownjob.QueueReconciler{
		Client:      mgr.GetClient(),
		TargetLabel: targetLabel,
	}).SetupWithManager(mgr, maxConcurrentCrownJobReconciles)
}

// checkKueueKinds returns an error if the cluster does not serve the Kueue kinds used by the CrownJob controllers.
func checkKueueKinds(mgr manager.Manager) error {
	for _, gvk := range []schema.GroupVersionKind{
		forge.KueueClusterQueueGVK, forge.KueueLocalQueueGVK, forge.KueueResourceFlavorGVK,
		forge.KueueWorkloadPriorityClassGVK, forge.KueueWorkloadGVK,
	} {
		if _, err := mgr.GetRESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version); err != nil {
			return fmt.Errorf("the CrownJob controller needs Kueue, but the cluster does not serve %s %s (%w): install Kueue, or set --enable-crownjob=false",
				gvk.Kind, gvk.GroupVersion(), err)
		}
	}
	return nil
}
