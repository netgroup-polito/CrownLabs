package instctrl

import (
	"context"
	"fmt"
	"io"
	"net/http"

	controlplanekamajiv1 "github.com/clastix/cluster-api-control-plane-provider-kamaji/api/v1alpha1"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	addonsv1beta1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"

	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	infrav1 "sigs.k8s.io/cluster-api-provider-kubevirt/api/v1alpha1"
	capiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// InstanceReconciler enforces the Cluster API environment for a CrownLabs instance
// Kubernetes resources required to start a CrownLabs environment.
func (r *InstanceReconciler) EnforceClusterEnvironment(ctx context.Context) error {
	r.enforceCluster(ctx)
	r.enforceKamajiInfra(ctx)
	r.enforceKamajiControlPlane(ctx)

	// enforce a machinedeployment for VM management
	r.enforceMachineDeployment(ctx)
	// enforce a worker virtual machine template
	r.enforceKubevirtMachine(ctx)
	// enforce a boostrap for woker virtual machines
	r.enforceBootstrap(ctx)
	r.enforceCalicoCni(ctx)
	// install cni and export kubeconfig
	// echo to template status
	//r.updatetemplatestatus(ctx)
	return nil
}

// enforceCluster creates or updates the Cluster resource and sets its OwnerRef
func (r *InstanceReconciler) enforceCluster(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	cluster := environment.Cluster
	cl := &capiv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-cluster", cluster.Name),
			Namespace: instance.Namespace,
		},
	}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, cl, func() error {
		if cl.CreationTimestamp.IsZero() {
			// align infrastructure and controlPlane refs
			cl.Spec = forge.ClusterSpec(instance, environment)
		}
		cl.SetLabels(forge.InstanceObjectLabels(cl.GetLabels(), instance))
		return ctrl.SetControllerReference(instance, cl, r.Scheme)
	})
	if err != nil {
		log.Error(err, "failed to enforce cluster", "cluster", klog.KObj(cl))
		return err
	}
	log.V(utils.FromResult(res)).Info("Cluster enforced", "Cluster", klog.KObj(cl), "result", res)
	return nil
}

// kamaji infra
func (r *InstanceReconciler) enforceKamajiInfra(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	cluster := environment.Cluster
	infra := &infrav1.KubevirtCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-infra", cluster.Name),
			Namespace:   instance.Namespace,
			Annotations: map[string]string{"cluster.x-k8s.io/managed-by": "kamaji"},
		},
	}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, infra, func() error {
		if infra.Labels == nil {
			infra.Labels = map[string]string{}
		}
		infra.SetLabels(forge.InstanceObjectLabels(infra.GetLabels(), instance))
		return nil
	})
	if err != nil {
		log.Error(err, "failed to enforce infrastructure", "infra", klog.KObj(infra))
		return err
	}
	log.V(utils.FromResult(res)).Info("Infrastructure enforced", "infra", klog.KObj(infra), "result", res)
	return nil
}

// kamaji controlplane
func (r *InstanceReconciler) enforceKamajiControlPlane(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	cluster := environment.Cluster
	controlplane := cluster.ControlPlane
	cp := &controlplanekamajiv1.KamajiControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-control-plane", cluster.Name),
			Namespace: instance.Namespace,
		},
	}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, cp, func() error {
		if cp.CreationTimestamp.IsZero() {
			host := forge.HostName(r.ServiceUrls.WebsiteBaseURL, environment.Mode)
			cp.Spec = forge.KamajiControlPlaneSpec(environment, host)
		}
		cp.Spec.Replicas = ptr.To(int32(controlplane.Replicas))
		if cp.Labels == nil {
			cp.Labels = map[string]string{}
		}
		cp.SetLabels(forge.InstanceObjectLabels(cp.GetLabels(), instance))
		return nil
	})
	if err != nil {
		log.Error(err, "failed to enforce controlplane", "cp", klog.KObj(cp))
		return err
	}
	log.V(utils.FromResult(res)).Info("ControlPlane enforced", "cp", klog.KObj(cp), "result", res)
	return nil
}

// enforceMachineDeployment creates or updates the MachineDeployment and labels it
func (r *InstanceReconciler) enforceMachineDeployment(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	cluster := environment.Cluster
	machinedeployment := cluster.MachineDeploy
	md := &capiv1.MachineDeployment{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-md", cluster.Name), Namespace: instance.Namespace},
	}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, md, func() error {
		if md.CreationTimestamp.IsZero() {
			md.Spec.ClusterName = fmt.Sprintf("%s-cluster", cluster.Name)
			md.Spec.Template.Spec = forge.MachineDeploymentSepc(instance, environment)
		}
		md.Spec.Replicas = ptr.To(int32(machinedeployment.Replicas))
		if md.Labels == nil {
			md.Labels = map[string]string{}
		}
		md.SetLabels(forge.InstanceObjectLabels(md.GetLabels(), instance))
		return nil
	})
	if err != nil {
		log.Error(err, "failed to enforce machinedeployment")
		return err
	}
	log.V(utils.FromResult(res)).Info("virtualmachine enforced", "virtualmachine", klog.KObj(md), "result", res)
	return nil
}

