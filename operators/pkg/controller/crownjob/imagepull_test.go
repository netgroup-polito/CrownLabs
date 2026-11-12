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

package crownjob_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/crownjob"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var eventCounter = 0

// createPodEvent creates an event about a container of the given pod, as the kubelet records it (legacy events, whose
// times have a precision of one second).
func createPodEvent(pod *corev1.Pod, container, reason, message string, at time.Time) {
	eventCounter++
	fieldPath := fmt.Sprintf("spec.containers{%s}", container)
	if container == forge.CrownJobRunFolderContainerName {
		fieldPath = fmt.Sprintf("spec.initContainers{%s}", container)
	}
	eventType := corev1.EventTypeNormal
	if reason == "Failed" || reason == "BackOff" {
		eventType = corev1.EventTypeWarning
	}
	timestamp := metav1.NewTime(at.Truncate(time.Second))
	ExpectWithOffset(1, k8sClient.Create(ctx, &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s.%d", pod.Name, eventCounter), Namespace: pod.Namespace},
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", APIVersion: "v1", Namespace: pod.Namespace, Name: pod.Name,
			UID: pod.UID, FieldPath: fieldPath},
		Reason: reason, Message: message, Type: eventType,
		Source:         corev1.EventSource{Component: "kubelet", Host: "node-1"},
		FirstTimestamp: timestamp, LastTimestamp: timestamp, Count: 1,
	})).To(Succeed())
}

