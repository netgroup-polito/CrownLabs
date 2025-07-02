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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	infrav1 "sigs.k8s.io/cluster-api-provider-kubevirt/api/v1alpha1"
	capiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	addonsv1beta1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// InstanceReconciler enforces the Cluster API environment for a CrownLabs instance
// Kubernetes resources required to start a CrownLabs environment.
func (r *InstanceReconciler) EnforceClusterEnvironment(ctx context.Context) error {
	r.enforceClusterRole(ctx)
	r.enforceGuiDeployment(ctx)
	r.enforceGuiSVC(ctx)
	r.enforceCluster(ctx)
	r.enforceKamajiInfra(ctx)
	r.enforceKamajiControlPlane(ctx)
	r.enforceroleBinding(ctx)
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

// deploy cni
func (r *InstanceReconciler) enforceCalicoCni(ctx context.Context) error {
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
		return ctrl.SetControllerReference(instance, crs, r.Scheme)
	})
	if err != nil {
		log.Error(err, "failed to create configmap crs")
		return err
	}
	log.V(utils.FromResult(res)).Info("crs enforced")
	return nil
}

func (r *InstanceReconciler) enforceGuiDeployment(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	log := ctrl.LoggerFrom(ctx)
	ns := instance.Namespace
	environment := clctx.EnvironmentFrom(ctx)

	dep := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "capi-visualizer",
			Namespace: ns,
		},
	}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &dep, func() error {
		if dep.CreationTimestamp.IsZero() {
			dep.Labels = map[string]string{
				"app": "capi-visualizer",
			}
			dep.Spec = forge.GuiDeploymentSpec(instance, environment)
		}
		return ctrl.SetControllerReference(instance, &dep, r.Scheme)
	})
	if err != nil {
		log.Error(err, "failed to create configmap GuiDeployment")
		return err
	}
	log.V(utils.FromResult(res)).Info("GuiDeployment enforced")
	return nil
}

func (r *InstanceReconciler) enforceGuiSVC(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	log := ctrl.LoggerFrom(ctx)
	ns := instance.Namespace
	environment := clctx.EnvironmentFrom(ctx)

	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "capi-visualizer",
			Namespace: ns,
		},
	}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		if svc.CreationTimestamp.IsZero() {
			svc.Labels = map[string]string{
				"app": "capi-visualizer",
			}
			svc.Spec = forge.GuiServiceSpec(instance, environment)
		}
		return ctrl.SetControllerReference(instance, &svc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "failed to create configmap GuiService")
		return err
	}
	log.V(utils.FromResult(res)).Info("GuiService enforced")
	svccnt := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "capi-visualizer",
			Namespace: ns,
		},
	}
	res, err = ctrl.CreateOrUpdate(ctx, r.Client, svccnt, func() error {
		return ctrl.SetControllerReference(instance, svccnt, r.Scheme)
	})
	if err != nil {
		log.Error(err, "failed to create configmap GuiServiceacnt")
		return err
	}
	log.V(utils.FromResult(res)).Info("GuiServicecnt enforced")
	return nil
}

func (r *InstanceReconciler) enforceroleBinding(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	log := ctrl.LoggerFrom(ctx)
	ns := instance.Namespace

	roleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "capi-visualizer" + ns,
		},
	}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, roleBinding, func() error {
		if roleBinding.CreationTimestamp.IsZero() {
			roleBinding.RoleRef = rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "capi-visualizer",
			}
			roleBinding.Subjects = []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "capi-visualizer",
					Namespace: ns,
				},
			}
		}
		return nil
	})
	if err != nil {
		log.Error(err, "failed to create configmap roleBinding")
		return err
	}
	log.V(utils.FromResult(res)).Info("roleBinding enforced")
	return nil
}

func (r *InstanceReconciler) enforceClusterRole(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	ns := instance.Namespace
	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "capi-visualizer" + ns,
		},
	}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, role, func() error {
		if role.CreationTimestamp.IsZero() {
			role.Rules = []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"*"},
					Verbs:     []string{"get", "list", "watch"},
				},
				{
					APIGroups: []string{"apiextensions.k8s.io"},
					Resources: []string{"customresourcedefinitions"},
					Verbs:     []string{"get", "list", "watch"},
				},
				{
					APIGroups: []string{
						"cluster.x-k8s.io",
						"bootstrap.cluster.x-k8s.io",
						"addons.cluster.x-k8s.io",
						"infrastructure.cluster.x-k8s.io",
						"controlplane.cluster.x-k8s.io",
						"ipam.cluster.x-k8s.io",
						"runtime.cluster.x-k8s.io",
					},
					Resources: []string{"*"},
					Verbs:     []string{"*"},
				},
				{
					APIGroups: []string{"*"},
					Resources: []string{"*"},
					Verbs:     []string{"get", "list", "watch"},
				},
				{
					NonResourceURLs: []string{"*"},
					Verbs:           []string{"*"},
				},
			}
		}
		return nil
	})
	if err != nil {
		log.Error(err, "failed to create configmap Role")
		return err
	}
	log.V(utils.FromResult(res)).Info("Role enforced")
	return nil
}