// enforceKubevirtMachine creates or updates KubevirtMachineTemplates with RunStrategy and DV mapping
func (r *InstanceReconciler) enforceKubevirtMachine(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	cluster := environment.Cluster
	// worker template
	wmworker := infrav1.KubevirtMachineTemplate{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-md-worker", cluster.Name), Namespace: instance.Namespace}}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &wmworker, func() error {
		if wmworker.CreationTimestamp.IsZero() {
			wmworker.Spec.Template.Spec.BootstrapCheckSpec.CheckStrategy = "ssh"

			vmSpec := forge.ClusterVMSpec(environment)
			wmworker.Spec.Template.Spec.VirtualMachineTemplate.Spec = vmSpec
		}
		if wmworker.Labels == nil {
			wmworker.Labels = map[string]string{}
		}
		wmworker.Labels[capiv1.ClusterNameLabel] = fmt.Sprintf("%s-cluster", cluster.Name)
		return nil
	})
	if err != nil {
		log.Error(err, "failed to enforce virtualmachine-worker")
		return err
	}
	log.V(utils.FromResult(res)).Info("virtualmachine-worker enforced")

	return nil
}

// enforceBootstrap creates or updates the KubeadmConfigTemplate and labels it
func (r *InstanceReconciler) enforceBootstrap(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	cluster := environment.Cluster
	bt := bootstrapv1.KubeadmConfigTemplate{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-md-bootstrap", cluster.Name), Namespace: instance.Namespace}}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &bt, func() error {
		if bt.CreationTimestamp.IsZero() {
			bt.Spec.Template.Spec.JoinConfiguration = &bootstrapv1.JoinConfiguration{
				NodeRegistration: bootstrapv1.NodeRegistrationOptions{
					KubeletExtraArgs: map[string]string{},
				},
			}
		}
		if bt.Labels == nil {
			bt.Labels = map[string]string{}
		}
		bt.SetLabels(forge.InstanceObjectLabels(bt.GetLabels(), instance))
		return nil
	})
	if err != nil {
		log.Error(err, "failed to enforce bootstrap")
		return err
	}
	log.V(utils.FromResult(res)).Info("bootstrap enforced")
	return nil
}

// insertKubeConfig export the KUBECONFIG into kubeconfig folders

// update the template status with the address of  relative kubeconfig file
// func (r *InstanceReconciler) updatetemplatestatus(ctx context.Context) error {
// 	log := ctrl.LoggerFrom(ctx)
// 	environment := clctx.EnvironmentFrom(ctx)
// 	instance := clctx.InstanceFrom(ctx)
// 	cluster := environment.Cluster
// 	var tmpl clv1alpha2.Template
// 	if err := r.Get(ctx, client.ObjectKey{
// 		Name:      instance.Spec.Template.Name,
// 		Namespace: instance.Spec.Template.Namespace,
// 	}, &tmpl); err != nil {
// 		return err
// 	}
// 	tmpl.Status.KubeConfigs = []clv1alpha2.KubeconfigTemplate{{
// 		Name:        fmt.Sprintf("%s-cluster", cluster.Name),
// 		FileAddress: fmt.Sprintf("./kubeconfigs/%s-instance.kubeconfig", instance.Name),
// 	}}
// 	if err := r.Status().Update(ctx, &tmpl); err != nil {
// 		log.Error(err, "failed to update template status")
// 		return err
// 	}
// 	return nil
// }

// deploy cni
func (r *InstanceReconciler) enforceCalicoCni(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	log := ctrl.LoggerFrom(ctx)
	ns := instance.Namespace
	environment := clctx.EnvironmentFrom(ctx)
	cluster := environment.Cluster
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
		},
	}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, cm, func() error {
		if cm.CreationTimestamp.IsZero() {
			cm.Labels = map[string]string{
				"kamaji.clastix.io/tenant": tenant,
			}
			cm.Data = map[string]string{
				"calico.yaml": string(yamlBytes),
			}
		}
		if cm.Labels == nil {
			cm.Labels = map[string]string{}
		}
		cm.Labels[capiv1.ClusterNameLabel] = fmt.Sprintf("%s-cluster", cluster.Name)
		return ctrl.SetControllerReference(instance, cm, r.Scheme)
	})
	if err != nil {
		log.Error(err, "failed to create configmap calico")
		return err
	}
	log.V(utils.FromResult(res)).Info("calico configmap enforced")
	crs := &addonsv1beta1.ClusterResourceSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "install-calico",
			Namespace: ns,
		},
	}
	res, err = ctrl.CreateOrUpdate(ctx, r.Client, crs, func() error {
		if crs.CreationTimestamp.IsZero() {
			crs.Spec = addonsv1beta1.ClusterResourceSetSpec{
				Strategy: string(addonsv1beta1.ClusterResourceSetStrategyApplyOnce),
				ClusterSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"crownlabs.polito.it/tenant": tenant,
					},
				},
				Resources: []addonsv1beta1.ResourceRef{
					{
						Name: "calico-manifest",
						Kind: "ConfigMap",
					},
				},
			}
		}
		if crs.Labels == nil {
			crs.Labels = map[string]string{}
		}
		crs.Labels[capiv1.ClusterNameLabel] = fmt.Sprintf("%s-cluster", cluster.Name)
		return ctrl.SetControllerReference(instance, crs, r.Scheme)
	})
	if err != nil {
		log.Error(err, "failed to create configmap crs")
		return err
	}
	log.V(utils.FromResult(res)).Info("crs enforced")
	return nil
}
