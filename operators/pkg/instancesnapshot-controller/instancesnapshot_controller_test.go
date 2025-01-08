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

package instancesnapshot_controller_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("InstancesnapshotController", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		InstanceName         = "test-instance"
		WorkingNamespace     = "working-namespace"
		TemplateName         = "test-template"
		TenantName           = "test-tenant"
		InstanceSnapshotName = "isnap-sample"

		timeout  = time.Second * 20
		interval = time.Millisecond * 500
	)

	var (
		workingNs = v1.Namespace{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: WorkingNamespace,
				Labels: map[string]string{
					"test-suite": "true",
				},
			},
			Spec:   v1.NamespaceSpec{},
			Status: v1.NamespaceStatus{},
		}
		templateEnvironment = crownlabsv1alpha2.TemplateSpec{
			WorkspaceRef: crownlabsv1alpha2.GenericRef{},
			PrettyName:   "My Template",
			Description:  "Description of my template",
			EnvironmentList: []crownlabsv1alpha2.Environment{
				{
					Name:       "Env-1",
					GuiEnabled: true,
					Resources: crownlabsv1alpha2.EnvironmentResources{
						CPU:                   1,
						ReservedCPUPercentage: 1,
						Memory:                resource.MustParse("1024M"),
					},
					EnvironmentType: crownlabsv1alpha2.ClassVM,
					Persistent:      true,
					Image:           "crownlabs/vm",
				},
			},
			DeleteAfter: "",
		}
		template = crownlabsv1alpha2.Template{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      TemplateName,
				Namespace: WorkingNamespace,
			},
			Spec:   templateEnvironment,
			Status: crownlabsv1alpha2.TemplateStatus{},
		}
		instance = crownlabsv1alpha2.Instance{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      InstanceName,
				Namespace: WorkingNamespace,
			},
			Spec: crownlabsv1alpha2.InstanceSpec{
				Running: false,
				Template: crownlabsv1alpha2.GenericRef{
					Name:      TemplateName,
					Namespace: WorkingNamespace,
				},
				Tenant: crownlabsv1alpha2.GenericRef{
					Name: TenantName,
				},
			},
			Status: crownlabsv1alpha2.InstanceStatus{},
		}

		instanceSnapshot = crownlabsv1alpha2.InstanceSnapshot{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      InstanceSnapshotName,
				Namespace: WorkingNamespace,
			},
			Spec: crownlabsv1alpha2.InstanceSnapshotSpec{
				Instance: crownlabsv1alpha2.GenericRef{
					Name:      InstanceName,
					Namespace: WorkingNamespace,
				},
				ImageName: "test-image",
			},
		}
	)

	BeforeEach(func() {
		By("Preparing the environment for the new test")
		newNs := workingNs.DeepCopy()
		newTemplate := template.DeepCopy()
		newInstance := instance.DeepCopy()
		By("Creating the namespace where to create instance and template")
		err := k8sClient.Create(ctx, newNs)
		if err != nil && errors.IsAlreadyExists(err) {
			By("Cleaning up the environment")
			By("Deleting template")
			Expect(k8sClient.Delete(ctx, &template)).Should(Succeed())
			By("Deleting instance")
			Expect(k8sClient.Delete(ctx, &instance)).Should(Succeed())
		} else if err != nil {
			Fail(fmt.Sprintf("Unable to create namespace -> %s", err))
		}

		By("By checking that the namespace has been created")
		createdNs := &v1.Namespace{}

		nsLookupKey := types.NamespacedName{Name: WorkingNamespace}
		doesEventuallyExists(ctx, nsLookupKey, createdNs, BeTrue(), timeout, interval)

		By("Creating the template")
		Expect(k8sClient.Create(ctx, newTemplate)).Should(Succeed())

		By("By checking that the template has been created")
		templateLookupKey := types.NamespacedName{Name: TemplateName, Namespace: WorkingNamespace}
		createdTemplate := &crownlabsv1alpha2.Template{}

		doesEventuallyExists(ctx, templateLookupKey, createdTemplate, BeTrue(), timeout, interval)

		By("Creating the instance")
		Expect(k8sClient.Create(ctx, newInstance)).Should(Succeed())

		By("Checking that the instance has been created")
		instanceLookupKey := types.NamespacedName{Name: InstanceName, Namespace: WorkingNamespace}
		createdInstance := &crownlabsv1alpha2.Instance{}

		doesEventuallyExists(ctx, instanceLookupKey, createdInstance, BeTrue(), timeout, interval)
	})

	Context("Creating a snapshot of a persistent VM", func() {
		It("Should start snapshot creation", func() {
			newInstanceSnapshot := instanceSnapshot.DeepCopy()
			newInstanceSnapshot.Name = fmt.Sprintf("isnap-sample-%v", rand.Int())
			checkIsnapSuccessfulCreation(ctx, newInstanceSnapshot, WorkingNamespace, timeout, interval)

			By("Changing the job status to completed and set Job start and end time")
			jobLookupKey := types.NamespacedName{Name: newInstanceSnapshot.Name, Namespace: WorkingNamespace}
			snapjob := &batch.Job{}
			Expect(k8sClient.Get(ctx, jobLookupKey, snapjob)).Should(Succeed())
			snapjob.Status.Conditions = []batch.JobCondition{
				{Type: batch.JobComplete, Status: v1.ConditionTrue},
			}
			snapjob.Status.CompletionTime = &metav1.Time{Time: time.Now()}
			snapjob.Status.StartTime = &metav1.Time{Time: time.Now().Add(-4 * time.Minute)}
			Expect(k8sClient.Status().Update(ctx, snapjob)).Should(Succeed())

			By("Checking if the InstanceSnapshot status is Completed")
			checkIsnapStatus(ctx, newInstanceSnapshot.Name, WorkingNamespace, crownlabsv1alpha2.Completed, timeout, interval)
		})

		It("Should start snapshot creation given an environment name", func() {
			newInstanceSnapshot := instanceSnapshot.DeepCopy()
			newInstanceSnapshot.Name = fmt.Sprintf("isnap-sample-%v", rand.Int())
			newInstanceSnapshot.Spec.Environment.Name = templateEnvironment.EnvironmentList[0].Name
			checkIsnapSuccessfulCreation(ctx, newInstanceSnapshot, WorkingNamespace, timeout, interval)

			By("Changing the job status to completed without setting Job start and end time")
			jobLookupKey := types.NamespacedName{Name: newInstanceSnapshot.Name, Namespace: WorkingNamespace}
			snapjob := &batch.Job{}
			Expect(k8sClient.Get(ctx, jobLookupKey, snapjob)).Should(Succeed())
			snapjob.Status.Conditions = []batch.JobCondition{
				{Type: batch.JobComplete, Status: v1.ConditionTrue},
			}
			Expect(k8sClient.Status().Update(ctx, snapjob)).Should(Succeed())

			By("Checking if the InstanceSnapshot status is Completed")
			checkIsnapStatus(ctx, newInstanceSnapshot.Name, WorkingNamespace, crownlabsv1alpha2.Completed, timeout, interval)
		})
	})

	Context("Testing incorrect environment configurations", func() {
		It("Should fail: the VM is running", func() {
			By("Getting current instance")
			currentInstance := &crownlabsv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: InstanceName, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, instanceLookupKey, currentInstance)).Should(Succeed())

			By("Setting instance as powered on")
			currentInstance.Spec.Running = true
			Expect(k8sClient.Update(ctx, currentInstance)).Should(Succeed())

			newInstanceSnapshot := instanceSnapshot.DeepCopy()
			newInstanceSnapshot.Name = fmt.Sprintf("isnap-sample-%v", rand.Int())
			checkIsnapCreationFailure(ctx, newInstanceSnapshot, WorkingNamespace, timeout, interval)
		})

		It("Should fail: vm is not persistent", func() {
			By("Getting current Template")
			currentTemplate := &crownlabsv1alpha2.Template{}
			templateLookupKey := types.NamespacedName{Name: TemplateName, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, templateLookupKey, currentTemplate)).Should(Succeed())

			By("Setting environment VM as not persistent")
			currentTemplate.Spec.EnvironmentList[0].Persistent = false
			Expect(k8sClient.Update(ctx, currentTemplate)).Should(Succeed())

			newInstanceSnapshot := instanceSnapshot.DeepCopy()
			newInstanceSnapshot.Name = fmt.Sprintf("isnap-sample-%v", rand.Int())
			checkIsnapCreationFailure(ctx, newInstanceSnapshot, WorkingNamespace, timeout, interval)
		})

		It("Should fail: environment is a container", func() {
			By("Getting current Template")
			currentTemplate := &crownlabsv1alpha2.Template{}
			templateLookupKey := types.NamespacedName{Name: TemplateName, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, templateLookupKey, currentTemplate)).Should(Succeed())

			By("Setting environment as Container")
			currentTemplate.Spec.EnvironmentList[0].EnvironmentType = crownlabsv1alpha2.ClassContainer
			Expect(k8sClient.Update(ctx, currentTemplate)).Should(Succeed())

			newInstanceSnapshot := instanceSnapshot.DeepCopy()
			newInstanceSnapshot.Name = fmt.Sprintf("isnap-sample-%v", rand.Int())
			checkIsnapCreationFailure(ctx, newInstanceSnapshot, WorkingNamespace, timeout, interval)
		})

		It("Should fail: template does not exist", func() {
			By("Getting current instance")
			currentInstance := &crownlabsv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: InstanceName, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, instanceLookupKey, currentInstance)).Should(Succeed())

			By("Changing template with a non-existing one")
			currentInstance.Spec.Template.Name = "invalid"
			Expect(k8sClient.Update(ctx, currentInstance)).Should(Succeed())

			newInstanceSnapshot := instanceSnapshot.DeepCopy()
			newInstanceSnapshot.Name = fmt.Sprintf("isnap-sample-%v", rand.Int())
			checkIsnapCreationFailure(ctx, newInstanceSnapshot, WorkingNamespace, timeout, interval)
		})

		It("Should fail: instance does not exist", func() {
			By("Setting not existing instance name")
			newInstanceSnapshot := instanceSnapshot.DeepCopy()
			newInstanceSnapshot.Name = fmt.Sprintf("isnap-sample-%v", rand.Int())
			newInstanceSnapshot.Spec.Instance.Name = "invalid"
			checkIsnapCreationFailure(ctx, newInstanceSnapshot, WorkingNamespace, timeout, interval)
		})
	})

	Context("Testing snapshotting job failures", func() {
		It("Should fail: job failed", func() {
			newInstanceSnapshot := instanceSnapshot.DeepCopy()
			newInstanceSnapshot.Name = fmt.Sprintf("isnap-sample-%v", rand.Int())
			checkIsnapSuccessfulCreation(ctx, newInstanceSnapshot, WorkingNamespace, timeout, interval)

			By("Changing the job status to failed")
			jobLookupKey := types.NamespacedName{Name: newInstanceSnapshot.Name, Namespace: WorkingNamespace}
			snapjob := &batch.Job{}
			Expect(k8sClient.Get(ctx, jobLookupKey, snapjob)).Should(Succeed())
			snapjob.Status.Conditions = []batch.JobCondition{
				{Type: batch.JobFailed, Status: v1.ConditionTrue},
			}
			Expect(k8sClient.Status().Update(ctx, snapjob)).Should(Succeed())

			By("Checking if the InstanceSnapshot status is Failed")
			checkIsnapStatus(ctx, newInstanceSnapshot.Name, WorkingNamespace, crownlabsv1alpha2.Failed, timeout, interval)
		})
	})
})

