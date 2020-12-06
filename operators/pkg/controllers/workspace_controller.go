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

package controllers

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=crownlabs.polito.it,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crownlabs.polito.it,resources=workspaces/status,verbs=get;update;patch

// Reconcile reconciles the state of a workspace resource
func (r *WorkspaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	var ws crownlabsv1alpha1.Workspace

	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		// reconcile was triggered by a delete request
		klog.Infof("Workspace %s deleted", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	klog.Infof("Reconciling workspace %s", req.Name)

	nsName := fmt.Sprintf("workspace-%s", ws.Name)
	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}

	nsOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &ns, func() error {
		updateNamespace(ws, &ns, nsName)
		return ctrl.SetControllerReference(&ws, &ns, r.Scheme)
	})
	if err != nil {
		klog.Error("Unable to create or update namespace", err)
		// update status of workspace with failed namespace creation
		ws.Status.Namespace.Created = false
		ws.Status.Namespace.Name = ""
		// return anyway the error to allow new reconcile, independently of outcome of status update
		if err := r.Status().Update(ctx, &ws); err != nil {
			klog.Error("Unable to update status", err)
		}
		return ctrl.Result{}, err
	}
	klog.Infof("Namespace %s for workspace %s %s", nsName, req.Name, nsOpRes)

	// update status of workspace with info about namespace, success
	ws.Status.Namespace.Created = true
	ws.Status.Namespace.Name = nsName
	if err := r.Status().Update(ctx, &ws); err != nil {
		// if status update fails, still try to reconcile later
		klog.Error("Unable to update status", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha1.Workspace{}).
		Complete(r)
}

func updateNamespace(ws crownlabsv1alpha1.Workspace, ns *v1.Namespace, wsnsName string) {
	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}
	ns.Labels["crownlabs.polito.it/type"] = "workspace"
}
