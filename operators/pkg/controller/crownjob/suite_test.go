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

// Package crownjob_test implements the crownjob controllers tests.
package crownjob_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2/textlogger"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	apicommon "github.com/netgroup-polito/CrownLabs/operators/api/common"
	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	ctrlcommon "github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/crownjob"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/tests"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const (
	tenantNamespace     = "tenant-student"
	otherNamespace      = "tenant-other"
	untargetedNamespace = "tenant-untargeted"
	workspaceNamespace  = "workspace-ws"
	workspaceName       = "ws"
	noQueueWorkspace    = "ws-noqueue"
	tenantName          = "student"
	envName             = "work"
	image               = "harbor.example.com/standalone:v1"
	gpu                 = "nvidia.com/gpu"
	fpga                = "example.com/fpga"
	highPriorityClass   = "p-high"
	logHelperImage      = "registry.k8s.io/kubectl:v1.34.1"
	finishedJobTTL      = 10 * time.Minute
)

var (
	ctx                context.Context
	k8sClient          client.Client
	recorder           *record.FakeRecorder
	crownJobReconciler crownjob.Reconciler
	queueReconciler    crownjob.QueueReconciler
	instanceReconciler crownjob.InstanceReconciler
	testEnv            *envtest.Environment
	stopManager        context.CancelFunc

	targetLabel = ctrlcommon.NewLabel("crownlabs.polito.it/operator-selector", "test")
)

func TestCrownJob(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CrownJob Controller Suite")
}

var _ = BeforeSuite(func() {
	tests.LogsToGinkgoWriter()

	// The Kueue CRDs are in a subdirectory, which neither the other suites nor "kubectl apply -f tests/crds" read.
	testEnv = tests.ForgeEnvtestEnv(true)
	testEnv.CRDDirectoryPaths = append(testEnv.CRDDirectoryPaths, filepath.Join("..", "..", "..", "tests", "crds", "kueue"))

	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	Expect(clv1alpha2.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(clv1alpha1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())

	ctrl.SetLogger(textlogger.NewLogger(textlogger.NewConfig()))
	ctx = ctrl.LoggerInto(context.Background(), ctrl.Log)

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	recorder = record.NewFakeRecorder(1024)
	crownJobReconciler = crownjob.Reconciler{
		Client:             k8sClient,
		APIReader:          k8sClient,
		TargetLabel:        targetLabel,
		EventsRecorder:     recorder,
		RuntimeClasses:     map[string]string{gpu: "nvidia", fpga: "fpga"},
		DefaultMaxRunTime:  2 * time.Hour,
		LogHelperImage:     logHelperImage,
		FinishedJobTTL:     finishedJobTTL,
		ReconcileDeferHook: GinkgoRecover,
	}
	queueReconciler = crownjob.QueueReconciler{
		Client:             k8sClient,
		TargetLabel:        targetLabel,
		ReconcileDeferHook: GinkgoRecover,
	}

	// The instance reconciler lists the CrownJobs through a field index, which only a cache serves: a manager is
	// started for its cache only (no controller is registered), and Reconcile is called by the tests.
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme.Scheme, Metrics: metricsserver.Options{BindAddress: "0"}})
	Expect(err).ToNot(HaveOccurred())
	Expect(crownjob.IndexCrownJobsByInstance(ctx, mgr.GetFieldIndexer())).To(Succeed())
	var managerCtx context.Context
	managerCtx, stopManager = context.WithCancel(ctx)
	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(managerCtx)).To(Succeed())
	}()
	Expect(mgr.GetCache().WaitForCacheSync(ctx)).To(BeTrue())
	instanceReconciler = crownjob.InstanceReconciler{
		Client:             mgr.GetClient(),
		APIReader:          k8sClient,
		TargetLabel:        targetLabel,
		EventsRecorder:     recorder,
		ReconcileDeferHook: GinkgoRecover,
	}

	createFixtures()
})

var _ = AfterSuite(func() {
	stopManager()
	Expect(testEnv.Stop()).To(Succeed())
})

