package instctrl_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	controlplanekamajiv1 "github.com/clastix/cluster-api-control-plane-provider-kamaji/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instctrl"

	infrav1 "sigs.k8s.io/cluster-api-provider-kubevirt/api/v1alpha1"
	capiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	addonsv1beta1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
)

type stubTransport struct{ orig http.RoundTripper }

func (s stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "raw.githubusercontent.com" &&
		strings.HasPrefix(req.URL.Path, "/projectcalico/calico/") &&
		strings.HasSuffix(req.URL.Path, "/manifests/calico.yaml") {
		body := io.NopCloser(strings.NewReader(`apiVersion: v1
kind: Namespace
metadata:
  name: calico-system
`))
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       body,
			Request:    req,
		}, nil
	}
	return s.orig.RoundTrip(req)
}

var _ = Describe("Cluster Environment Management", func() {
	var (
		ctx           context.Context
		clientBuilder fake.ClientBuilder
		reconciler    instctrl.InstanceReconciler
		testScheme    *runtime.Scheme

		instance    clv1alpha2.Instance
		template    clv1alpha2.Template
		tenant      clv1alpha2.Tenant
		environment clv1alpha2.Environment

		err           error
		origTransport http.RoundTripper
	)

	const (
		instanceName      = "kubernetes-0000"
		instanceNamespace = "tenant-tester"
		templateName      = "kubernetes"
		templateNamespace = "workspace-netgroup"
		workspaceName     = "netgroup"
		environmentName   = "control-plane"
		tenantName        = "tester"
		clusterName       = "test-cluster"
	)

	BeforeEach(func() {
		ctx = ctrl.LoggerInto(context.Background(), logr.Discard())

		origTransport = http.DefaultTransport
		http.DefaultTransport = stubTransport{orig: origTransport}

		testScheme = runtime.NewScheme()
		Expect(scheme.AddToScheme(testScheme)).To(Succeed())
		Expect(clv1alpha2.AddToScheme(testScheme)).To(Succeed())
		Expect(capiv1.AddToScheme(testScheme)).To(Succeed())
		Expect(infrav1.AddToScheme(testScheme)).To(Succeed())
		Expect(controlplanekamajiv1.AddToScheme(testScheme)).To(Succeed())
		Expect(bootstrapv1.AddToScheme(testScheme)).To(Succeed())
		Expect(addonsv1beta1.AddToScheme(testScheme)).To(Succeed())
		Expect(appsv1.AddToScheme(testScheme)).To(Succeed())
		Expect(corev1.AddToScheme(testScheme)).To(Succeed())
		Expect(rbacv1.AddToScheme(testScheme)).To(Succeed())

		clientBuilder = *fake.NewClientBuilder().WithScheme(testScheme)

		instance = clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instanceName,
				Namespace: instanceNamespace,
				UID:       "test-uid-12345",
			},
			Spec: clv1alpha2.InstanceSpec{
				Running:  true,
				Template: clv1alpha2.GenericRef{Name: templateName, Namespace: templateNamespace},
				Tenant:   clv1alpha2.GenericRef{Name: tenantName},
			},
		}

		template = clv1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: templateName, Namespace: templateNamespace},
			Spec: clv1alpha2.TemplateSpec{
				WorkspaceRef: clv1alpha2.GenericRef{Name: workspaceName},
			},
		}

		tenant = clv1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{Name: tenantName},
			Spec:       clv1alpha2.TenantSpec{PublicKeys: []string{"ssh-rsa key1", "ssh-rsa key2"}},
		}

		envJSON := fmt.Sprintf(`{
		  "name": %q,
		  "mountMyDriveVolume": true,
		  "mode": "Standard",
		  "cluster": {
		    "name": %q,
		    "controlPlane": { "replicas": 1 },
		    "machineDeploy": { "replicas": 2 }
		  }
		}`, environmentName, clusterName)
		Expect(json.Unmarshal([]byte(envJSON), &environment)).To(Succeed())
	})

	AfterEach(func() {
		http.DefaultTransport = origTransport
	})

	JustBeforeEach(func() {
		client := clientBuilder.Build()
		reconciler = instctrl.InstanceReconciler{
			Client: client,
			Scheme: testScheme,
			ServiceUrls: instctrl.ServiceUrls{
				WebsiteBaseURL: "https://crownlabs.example.com",
			},
		}

		ctx, _ = clctx.InstanceInto(ctx, &instance)
		ctx, _ = clctx.TemplateInto(ctx, &template)
		ctx, _ = clctx.TenantInto(ctx, &tenant)
		ctx, _ = clctx.EnvironmentInto(ctx, &environment)
	})

	Describe("The EnforceClusterEnvironment function", func() {
		JustBeforeEach(func() {
			err = reconciler.EnforceClusterEnvironment(ctx)
		})

		When("creating a complete cluster environment", func() {
			It("Should succeed", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("Should create the cluster resource", func() {
				var cluster capiv1.Cluster
				clusterKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-cluster", clusterName),
					Namespace: instanceNamespace,
				}
				Expect(reconciler.Get(ctx, clusterKey, &cluster)).To(Succeed())
				Expect(cluster.GetLabels()).To(HaveKeyWithValue("crownlabs.polito.it/instance", instanceName))
				Expect(cluster.GetOwnerReferences()).To(HaveLen(1))
				Expect(cluster.GetOwnerReferences()[0].Name).To(Equal(instance.Name))
			})

			It("Should create the Kamaji infrastructure", func() {
				var infra infrav1.KubevirtCluster
				infraKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-infra", clusterName),
					Namespace: instanceNamespace,
				}
				Expect(reconciler.Get(ctx, infraKey, &infra)).To(Succeed())
				Expect(infra.GetAnnotations()).To(HaveKeyWithValue("cluster.x-k8s.io/managed-by", "kamaji"))
				Expect(infra.GetLabels()).To(HaveKeyWithValue("crownlabs.polito.it/instance", instanceName))
			})

			It("Should create the Kamaji control plane", func() {
				var cp controlplanekamajiv1.KamajiControlPlane
				cpKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-control-plane", clusterName),
					Namespace: instanceNamespace,
				}
				Expect(reconciler.Get(ctx, cpKey, &cp)).To(Succeed())
				Expect(cp.GetLabels()).To(HaveKeyWithValue("crownlabs.polito.it/instance", instanceName))
				Expect(cp.Spec.Replicas).ToNot(BeNil())
			})

			It("Should create the machine deployment", func() {
				var md capiv1.MachineDeployment
				mdKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-md", clusterName),
					Namespace: instanceNamespace,
				}
				Expect(reconciler.Get(ctx, mdKey, &md)).To(Succeed())
				Expect(md.Spec.ClusterName).To(Equal(fmt.Sprintf("%s-cluster", clusterName)))
				Expect(md.GetLabels()).To(HaveKeyWithValue("crownlabs.polito.it/instance", instanceName))
				Expect(md.Spec.Replicas).ToNot(BeNil())
			})

			It("Should create the KubeVirt machine template", func() {
				var kvmt infrav1.KubevirtMachineTemplate
				kvmtKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-md-worker", clusterName),
					Namespace: instanceNamespace,
				}
				Expect(reconciler.Get(ctx, kvmtKey, &kvmt)).To(Succeed())
				Expect(kvmt.Spec.Template.Spec.BootstrapCheckSpec.CheckStrategy).To(Equal("ssh"))
				Expect(kvmt.Labels).To(HaveKey(capiv1.ClusterNameLabel))
			})

			It("Should create the bootstrap configuration", func() {
				var bt bootstrapv1.KubeadmConfigTemplate
				btKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-md-bootstrap", clusterName),
					Namespace: instanceNamespace,
				}
				Expect(reconciler.Get(ctx, btKey, &bt)).To(Succeed())
				Expect(bt.Spec.Template.Spec.JoinConfiguration).ToNot(BeNil())
				Expect(bt.GetLabels()).To(HaveKeyWithValue("crownlabs.polito.it/instance", instanceName))
			})

			It("Should create the Calico CNI resources", func() {
				// ConfigMap
				var cm corev1.ConfigMap
				cmKey := types.NamespacedName{
					Name:      "calico-manifest",
					Namespace: instanceNamespace,
				}
				Expect(reconciler.Get(ctx, cmKey, &cm)).To(Succeed())
				Expect(cm.Data).To(HaveKey("calico.yaml"))
				Expect(cm.Labels).To(HaveKeyWithValue("kamaji.clastix.io/tenant", tenantName))

				// ClusterResourceSet
				var crs addonsv1beta1.ClusterResourceSet
				crsKey := types.NamespacedName{
					Name:      "install-calico",
					Namespace: instanceNamespace,
				}
				Expect(reconciler.Get(ctx, crsKey, &crs)).To(Succeed())
				Expect(crs.Spec.Strategy).To(Equal(string(addonsv1beta1.ClusterResourceSetStrategyApplyOnce)))
				Expect(crs.Spec.ClusterSelector.MatchLabels).To(HaveKeyWithValue("crownlabs.polito.it/tenant", tenantName))
			})

			It("Should create the GUI deployment", func() {
				var dep appsv1.Deployment
				depKey := types.NamespacedName{
					Name:      "capi-visualizer",
					Namespace: instanceNamespace,
				}
				Expect(reconciler.Get(ctx, depKey, &dep)).To(Succeed())
				Expect(dep.Labels).To(HaveKeyWithValue("app", "capi-visualizer"))
				Expect(dep.GetOwnerReferences()).To(HaveLen(1))
			})

			It("Should create the GUI service and service account", func() {
				var svc corev1.Service
				svcKey := types.NamespacedName{
					Name:      "capi-visualizer",
					Namespace: instanceNamespace,
				}
				Expect(reconciler.Get(ctx, svcKey, &svc)).To(Succeed())
				Expect(svc.Labels).To(HaveKeyWithValue("app", "capi-visualizer"))

				var sa corev1.ServiceAccount
				saKey := types.NamespacedName{
					Name:      "capi-visualizer",
					Namespace: instanceNamespace,
				}
				Expect(reconciler.Get(ctx, saKey, &sa)).To(Succeed())
				Expect(sa.GetOwnerReferences()).To(HaveLen(1))
			})

			It("Should create the RBAC resources", func() {
				var cr rbacv1.ClusterRole
				crKey := types.NamespacedName{
					Name: "capi-visualizer" + instanceNamespace,
				}
				Expect(reconciler.Get(ctx, crKey, &cr)).To(Succeed())
				Expect(cr.Rules).ToNot(BeEmpty())

				var crb rbacv1.ClusterRoleBinding
				crbKey := types.NamespacedName{
					Name: "capi-visualizer" + instanceNamespace,
				}
				Expect(reconciler.Get(ctx, crbKey, &crb)).To(Succeed())
				Expect(crb.RoleRef.Name).To(Equal("capi-visualizer"))
				Expect(crb.Subjects).To(HaveLen(1))
				Expect(crb.Subjects[0].Namespace).To(Equal(instanceNamespace))
			})
		})

		When("the resources already exist", func() {
			BeforeEach(func() {
				existingCluster := capiv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-cluster", clusterName),
						Namespace: instanceNamespace,
					},
				}
				existingDep := appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capi-visualizer",
						Namespace: instanceNamespace,
					},
				}
				clientBuilder.WithObjects(&existingCluster, &existingDep)
			})

			It("Should succeed", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("Should handle existing resources gracefully", func() {
				var cluster capiv1.Cluster
				clusterKey := types.NamespacedName{
					Name:      fmt.Sprintf("%s-cluster", clusterName),
					Namespace: instanceNamespace,
				}
				Expect(reconciler.Get(ctx, clusterKey, &cluster)).To(Succeed())
				Expect(cluster.GetLabels()).To(HaveKeyWithValue("crownlabs.polito.it/instance", instanceName))
			})
		})
	})

	Describe("Idempotency", func() {
		It("Should handle multiple calls without errors", func() {
			// First call
			err := reconciler.EnforceClusterEnvironment(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Second call
			err = reconciler.EnforceClusterEnvironment(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Verify resources still exist
			var cluster capiv1.Cluster
			clusterKey := types.NamespacedName{
				Name:      fmt.Sprintf("%s-cluster", clusterName),
				Namespace: instanceNamespace,
			}
			Expect(reconciler.Get(ctx, clusterKey, &cluster)).To(Succeed())
		})
	})

	Describe("Owner References", func() {
		It("Should set owner references on all resources", func() {
			err := reconciler.EnforceClusterEnvironment(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Sample check for owner references
			var cm corev1.ConfigMap
			cmKey := types.NamespacedName{
				Name:      "calico-manifest",
				Namespace: instanceNamespace,
			}
			Expect(reconciler.Get(ctx, cmKey, &cm)).To(Succeed())

			ownerRefs := cm.GetOwnerReferences()
			Expect(ownerRefs).To(HaveLen(1))
			Expect(ownerRefs[0].Name).To(Equal(instance.Name))
			Expect(ownerRefs[0].Kind).To(Equal("Instance"))
			Expect(ownerRefs[0].Controller).To(Equal(ptr.To(true)))
			Expect(ownerRefs[0].BlockOwnerDeletion).To(Equal(ptr.To(true)))
		})
	})
})
