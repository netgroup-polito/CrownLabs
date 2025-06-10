package instctrl

import (
	"context"
	"fmt"

	controlplanekamajiv1 "github.com/clastix/cluster-api-control-plane-provider-kamaji/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	infrav1 "sigs.k8s.io/cluster-api-provider-kubevirt/api/v1alpha1"
	capiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InstanceReconciler enforces the Cluster API environment for a CrownLabs instance
// Kubernetes resources required to start a CrownLabs environment.
func (r *InstanceReconciler) EnforceClusterEnvironment(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	Provider := environment.Cluster.ControlPlane.Provider
	instance := clctx.InstanceFrom(ctx)
	host := forge.HostName(r.ServiceUrls.WebsiteBaseURL, environment.Mode)
	r.enforceCluster(ctx)
	// choose the a proper controlplabe provider
	if Provider == clv1alpha2.ProviderKubeadm {
		r.enforceKubeadmInfra(ctx)
		r.enforceKubeadmControlPlane(ctx)
	} else {
		r.enforceKamajiInfra(ctx)
		r.enforceKamajiControlPlane(ctx)
	}
	// enforce a machinedeployment for VM management
	r.enforceMachineDeployment(ctx)
	// enforce a worker virtual machine template
	r.enforceKubevirtMachine(ctx)
	// enforce a boostrap for woker virtual machines
	r.enforceBootstrap(ctx)
	// Enforce the service and the ingress to expose the environment.
	err := r.EnforceInstanceExposition(ctx)
	if err != nil {
		log.Error(err, "failed to enforce the instance exposition objects")
		return err
	}
	// install cni and export kubeconfig
	forge.Insinstallcni(instance, environment, host)
	// echo to template status
	r.updatetemplatestatus(ctx)
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

// enforceInfra creates or updates the KubevirtCluster resource and labels it for CAPI
func (r *InstanceReconciler) enforceKubeadmInfra(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	cluster := environment.Cluster
	infra := &infrav1.KubevirtCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-infra", cluster.Name),
			Namespace: instance.Namespace,
		},
	}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, infra, func() error {
		if infra.CreationTimestamp.IsZero() {
			infra.Spec.ControlPlaneServiceTemplate.Spec.Type = corev1.ServiceType(cluster.ServiceType)
		}
		if infra.Labels == nil {
			infra.Labels = map[string]string{}
		}
		infra.SetLabels(forge.InstanceObjectLabels(infra.GetLabels(), instance))
		// unnecessary to set contoller ref, all will be managed by cluster
		return nil
	})
	if err != nil {
		log.Error(err, "failed to enforce infrastructure", "infra", klog.KObj(infra))
		return err
	}
	log.V(utils.FromResult(res)).Info("Infrastructure enforced", "infra", klog.KObj(infra), "result", res)
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

// enforceControlPlane creates or updates the KubeadmControlPlane resource and labels it
func (r *InstanceReconciler) enforceKubeadmControlPlane(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	cluster := environment.Cluster
	controlplane := cluster.ControlPlane
	cp := &controlplanev1.KubeadmControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-control-plane", cluster.Name),
			Namespace: instance.Namespace,
		},
	}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, cp, func() error {

		if cp.CreationTimestamp.IsZero() {
			cp.Spec.Version = cluster.Version
			host := forge.HostName(r.ServiceUrls.WebsiteBaseURL, environment.Mode)
			cp.Spec = forge.ClusterControlPlaneSepc(instance, environment, host)
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
	controlplane := cluster.ControlPlane

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

	// control-plane template
	if controlplane.Provider == v1alpha2.ProviderKubeadm {
		wmcp := infrav1.KubevirtMachineTemplate{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-control-plane-machine", cluster.Name), Namespace: instance.Namespace}}
		res, err := ctrl.CreateOrUpdate(ctx, r.Client, &wmcp, func() error {

			if wmcp.CreationTimestamp.IsZero() {
				wmcp.Spec.Template.Spec.BootstrapCheckSpec.CheckStrategy = "ssh"

				vmSpec := forge.ClusterVMSpec(environment)
				wmcp.Spec.Template.Spec.VirtualMachineTemplate.Spec = vmSpec
			}

			if wmcp.Labels == nil {
				wmcp.Labels = map[string]string{}
			}
			wmcp.Labels[capiv1.ClusterNameLabel] = fmt.Sprintf("%s-cluster", cluster.Name)
			return nil
		})
		if err != nil {
			log.Error(err, "failed to enforce virtualmachine-control-plane")
			return err
		}
		log.V(utils.FromResult(res)).Info("virtualmachine-control-plane enforced")
	}
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
func (r *InstanceReconciler) updatetemplatestatus(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	cluster := environment.Cluster

	var tmpl clv1alpha2.Template
	if err := r.Get(ctx, client.ObjectKey{
		Name:      instance.Spec.Template.Name,
		Namespace: instance.Spec.Template.Namespace,
	}, &tmpl); err != nil {
		return err
	}

	tmpl.Status.KubeConfigs = []clv1alpha2.KubeconfigTemplate{{
		Name:        fmt.Sprintf("%s-cluster", cluster.Name),
		FileAddress: fmt.Sprintf("./kubeconfigs/%s-cluster.kubeconfig", cluster.Name),
	}}

	if err := r.Status().Update(ctx, &tmpl); err != nil {
		log.Error(err, "failed to update template status")
		return err
	}
	return nil
}
