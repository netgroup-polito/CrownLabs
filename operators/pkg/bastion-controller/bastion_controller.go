/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bastion_controller

import (
	"context"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crownlabsalpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
)

// BastionReconciler reconciles a Bastion object
type BastionReconciler struct {
	client.Client
	Log                logr.Logger
	Scheme             *runtime.Scheme
	AuthorizedKeysPath string
}

func closeFile(f *os.File, log logr.Logger) {
	err := f.Close()
	if err != nil {
		log.Error(err, "unable to close the file authorized_keys")
	}
}

// +kubebuilder:rbac:groups=crownlabs.polito.it,resources=tenants,verbs=list

func (r *BastionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("tenant", req.NamespacedName)
	log.Info("reconciling bastion")

	// Get tenants resources
	var tenants crownlabsalpha1.TenantList
	if err := r.List(ctx, &tenants); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Collect public keys of tenants
	keys := make([]string, 0, len(tenants.Items))
	for _, tenant := range tenants.Items {
		keys = append(keys, tenant.Spec.PublicKeys...)
	}

	authorizedKeys := strings.Join(keys[:], "\n")

	log.Info(r.AuthorizedKeysPath)

	f, err := os.Create(r.AuthorizedKeysPath)
	if err != nil {
		log.Error(err, "unable to create the file authorized_keys")
		return ctrl.Result{}, nil
	}

	defer closeFile(f, log)

	_, err = f.WriteString(authorizedKeys)
	if err != nil {
		log.Error(err, "unable to write to authorized_keys")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *BastionReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsalpha1.Tenant{}).
		Complete(r)
}
