package instanceset_landing

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// Client is the global k8s client.
var Client client.Client

// PrepareClient initializes the global k8s client.
func PrepareClient() error {
	if Client != nil {
		return nil
	}

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(clv1alpha2.AddToScheme(scheme))

	kubeconfig, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("k8s config error: %w", err)
	}

	Client, err = client.New(kubeconfig, client.Options{Scheme: scheme})
	return err
}
