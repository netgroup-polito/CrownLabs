package instautoctrl_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegaTypes "github.com/onsi/gomega/types"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/instautoctrl"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instautoctrl/mocks"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instctrl"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/tests"
	virtv1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var ctx context.Context
var cancel context.CancelFunc
var cfg *rest.Config
var k8sClient client.Client
var k8sClientExpiration client.Client
var testEnv *envtest.Environment
var instanceInactiveTerminationReconciler instautoctrl.InstanceInactiveTerminationReconciler
var mockCtrl *gomock.Controller
var mockProm *mocks.MockPrometheusClientInterface
var instanceReconciler instctrl.InstanceReconciler

// var log logr.Logger

func TestInstautoctrl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Instautoctrl Suite")
}

// var testLogger = zap.New(zap.UseDevMode(true)).WithName("test")
var _ = BeforeSuite(func() {
	ctx, cancel = context.WithCancel(context.Background())
	tests.LogsToGinkgoWriter()
	// opts := zap.Options{
	// 	Development: true,
	// }
	// opts.BindFlags(flag.CommandLine)
	// log = zap.New(zap.UseFlagOptions(&opts))
	// ctrl.SetLogger(log)

	By("bootstrapping test environment")

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "deploy", "crds"),
			filepath.Join("..", "..", "..", "tests", "crds")},
		ErrorIfCRDPathMissing: true,
	}
	mockCtrl = gomock.NewController(GinkgoT())
	mockProm = mocks.NewMockPrometheusClientInterface(mockCtrl)
	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	Expect(crownlabsv1alpha2.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(virtv1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(cdiv1beta1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

func doesEventuallyExists(ctx context.Context, objLookupKey types.NamespacedName, targetObj client.Object, expectedStatus gomegaTypes.GomegaMatcher, timeout, interval time.Duration, k8sClient client.Client) {
	Eventually(func() bool {
		err := k8sClient.Get(ctx, objLookupKey, targetObj)
		return err == nil
	}, timeout, interval).Should(expectedStatus)
}