var _ = Describe("Telling whether the image of a starting job is being downloaded (P11)", func() {
	sh := []string{"sh", "-c", "true"}
	pulling := "Pulling image \"" + image + "\""
	pulled := "Successfully pulled image \"" + image + "\" in 3.2s (3.2s including waiting). Image size: 123456 bytes."
	present := "Container image \"" + image + "\" already present on machine and can be accessed by the pod"

	type podEvent struct {
		container, reason, message string
		// seconds is the time of the event, in seconds after a base time.
		seconds int
	}
	type pullCase struct {
		events []podEvent
		// initWaiting is the reason the run-folder init container waits for (PodInitializing if empty).
		initWaiting string
		// otherPod tells to record the events about another pod of the namespace.
		otherPod bool
		// initCreated tells that the run-folder init container has been created (and has ended).
		initCreated bool
		message     string
	}

	DescribeTable("Reading the waiting containers and the events of the kubelet about the job pod",
		func(c pullCase) {
			cj, job := startJob("inst", sh)
			initWaiting := c.initWaiting
			if initWaiting == "" {
				initWaiting = "PodInitializing"
			}
			initStatus := waiting(forge.CrownJobRunFolderContainerName, initWaiting)
			if c.initCreated {
				initStatus = terminated(forge.CrownJobRunFolderContainerName, 0)
			}
			pod := createPod(job, &corev1.PodStatus{
				Phase:                 corev1.PodPending,
				Conditions:            []corev1.PodCondition{{Type: corev1.PodReadyToStartContainers, Status: corev1.ConditionTrue}},
				InitContainerStatuses: []corev1.ContainerStatus{initStatus},
				ContainerStatuses:     []corev1.ContainerStatus{waiting(envName, "PodInitializing")},
			})

			about := pod
			if c.otherPod {
				about = &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: job.Name + "-other", Namespace: tenantNamespace},
					Spec: *job.Spec.Template.Spec.DeepCopy()}
				create(about)
			}
			base := time.Now().Add(-time.Minute)
			for _, e := range c.events {
				createPodEvent(about, e.container, e.reason, e.message, base.Add(time.Duration(e.seconds)*time.Second))
			}

			reconcileCrownJob(cj)
			expectPhase(cj, clv1alpha2.CrownJobPhaseStarting, crownjob.ReasonAdmitted, c.message)
		},
		Entry("no event: the containers wait to be created, as with the image on the node (P10)", pullCase{
			message: "Your job is starting."}),
		Entry("Pulling, and no Pulled yet", pullCase{
			events:  []podEvent{{forge.CrownJobRunFolderContainerName, "Pulling", pulling, 0}},
			message: "Your job is starting (downloading the image)."}),
		Entry("Pulling, then Pulled two seconds later", pullCase{
			events: []podEvent{{forge.CrownJobRunFolderContainerName, "Pulling", pulling, 0},
				{forge.CrownJobRunFolderContainerName, "Pulled", pulled, 2}},
			message: "Your job is starting."}),
		Entry("Pulling and Pulled in the same second (a pull shorter than a second)", pullCase{
			events: []podEvent{{forge.CrownJobRunFolderContainerName, "Pulled", pulled, 0},
				{forge.CrownJobRunFolderContainerName, "Pulling", pulling, 0}},
			message: "Your job is starting."}),
		Entry("Pulling and Pulled in the same second, recorded in the other order", pullCase{
			events: []podEvent{{forge.CrownJobRunFolderContainerName, "Pulling", pulling, 0},
				{forge.CrownJobRunFolderContainerName, "Pulled", pulled, 0}},
			message: "Your job is starting."}),
		Entry("Pulled: the image already present on the node", pullCase{
			events:  []podEvent{{forge.CrownJobRunFolderContainerName, "Pulled", present, 0}},
			message: "Your job is starting."}),
		Entry("Pulling, then the pull failed", pullCase{
			events: []podEvent{{forge.CrownJobRunFolderContainerName, "Pulling", pulling, 0},
				{forge.CrownJobRunFolderContainerName, "Failed", "Failed to pull image \"" + image + "\": not found", 1}},
			message: "Your job is starting (the image cannot be downloaded: retrying)."}),
		Entry("a back-off after a failed pull", pullCase{
			events: []podEvent{{forge.CrownJobRunFolderContainerName, "Pulling", pulling, 0},
				{forge.CrownJobRunFolderContainerName, "BackOff", "Back-off pulling image \"" + image + "\"", 1}},
			message: "Your job is starting (the image cannot be downloaded: retrying)."}),
		Entry("a Failed event that is not about a pull, after the image was pulled", pullCase{
			events: []podEvent{{forge.CrownJobRunFolderContainerName, "Pulled", present, 0},
				{forge.CrownJobRunFolderContainerName, "Failed", "Error: failed to create containerd task: exit status 1", 1}},
			message: "Your job is starting."}),
		Entry("the init container pulled, and the app container pulling another image in the same second", pullCase{
			events: []podEvent{{forge.CrownJobRunFolderContainerName, "Pulling", pulling, 0},
				{forge.CrownJobRunFolderContainerName, "Pulled", pulled, 5}, {envName, "Pulling", "Pulling image \"other:v1\"", 5}},
			message: "Your job is starting (downloading the image)."}),
		Entry("Pulling about another pod of the namespace (the events are read by the UID of the job pod)", pullCase{
			events:   []podEvent{{forge.CrownJobRunFolderContainerName, "Pulling", pulling, 0}},
			otherPod: true,
			message:  "Your job is starting."}),
		Entry("Pulling with no Pulled (the event was lost), for a container already created", pullCase{
			events:      []podEvent{{forge.CrownJobRunFolderContainerName, "Pulling", pulling, 0}},
			initCreated: true,
			message:     "Your job is starting."}),
		Entry("a container waiting in ErrImagePull, before any event", pullCase{
			initWaiting: "ErrImagePull",
			message:     "Your job is starting (the image cannot be downloaded: retrying)."}),
		Entry("Pulling again (a retry) while the container waits in ImagePullBackOff", pullCase{
			events: []podEvent{{forge.CrownJobRunFolderContainerName, "Failed", "Failed to pull image \"" + image + "\": timeout", 0},
				{forge.CrownJobRunFolderContainerName, "Pulling", pulling, 20}},
			initWaiting: "ImagePullBackOff",
			message:     "Your job is starting (the image cannot be downloaded: retrying)."}),
	)
})
