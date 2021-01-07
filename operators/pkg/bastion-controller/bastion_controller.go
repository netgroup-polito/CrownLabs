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
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

// +kubebuilder:rbac:groups=crownlabs.polito.it,resources=tenants,verbs=get;watch

func (r *BastionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("tenant", req.NamespacedName)
	log.Info("reconciling bastion")

	tenant := &crownlabsalpha1.Tenant{}
	deleted := false

	if err := r.Get(ctx, req.NamespacedName, tenant); apierrors.IsNotFound(err) {
		deleted = true
	} else if err != nil {
		return ctrl.Result{}, err
	}

	var keys []string

	if _, err := os.Stat(r.AuthorizedKeysPath); err == nil {
		// if the file exists, read the whole file in a []byte
		data, err := ioutil.ReadFile(r.AuthorizedKeysPath)
		if err != nil {
			log.Error(err, "unable to read the file authorized_keys")
			return ctrl.Result{}, err
		}

		if len(data) > 0 {
			keys = decomposeAndPurgeEntries(strings.Split(string(data), string("\n")), req.NamespacedName.Name)
		}
	}

	if !deleted {
		// if the event was NOT a deletion, add the tenant's keys. Otherwise nothing to do.
		keys = composeAndMarkEntries(keys, tenant.Spec.PublicKeys, req.NamespacedName.Name)
	}

	f, err := os.Create(r.AuthorizedKeysPath)
	if err != nil {
		log.Error(err, "unable to create the file authorized_keys")
		return ctrl.Result{}, nil
	}

	defer closeFile(f, log)

	if len(keys) > 0 {
		_, err = f.Write([]byte(strings.Join(keys, string("\n"))))
		if err != nil {
			log.Error(err, "unable to write to authorized_keys")
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *BastionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsalpha1.Tenant{}).
		Complete(r)
}