func checkIsnapStatus(ctx context.Context, isnapName, workingNamespace string, desiredStatus crownlabsv1alpha2.SnapshotStatus, timeout, interval time.Duration) {
	isnapLookupKey := types.NamespacedName{Name: isnapName, Namespace: workingNamespace}
	retrievedIsnap := &crownlabsv1alpha2.InstanceSnapshot{}
	Eventually(func() crownlabsv1alpha2.SnapshotStatus {
		err := k8sClient.Get(ctx, isnapLookupKey, retrievedIsnap)
		if err != nil {
			return ""
		}
		return retrievedIsnap.Status.Phase
	}, timeout, interval).Should(Equal(desiredStatus))
}

func checkIsnapSuccessfulCreation(ctx context.Context, isnap *crownlabsv1alpha2.InstanceSnapshot, workingNamespace string, timeout, interval time.Duration) {
	By("Creating the InstanceSnapshot resource")
	Expect(k8sClient.Create(ctx, isnap)).Should(Succeed())

	By("Checking that the instance snapshot has been created")
	instanceSnapshotLookupKey := types.NamespacedName{Name: isnap.Name, Namespace: workingNamespace}
	createdInstanceSnapshot := &crownlabsv1alpha2.InstanceSnapshot{}

	doesEventuallyExists(ctx, instanceSnapshotLookupKey, createdInstanceSnapshot, BeTrue(), timeout, interval)

	By("Checking if the job for the creation of the snapshot has been created")
	jobLookupKey := types.NamespacedName{Name: isnap.Name, Namespace: workingNamespace}
	createdJob := &batch.Job{}

	doesEventuallyExists(ctx, jobLookupKey, createdJob, BeTrue(), timeout, interval)

	By("Checking the owner reference of the job")
	Expect(createdJob.ObjectMeta.OwnerReferences).To(ContainElement(MatchFields(IgnoreExtras, Fields{
		"UID": Equal(createdInstanceSnapshot.UID),
	})))
}

func checkIsnapCreationFailure(ctx context.Context, isnap *crownlabsv1alpha2.InstanceSnapshot, workingNamespace string, timeout, interval time.Duration) {
	By("Creating the InstanceSnapshot resource")
	Expect(k8sClient.Create(ctx, isnap)).Should(Succeed())

	By("Checking that the instance has been created")
	instanceSnapshotLookupKey := types.NamespacedName{Name: isnap.Name, Namespace: workingNamespace}
	createdInstanceSnapshot := &crownlabsv1alpha2.InstanceSnapshot{}

	doesEventuallyExists(ctx, instanceSnapshotLookupKey, createdInstanceSnapshot, BeTrue(), timeout, interval)

	By("Checking that the InstanceSnapshot failed")
	checkIsnapStatus(ctx, isnap.Name, workingNamespace, crownlabsv1alpha2.Failed, timeout, interval)
}
