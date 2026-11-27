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

package instctrl

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"reflect"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/clcontext"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

func testCACertificate(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "kubernetes"},
		NotBefore: time.Now(), NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

const testKubeconfig = `apiVersion: v1
kind: Config
clusters:
- name: wl
  cluster:
    server: https://10.0.0.1:6443
    insecure-skip-tls-verify: true
users:
- name: admin
  user:
    token: secret
contexts:
- name: wl
  context: {cluster: wl, user: admin}
current-context: wl
`

func TestEnforceClusterAPIServerExposure(t *testing.T) {
	const ns = "tenant-alice"
	instance := &clv1alpha2.Instance{
		ObjectMeta: metav1.ObjectMeta{Name: "inst", Namespace: ns},
		Spec:       clv1alpha2.InstanceSpec{Tenant: clv1alpha2.GenericRef{Name: "alice"}},
		Status:     clv1alpha2.InstanceStatus{Environments: []clv1alpha2.InstanceStatusEnv{{Name: "k8s"}}},
	}
	template := &clv1alpha2.Template{Spec: clv1alpha2.TemplateSpec{WorkspaceRef: clv1alpha2.GenericRef{Name: "netlab"}}}
	env := &clv1alpha2.Environment{Name: "k8s", EnvironmentType: clv1alpha2.ClassCluster}
	names := forge.ClusterNamesFromInstance(instance, env)
	expo := forge.ExpositionConfig{GatewayAPIMode: true, GatewayName: "gw", GatewayNamespace: "gw-ns", WebsiteBaseURL: "crownlabs.example.com"}
	oidc := forge.ClusterAPIServerOpts{OIDCIssuerURL: "https://auth.example.com", OIDCClientID: "k8s", PathPrefix: "/clusters"}
	path := "/clusters/" + ns + "/" + names.Cluster

	caSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: names.ControlPlane + "-ca", Namespace: ns},
		Data:       map[string][]byte{"ca.crt": testCACertificate(t)},
	}
	kubeconfigSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: names.Cluster + "-kubeconfig", Namespace: ns},
		Data:       map[string][]byte{"value": []byte(testKubeconfig)},
	}
	acceptedRoute := &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Name: forge.ClusterAPIServerExposureName(names.Cluster), Namespace: ns},
		Spec:       forge.ClusterAPIServerHTTPRouteSpec(path, names.ControlPlane, &expo),
		Status: gatewayv1.HTTPRouteStatus{RouteStatus: gatewayv1.RouteStatus{Parents: []gatewayv1.RouteParentStatus{{
			ParentRef:  forge.ParentReference("gw", "gw-ns"),
			Conditions: []metav1.Condition{{Type: string(gatewayv1.RouteConditionAccepted), Status: metav1.ConditionTrue}},
		}}}},
	}
	initialized := map[string]interface{}{"initialization": map[string]interface{}{"controlPlaneInitialized": true}}

	tests := []struct {
		name         string
		expo         forge.ExpositionConfig
		oidc         forge.ClusterAPIServerOpts
		objects      []client.Object
		status       map[string]interface{}
		wantErr      bool
		wantRoute    bool
		wantAdmins   bool
		wantURL      string
		wantAccepted bool
	}{
		{name: "disabled without OIDC", expo: expo, objects: []client.Object{caSecret}},
		{name: "disabled without gateway API", expo: forge.ExpositionConfig{}, oidc: oidc, objects: []client.Object{caSecret}},
		{name: "waiting for the CA", expo: expo, oidc: oidc},
		{name: "invalid CA", expo: expo, oidc: oidc, wantErr: true, objects: []client.Object{&corev1.Secret{
			ObjectMeta: caSecret.ObjectMeta, Data: map[string][]byte{"ca.crt": []byte("garbage")},
		}}},
		{name: "control plane not initialized", expo: expo, oidc: oidc, objects: []client.Object{caSecret, kubeconfigSecret}, wantRoute: true},
		{name: "route not accepted", expo: expo, oidc: oidc, status: initialized,
			objects: []client.Object{caSecret, kubeconfigSecret}, wantRoute: true, wantAdmins: true},
		{name: "exposed", expo: expo, oidc: oidc, status: initialized,
			objects:   []client.Object{caSecret, kubeconfigSecret, acceptedRoute.DeepCopy()},
			wantRoute: true, wantAdmins: true, wantURL: "https://crownlabs.example.com" + path, wantAccepted: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sch := runtime.NewScheme()
			for _, add := range []func(*runtime.Scheme) error{scheme.AddToScheme, clv1alpha2.AddToScheme, gatewayv1.Install} {
				if err := add(sch); err != nil {
					t.Fatal(err)
				}
			}
			c := fake.NewClientBuilder().WithScheme(sch).WithObjects(tt.objects...).Build()
			wl := fake.NewClientBuilder().WithScheme(sch).Build()
			var wlHost string

			inst := instance.DeepCopy()
			inst.Status.Environments[0].APIServerURL = "stale"
			ctx, _ := clctx.InstanceInto(context.Background(), inst)
			ctx, _ = clctx.TemplateInto(ctx, template)
			ctx, _ = clctx.EnvironmentInto(ctx, env)
			ctx = clctx.EnvironmentIndexInto(ctx, 0)
			r := &InstanceReconciler{
				Client: c, Scheme: sch, ExpositionConfig: tt.expo, ClusterAPIServerOpts: tt.oidc,
				WorkloadClusterClientBuilder: func(cfg *rest.Config) (client.Client, error) {
					wlHost = cfg.Host
					return wl, nil
				},
			}
			cluster := &unstructured.Unstructured{Object: map[string]interface{}{"status": tt.status}}

			err := r.enforceClusterAPIServerExposure(ctx, &names, cluster)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}

			key := client.ObjectKey{Namespace: ns, Name: forge.ClusterAPIServerExposureName(names.Cluster)}
			for _, obj := range []client.Object{&gatewayv1.HTTPRoute{}, &gatewayv1.BackendTLSPolicy{}, &corev1.ConfigMap{}} {
				err := c.Get(ctx, key, obj)
				if tt.wantRoute && err != nil {
					t.Fatalf("expected %T to be created: %v", obj, err)
				}
				if !tt.wantRoute && !kerrors.IsNotFound(err) {
					t.Fatalf("expected %T not to be created, got %v", obj, err)
				}
			}
			if tt.wantRoute {
				var policy gatewayv1.BackendTLSPolicy
				_ = c.Get(ctx, key, &policy)
				if got, want := string(policy.Spec.Validation.Hostname), names.ControlPlane+"."+ns+".svc"; got != want {
					t.Fatalf("backend TLS hostname: got %q, want %q", got, want)
				}
				var route gatewayv1.HTTPRoute
				_ = c.Get(ctx, key, &route)
				if got := *route.Spec.Rules[0].Matches[0].Path.Value; got != path {
					t.Fatalf("route path: got %q, want %q", got, path)
				}
			}

			var crb rbacv1.ClusterRoleBinding
			err = wl.Get(ctx, client.ObjectKey{Name: forge.ClusterAdminsClusterRoleBindingName}, &crb)
			if !tt.wantAdmins {
				if !kerrors.IsNotFound(err) {
					t.Fatalf("expected no cluster admins binding, got %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected cluster admins binding: %v", err)
				}
				if want := "https://" + names.ControlPlane + "." + ns + ".svc:6443"; wlHost != want {
					t.Fatalf("workload cluster host: got %q, want %q", wlHost, want)
				}
				wantSubjects := []rbacv1.Subject{
					{Kind: rbacv1.UserKind, APIGroup: rbacv1.GroupName, Name: "alice"},
					{Kind: rbacv1.GroupKind, APIGroup: rbacv1.GroupName, Name: "kubernetes:workspace-netlab:manager"},
				}
				if !reflect.DeepEqual(crb.Subjects, wantSubjects) || crb.RoleRef.Name != "cluster-admin" {
					t.Fatalf("unexpected binding: %+v", crb)
				}
			}

			status := inst.Status.Environments[0]
			if status.APIServerURL != tt.wantURL || status.ExpositionAccepted != tt.wantAccepted {
				t.Fatalf("status: got (%q, %t), want (%q, %t)", status.APIServerURL, status.ExpositionAccepted, tt.wantURL, tt.wantAccepted)
			}
		})
	}
}
