package instance_controller

import (
	"context"
	"time"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	virtv1 "kubevirt.io/client-go/api/v1"
	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	instance_creation "github.com/netgroup-polito/CrownLabs/operators/pkg/instance-creation"
)

// CreateVMEnvironment implements the logic to create all the different
// Kubernetes resources required to start a CrownLabs environment.
func (r *InstanceReconciler) CreateVMEnvironment(instance *crownlabsv1alpha2.Instance, environment *crownlabsv1alpha2.Environment, namespace, name string, vmStart time.Time) error {
	var user, password string
	var vmi *virtv1.VirtualMachineInstance
	ctx := context.TODO()
	err := instance_creation.GetWebdavCredentials(ctx, r.Client, r.WebdavSecretName, instance.Namespace, &user, &password)
	if err != nil {
		klog.Error("unable to get Webdav Credentials")
		klog.Error(err)
	} else {
		klog.Info("Webdav secrets obtained. Getting public keys. " + name)
	}
	var publicKeys []string
	if err = instance_creation.GetPublicKeys(ctx, r.Client, instance.Spec.Tenant, instance.Spec.Template, &publicKeys); err != nil {
		klog.Error("unable to get public keys")
		klog.Error(err)
	} else {
		klog.Info("Public keys obtained. Building cloud-init script. " + name)
	}

	// persistent feature
	if environment.Persistent {
		if cancontinue, err1 := r.createPersistentlogic(instance, environment, name); err1 != nil {
			return err1
			// If no errors have happened if datavolume is not succeeded no need to go on
		} else if !cancontinue {
			return nil
		}
	}

	// create secret
	secret := instance_creation.CreateCloudInitSecret(name, namespace, user, password, r.NextcloudBaseURL, publicKeys)
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, &secret, func() error {
		return ctrl.SetControllerReference(instance, &secret, r.Scheme)
	})
	if err != nil {
		r.setInstanceStatus(ctx, "Could not create secret "+secret.Name+" in namespace "+secret.Namespace, "Warning", "SecretNotCreated", instance, "", "")
	} else {
		r.setInstanceStatus(ctx, "Secret "+secret.Name+" correctly created in namespace "+secret.Namespace, "Normal", "SecretCreated", instance, "", "")
	}

	// create Service to expose the vm
	service := instance_creation.ForgeService(name, namespace)
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, &service, func() error {
		return ctrl.SetControllerReference(instance, &service, r.Scheme)
	})
	if err != nil {
		r.setInstanceStatus(ctx, "Could not create service "+service.Name+" in namespace "+service.Namespace, "Warning", "ServiceNotCreated", instance, "", "")
		return err
	}
	r.setInstanceStatus(ctx, "Service "+service.Name+" correctly created in namespace "+service.Namespace, "Normal", "ServiceCreated", instance, "", "")

	urlUUID := uuid.New().String()
	// create Ingress to manage the service
	ingress := instance_creation.ForgeIngress(name, namespace, &service, urlUUID, r.WebsiteBaseURL)
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, &ingress, func() error {
		return ctrl.SetControllerReference(instance, &ingress, r.Scheme)
	})
	if err != nil {
		r.setInstanceStatus(ctx, "Could not create ingress "+ingress.Name+" in namespace "+ingress.Namespace, "Warning", "IngressNotCreated", instance, "", "")
		return err
	}
	r.setInstanceStatus(ctx, "Ingress "+ingress.Name+" correctly created in namespace "+ingress.Namespace, "Normal", "IngressCreated", instance, "", "")

	if err = r.createOAUTHlogic(name, instance, namespace, urlUUID); err != nil {
		return err
	}

	// create vm
	vmi = &virtv1.VirtualMachineInstance{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	vmStatus := "VmiCreated"
	if environment.Persistent {
		vm := virtv1.VirtualMachine{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
		_, err = ctrl.CreateOrUpdate(ctx, r.Client, &vm, func() error {
			instance_creation.UpdateVirtualMachineSpec(&vm, environment, instance.Spec.Running)
			vm.Spec.Template.ObjectMeta.Labels = instance_creation.UpdateLabels(vm.Spec.Template.ObjectMeta.Labels, environment, name)
			return ctrl.SetControllerReference(instance, &vm, r.Scheme)
		})
		if !instance.Spec.Running {
			vmStatus = "VmiOff"
		}
	} else {
		_, err = ctrl.CreateOrUpdate(ctx, r.Client, vmi, func() error {
			if vmi.ObjectMeta.CreationTimestamp.IsZero() {
				instance_creation.UpdateVirtualMachineInstanceSpec(vmi, environment)
			}
			vmi.Labels = instance_creation.UpdateLabels(vmi.Labels, environment, name)
			return ctrl.SetControllerReference(instance, vmi, r.Scheme)
		})
	}
	if err != nil {
		r.setInstanceStatus(ctx, "Could not create vmi "+vmi.Name+" in namespace "+namespace, "Warning", "VmiNotCreated", instance, "", "")
		return err
	}
	r.setInstanceStatus(ctx, "VirtualMachineInstance "+vmi.Name+" correctly created in namespace "+namespace, "Normal", vmStatus, instance, "", "")

	if vmStatus != "VmiOff" {
		go r.getVmiStatus(ctx, environment.GuiEnabled, &service, &ingress, instance, vmi, vmStart)
	}

	return nil
}

func (r *InstanceReconciler) createOAUTHlogic(name string, instance *crownlabsv1alpha2.Instance, namespace, urlUUID string) error {
	ctx := context.TODO()

	// create Service for oauth2
	oauthService := instance_creation.ForgeOauth2Service(name, namespace)
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &oauthService, func() error {
		return ctrl.SetControllerReference(instance, &oauthService, r.Scheme)
	}); err != nil {
		r.setInstanceStatus(ctx, "Could not create service "+oauthService.Name+" in namespace "+oauthService.Namespace, "Warning", "Oauth2ServiceNotCreated", instance, "", "")
		return err
	}
	r.setInstanceStatus(ctx, "Service "+oauthService.Name+" correctly created in namespace "+oauthService.Namespace, "Normal", "Oauth2ServiceCreated", instance, "", "")

	// create Ingress to manage the oauth2 service
	oauthIngress := instance_creation.ForgeOauth2Ingress(name, namespace, &oauthService, urlUUID, r.WebsiteBaseURL)
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &oauthIngress, func() error {
		return ctrl.SetControllerReference(instance, &oauthIngress, r.Scheme)
	}); err != nil {
		r.setInstanceStatus(ctx, "Could not create ingress "+oauthIngress.Name+" in namespace "+oauthIngress.Namespace, "Warning", "Oauth2IngressNotCreated", instance, "", "")
		return err
	}
	r.setInstanceStatus(ctx, "Ingress "+oauthIngress.Name+" correctly created in namespace "+oauthIngress.Namespace, "Normal", "Oauth2IngressCreated", instance, "", "")

	// create Deployment for oauth2
	oauthDeploy := instance_creation.ForgeOauth2Deployment(name, namespace, urlUUID, r.Oauth2ProxyImage, r.OidcClientSecret, r.OidcProviderURL)
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &oauthDeploy, func() error {
		return ctrl.SetControllerReference(instance, &oauthDeploy, r.Scheme)
	}); err != nil {
		r.setInstanceStatus(ctx, "Could not create deployment "+oauthDeploy.Name+" in namespace "+oauthDeploy.Namespace, "Warning", "Oauth2DeployNotCreated", instance, "", "")
		return err
	}
	r.setInstanceStatus(ctx, "Deployment "+oauthDeploy.Name+" correctly created in namespace "+oauthDeploy.Namespace, "Normal", "Oauth2DeployCreated", instance, "", "")
	return nil
}

func (r *InstanceReconciler) createPersistentlogic(instance *crownlabsv1alpha2.Instance, environment *crownlabsv1alpha2.Environment, name string) (bool, error) {
	ctx := context.TODO()
	// create datavolume
	dv := cdiv1.DataVolume{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: instance.Namespace}}
	dvOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &dv, func() error {
		instance_creation.UpdateDataVolumeSpec(&dv, environment)
		return ctrl.SetControllerReference(instance, &dv, r.Scheme)
	})
	if err != nil {
		klog.Error(err, "unable to create or update DataVolume ")
		return false, err
	}
	klog.Infof("Data volume correctly %s", dvOpRes)

	// check if datavolume is succeeded
	if dv.Status.Phase != cdiv1.DataVolumePhase("Succeeded") {
		r.setInstanceStatus(ctx, "PVC "+dv.Name+" importing", "Normal", "Importing", instance, "", "")
		return false, nil
	}

	r.setInstanceStatus(ctx, "PVC "+dv.Name+" import completed", "Normal", "ImportCompleted", instance, "", "")
	return true, nil
}