// createFixtures creates the objects shared by all the tests: namespaces, workspaces, templates, tenants, instances
// and their deployments.
func createFixtures() {
	targeted := map[string]string{targetLabel.GetKey(): targetLabel.GetValue(), forge.LabelTypeKey: "tenant"}
	create(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: tenantNamespace, Labels: targeted}})
	create(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: otherNamespace, Labels: targeted}})
	create(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: untargetedNamespace, Labels: map[string]string{forge.LabelTypeKey: "tenant"}}})
	create(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: workspaceNamespace}})

	create(workspace(workspaceName, &clv1alpha1.QueueLimits{
		MaxAtOnce: map[string]resource.Quantity{
			"cpu": resource.MustParse("4"), "memory": resource.MustParse("8Gi"),
			gpu: resource.MustParse("1"), fpga: resource.MustParse("1"),
		},
		MaxRunTime: "6h",
	}))
	create(workspace(noQueueWorkspace, nil))
	// The ClusterQueue of the workspace, as the queue controller creates it.
	_, err := queueReconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKey{Name: workspaceName}})
	Expect(err).ToNot(HaveOccurred())

	wpc := &unstructured.Unstructured{}
	wpc.SetGroupVersionKind(forge.KueueWorkloadPriorityClassGVK)
	wpc.SetName(highPriorityClass)
	wpc.Object["value"] = int64(1000)
	create(wpc)

	create(tenant(tenantName, clv1alpha2.TenantWorkspaceEntry{Name: workspaceName, Role: clv1alpha2.User},
		clv1alpha2.TenantWorkspaceEntry{Name: noQueueWorkspace, Role: clv1alpha2.Manager}))
	create(tenant("candidate", clv1alpha2.TenantWorkspaceEntry{Name: workspaceName, Role: clv1alpha2.Candidate}))
	create(tenant("prio", clv1alpha2.TenantWorkspaceEntry{Name: workspaceName, Role: clv1alpha2.User, JobPriorityClass: highPriorityClass}))
	create(tenant("prio-missing", clv1alpha2.TenantWorkspaceEntry{Name: workspaceName, Role: clv1alpha2.User, JobPriorityClass: "ghost"}))

	editable := &clv1alpha2.TemplateJob{EnvironmentName: envName, CommandPolicy: clv1alpha2.JobCommandEditable, MaxRunTime: "1h"}
	createTemplate("tpl", workspaceName, editable, 1, 50)
	createTemplate("tpl-nojob", workspaceName, nil, 1, 50)
	createTemplate("tpl-fixed", workspaceName, &clv1alpha2.TemplateJob{
		EnvironmentName: envName, CommandPolicy: clv1alpha2.JobCommandFixed, Command: []string{"./run.sh"}}, 1, 50)
	createTemplate("tpl-imagedefault", workspaceName, &clv1alpha2.TemplateJob{
		EnvironmentName: envName, CommandPolicy: clv1alpha2.JobCommandImageDefault}, 1, 50)
	createTemplate("tpl-gpu", workspaceName, &clv1alpha2.TemplateJob{
		EnvironmentName: envName, OtherResources: map[string]resource.Quantity{gpu: resource.MustParse("1")}}, 1, 50)
	createTemplate("tpl-amd", workspaceName, &clv1alpha2.TemplateJob{
		EnvironmentName: envName, OtherResources: map[string]resource.Quantity{"amd.com/gpu": resource.MustParse("1")}}, 1, 50)
	createTemplate("tpl-conflict", workspaceName, &clv1alpha2.TemplateJob{
		EnvironmentName: envName, OtherResources: map[string]resource.Quantity{gpu: resource.MustParse("1"), fpga: resource.MustParse("1")}}, 1, 50)
	createTemplate("tpl-bigcpu", workspaceName, editable, 8, 100)
	createTemplate("tpl-noqueue", noQueueWorkspace, editable, 1, 50)
	createTemplate("tpl-nows", "", editable, 1, 50)
	createTemplate("tpl-ghostws", "ghost", editable, 1, 50)

	for name, tmpl := range map[string]string{
		"inst": "tpl", "inst-nojob": "tpl-nojob", "inst-fixed": "tpl-fixed", "inst-imagedefault": "tpl-imagedefault",
		"inst-gpu": "tpl-gpu", "inst-amd": "tpl-amd", "inst-conflict": "tpl-conflict", "inst-bigcpu": "tpl-bigcpu",
		"inst-noqueue": "tpl-noqueue", "inst-nows": "tpl-nows", "inst-ghostws": "tpl-ghostws", "inst-ghosttpl": "ghost",
	} {
		createInstance(tenantNamespace, name, tmpl, tenantName, true)
	}
	createInstance(tenantNamespace, "inst-candidate", "tpl", "candidate", true)
	createInstance(tenantNamespace, "inst-prio", "tpl", "prio", true)
	createInstance(tenantNamespace, "inst-prio-missing", "tpl", "prio-missing", true)
	createInstance(tenantNamespace, "inst-ghosttenant", "tpl", "ghost", true)
	createInstance(otherNamespace, "inst", "tpl", tenantName, true)
	createInstance(untargetedNamespace, "inst", "tpl", tenantName, true)
}

func create(obj client.Object) {
	ExpectWithOffset(1, k8sClient.Create(ctx, obj)).To(Succeed())
}

func workspace(name string, limits *clv1alpha1.QueueLimits) *clv1alpha1.Workspace {
	return &clv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: map[string]string{targetLabel.GetKey(): targetLabel.GetValue()}},
		Spec: clv1alpha1.WorkspaceSpec{
			PrettyName: name,
			Quota: apicommon.WorkspaceResourceQuota{
				ResourceSpec: apicommon.ResourceSpec{CPU: 10, Memory: resource.MustParse("10Gi")},
				Instances:    5,
			},
			QueueLimits: limits,
		},
	}
}

