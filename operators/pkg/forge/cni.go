package forge

import (
	"context"
	"io"
	"net/http"

	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	addonsv1beta1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeployCNI(ctx context.Context, k8sClient client.Client) error {
	instance := clctx.InstanceFrom(ctx)
	log := ctrl.LoggerFrom(ctx)
	ns := instance.Namespace
	tenant := instance.Spec.Tenant.Name
	resp, err := http.Get("https://raw.githubusercontent.com/projectcalico/calico/v3.28.0/manifests/calico.yaml")
	if err != nil {
		log.Error(err, "download Calico YAM")
		return err
	}
	defer resp.Body.Close()
	yamlBytes, _ := io.ReadAll(resp.Body)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "calico-manifest",
			Namespace: ns,
			Labels: map[string]string{
				"kamaji.clastix.io/tenant": tenant,
			},
		},
		Data: map[string]string{
			"calico.yaml": string(yamlBytes),
		},
	}

	if err := k8sClient.Patch(ctx, cm, client.Apply,
		client.ForceOwnership,
		client.FieldOwner("cni-bootstrap")); err != nil {
		log.Error(err, "apply ConfigMap")
		return err
	}
	crs := &addonsv1beta1.ClusterResourceSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "install-calico",
			Namespace: ns,
		},
		Spec: addonsv1beta1.ClusterResourceSetSpec{
			Strategy: string(addonsv1beta1.ClusterResourceSetStrategyApplyOnce),
			ClusterSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kamaji.clastix.io/tenant": tenant,
				},
			},
			Resources: []addonsv1beta1.ResourceRef{{
				Name: "calico-manifest",
				Kind: "ConfigMap",
			}},
		},
	}
	if err := k8sClient.Patch(ctx, crs, client.Apply,
		client.ForceOwnership,
		client.FieldOwner("cni-bootstrap")); err != nil {
		log.Error(err, "apply CRS")
		return err
	}
	return nil
}
