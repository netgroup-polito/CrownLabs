package forge

import (
	"context"
	"os"

	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createCalicoConfigMap(k8sClient client.Client, ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	data, err := os.ReadFile("./cniconfigs/calico.yaml")
	if err != nil {
		log.Error(err, "failed to enforce the instance exposition objects")
		return err
	}
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "calico-manifest",
			Namespace: instance.Namespace,
		},
		Data: map[string]string{
			"calico.yaml": string(data),
		},
	}
	return k8sClient.Create(ctx, cm)
}