func tenant(name string, entries ...clv1alpha2.TenantWorkspaceEntry) *clv1alpha2.Tenant {
	return &clv1alpha2.Tenant{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: map[string]string{targetLabel.GetKey(): targetLabel.GetValue()}},
		Spec: clv1alpha2.TenantSpec{
			FirstName: "Test", LastName: "Tenant", Email: name + "@example.com",
			Workspaces: entries,
		},
	}
}

func jobEnvironment(cpu int64, reservedCPUPercentage uint32) clv1alpha2.Environment {
	return clv1alpha2.Environment{
		Name:            envName,
		Image:           image,
		EnvironmentType: clv1alpha2.ClassStandalone,
		Persistent:      true,
		Resources: clv1alpha2.EnvironmentResources{
			ResourceSpec:          apicommon.ResourceSpec{CPU: cpu, Memory: resource.MustParse("1Gi"), Disk: resource.MustParse("1Gi")},
			ReservedCPUPercentage: reservedCPUPercentage,
		},
	}
}

func createTemplate(name, ws string, job *clv1alpha2.TemplateJob, cpu int64, reservedCPUPercentage uint32) {
	create(&clv1alpha2.Template{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: workspaceNamespace},
		Spec: clv1alpha2.TemplateSpec{
			PrettyName:      name,
			Description:     name,
			WorkspaceRef:    clv1alpha2.GenericRef{Name: ws},
			EnvironmentList: []clv1alpha2.Environment{jobEnvironment(cpu, reservedCPUPercentage)},
			Cleanup:         clv1alpha2.CleanupOptions{DeleteAfterCreation: "never", StopAfterInactivity: "never", DeleteAfterInactivity: "never"},
			Job:             job,
		},
	})
}

func instance(namespace, name, tmpl, tenantRef string) *clv1alpha2.Instance {
	return &clv1alpha2.Instance{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: clv1alpha2.InstanceSpec{
			Template: clv1alpha2.GenericRef{Name: tmpl, Namespace: workspaceNamespace},
			Tenant:   clv1alpha2.GenericRef{Name: tenantRef},
			Running:  true,
		},
	}
}

// createInstance creates an Instance and, if requested, its Deployment and its ReadWriteMany disk, forged as the
// instance controller does.
func createInstance(namespace, name, tmpl, tenantRef string, withDeployment bool) {
	inst := instance(namespace, name, tmpl, tenantRef)
	create(inst)
	if withDeployment {
		createDeployment(inst, tmpl)
		createDisk(inst, corev1.ReadWriteMany)
	}
}

// createDisk creates the PVC of the job environment of the given Instance, with the given access mode.
func createDisk(inst *clv1alpha2.Instance, accessMode corev1.PersistentVolumeAccessMode) {
	environment := jobEnvironment(1, 50)
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: forge.ObjectMetaWithSuffix(inst, envName),
		Spec:       forge.PVCSpec(accessMode, nil, &environment.Resources.Disk),
	}
	ExpectWithOffset(2, k8sClient.Create(ctx, &pvc)).To(Succeed())
}

func createDeployment(inst *clv1alpha2.Instance, tmpl string) {
	template := clv1alpha2.Template{ObjectMeta: metav1.ObjectMeta{Name: tmpl, Namespace: workspaceNamespace},
		Spec: clv1alpha2.TemplateSpec{EnvironmentList: []clv1alpha2.Environment{jobEnvironment(1, 50)}}}
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(&template), &template); err != nil {
		ExpectWithOffset(2, kerrors.IsNotFound(err)).To(BeTrue())
	}
	environment := template.Spec.EnvironmentList[0]
	deployment := appsv1.Deployment{
		ObjectMeta: forge.ObjectMetaWithSuffix(inst, environment.Name),
		Spec: forge.DeploymentSpec(inst, &template, &environment, nil,
			&forge.ContainerEnvOpts{ImagesTag: "v1", ContentToolsImg: "tools"}),
	}
	ExpectWithOffset(2, k8sClient.Create(ctx, &deployment)).To(Succeed())
}

// stopInstanceContainer simulates stopping the instance: the Deployment scaled to 0, and its pod gone.
func stopInstanceContainer(inst *clv1alpha2.Instance) {
	var deployment appsv1.Deployment
	key := forge.NamespacedNameWithSuffix(inst, envName)
	ExpectWithOffset(1, k8sClient.Get(ctx, key, &deployment)).To(Succeed())
	deployment.Spec.Replicas = ptr.To[int32](0)
	ExpectWithOffset(1, k8sClient.Update(ctx, &deployment)).To(Succeed())
	deployment.Status.Replicas = 0
	ExpectWithOffset(1, k8sClient.Status().Update(ctx, &deployment)).To(Succeed())
}

// getInstance returns the current Instance with the given name, in the tenant namespace.
func getInstance(name string) (*clv1alpha2.Instance, error) {
	var inst clv1alpha2.Instance
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: tenantNamespace, Name: name}, &inst)
	return &inst, err
}
