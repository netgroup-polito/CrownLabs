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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/clcontext"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// EnforceClusterEnvironment implements the logic to create all the Cluster API (and related providers)
// resources required to provision a nested Kubernetes cluster environment, based on the ClusterFlavor
// referenced by name through the environment image in the configured central namespace.
func (r *InstanceReconciler) EnforceClusterEnvironment(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	envIndex := clctx.EnvironmentIndexFrom(ctx)

	if environment.Cluster == nil {
		return fmt.Errorf("environment %q of type Cluster is missing the cluster specification", environment.Name)
	}

	namespace := r.ClusterFlavorNamespace
	if namespace == "" {
		namespace = forge.DefaultClusterFlavorNamespace
	}
	flavorName := types.NamespacedName{Name: environment.Image, Namespace: namespace}
	var flavor clv1alpha2.ClusterFlavor
	if err := r.Get(ctx, flavorName, &flavor); err != nil {
		log.Error(err, "failed retrieving the cluster flavor", "clusterflavor", flavorName)
		r.EventsRecorder.Eventf(instance, corev1.EventTypeWarning, EvClusterFlavorNotFound, EvClusterFlavorNotFoundMsg, flavorName.Namespace, flavorName.Name)
		return err
	}

	names := forge.ClusterNamesFromInstance(instance, environment)
	selectorLabels := forge.ClusterSelectorLabels(instance, environment)
	labels := forge.EnvironmentObjectLabels(nil, instance, environment)

	cluster, err := r.enforceClusterAPIObject(ctx, forge.GVKCluster, names.Cluster, instance.Namespace, labels, forge.ClusterSpec(names), nil)
	if err != nil {
		return err
	}

	if _, err := r.enforceClusterAPIObject(ctx, forge.GVKKubevirtCluster, names.Infrastructure, instance.Namespace, labels, nil, forge.KubevirtClusterAnnotations()); err != nil {
		return err
	}

	if _, err := r.enforceClusterAPIObject(ctx, forge.GVKKamajiControlPlane, names.ControlPlane, instance.Namespace, labels, forge.KamajiControlPlaneSpec(&flavor), nil); err != nil {
		return err
	}

	if _, err := r.enforceClusterAPIObject(ctx, forge.GVKMachineDeployment, names.MachineDeployment, instance.Namespace, labels, forge.MachineDeploymentSpec(names.Cluster, names, flavor.Spec.Version, environment.Cluster.WorkersCount), nil); err != nil {
		return err
	}

	machineTemplateSpec, err := forge.KubevirtMachineTemplateSpec(instance.Namespace, &flavor.Spec.Workers, &environment.Resources)
	if err != nil {
		return err
	}
	if _, err := r.enforceClusterAPIObject(ctx, forge.GVKKubevirtMachineTemplate, names.MachineDeployment, instance.Namespace, labels, machineTemplateSpec, nil); err != nil {
		return err
	}

	if _, err := r.enforceClusterAPIObject(ctx, forge.GVKKubeadmConfigTemplate, names.MachineDeployment, instance.Namespace, labels, forge.KubeadmConfigTemplateSpec(), nil); err != nil {
		return err
	}

	for i := range flavor.Spec.AddonList {
		addon := flavor.Spec.AddonList[i]
		addonName := forge.HelmChartProxyName(names, &addon)
		if _, err := r.enforceClusterAPIObject(ctx, forge.GVKHelmChartProxy, addonName, instance.Namespace, labels, forge.HelmChartProxySpec(&addon, selectorLabels), nil); err != nil {
			return err
		}
	}

	instance.Status.Environments[envIndex].Phase = r.RetrievePhaseFromCluster(cluster)
	return nil
}

// enforceClusterAPIObject creates or updates a single unstructured Cluster API (or related provider) resource,
// owned by the current instance, converting the provider's typed spec at the client boundary.
func (r *InstanceReconciler) enforceClusterAPIObject(
	ctx context.Context,
	gvk schema.GroupVersionKind,
	name, namespace string,
	labels map[string]string,
	spec interface{},
	annotations map[string]string,
) (*unstructured.Unstructured, error) {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)

	obj := forge.NewClusterAPIObject(gvk, name, namespace)
	var specMap map[string]interface{}
	if spec != nil {
		var err error
		specMap, err = clusterSpecToUnstructured(spec)
		if err != nil {
			return nil, fmt.Errorf("convert %s spec: %w", gvk.Kind, err)
		}
	}

	mutate := func() error {
		obj.SetGroupVersionKind(gvk)
		obj.SetLabels(labels)
		if specMap != nil {
			obj.Object["spec"] = runtime.DeepCopyJSON(specMap)
		}
		if annotations != nil {
			obj.SetAnnotations(annotations)
		}

		if gvk == forge.GVKCluster || gvk == forge.GVKHelmChartProxy {
			return ctrl.SetControllerReference(instance, obj, r.Scheme)
		}

		return nil
	}

	// Apply the mutation once upfront, so the desired manifest can be logged
	// before CreateOrUpdate possibly overwrites obj with the existing server state.
	if err := mutate(); err != nil {
		return nil, err
	}

	objYAML, err := yaml.Marshal(obj.Object)
	if err != nil {
		log.Error(err, "failed to serialize desired object", "kind", gvk.Kind, "object", klog.KRef(namespace, name))
		return nil, err
	}

	res, err := ctrl.CreateOrUpdate(ctx, r.Client, obj, mutate)
	if err != nil {
		log.Error(err, "failed to enforce object", "kind", gvk.Kind, "object", klog.KRef(namespace, name))
		return nil, err
	}
	log.V(utils.FromResult(res)).Info("object enforced", "kind", gvk.Kind, "object", klog.KRef(namespace, name), "result", objYAML)
	return obj, nil
}

func clusterSpecToUnstructured(spec interface{}) (map[string]interface{}, error) {
	// JSON normalizes unsigned KubeVirt fields to types supported by DeepCopyJSON.
	// The Kubernetes decoder preserves integers as int64 rather than float64.
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	var specMap map[string]interface{}
	if err := json.Unmarshal(data, &specMap); err != nil {
		return nil, err
	}
	return specMap, nil
}

// RetrievePhaseFromCluster converts the status of a Cluster API Cluster resource to the corresponding
// EnvironmentPhase of the instance.
func (r *InstanceReconciler) RetrievePhaseFromCluster(cluster *unstructured.Unstructured) clv1alpha2.EnvironmentPhase {
	if !cluster.GetDeletionTimestamp().IsZero() {
		return clv1alpha2.EnvironmentPhaseStopping
	}

	phase, _, _ := unstructured.NestedString(cluster.Object, "status", "phase")

	switch phase {
	case "Failed":
		return clv1alpha2.EnvironmentPhaseFailed
	case "Deleting":
		return clv1alpha2.EnvironmentPhaseStopping
	case "Provisioning", "Pending", "":
		return clv1alpha2.EnvironmentPhaseImporting
	case "Provisioned":
		controlPlaneReady, _, _ := unstructured.NestedBool(cluster.Object, "status", "controlPlaneReady")
		infrastructureReady, _, _ := unstructured.NestedBool(cluster.Object, "status", "infrastructureReady")
		if controlPlaneReady && infrastructureReady {
			return clv1alpha2.EnvironmentPhaseReady
		}
		return clv1alpha2.EnvironmentPhaseRunning
	default:
		return clv1alpha2.EnvironmentPhaseStarting
	}
}
