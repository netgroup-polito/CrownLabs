package instance_controller

import (
	"context"
	"time"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	instance_creation "github.com/netgroup-polito/CrownLabs/operators/pkg/instance-creation"
)

func (r *InstanceReconciler) CreateVMEnvironment(instance *crownlabsv1alpha2.Instance, environment *crownlabsv1alpha2.Environment, namespace, name string, vmStart time.Time) error {
	var user, password string
	// this is added so that all resources created for this Instance are destroyed when the Instance is deleted
	b := true
	globalOwnerReference := []metav1.OwnerReference{
		{
			APIVersion:         instance.APIVersion,
			Kind:               instance.Kind,
			Name:               instance.Name,
			UID:                instance.UID,
			BlockOwnerDeletion: &b,
		},
	}
	ctx := context.TODO()
	err := instance_creation.GetWebdavCredentials(ctx, r.Client, r.WebdavSecretName, instance.Namespace, &user, &password)
	if err != nil {
		klog.Error("unable to get Webdav Credentials")
		klog.Error(err)
	} else {
		klog.Info("Webdav secrets obtained. Getting public keys. " + instance.Name)
	}
	var publicKeys []string
	if err = instance_creation.GetPublicKeys(ctx, r.Client, instance.Spec.Tenant, instance.Spec.Template, &publicKeys); err != nil {
		klog.Error("unable to get public keys")
		klog.Error(err)
	} else {
		klog.Info("Public keys obtained. Building cloud-init script. " + instance.Name)
	}
	secret := instance_creation.CreateCloudInitSecret(name, namespace, user, password, r.NextcloudBaseURL, publicKeys, globalOwnerReference)
	if err = instance_creation.CreateOrUpdate(ctx, r.Client, secret); err != nil {
		r.setInstanceStatus(ctx, "Could not create secret "+secret.Name+" in namespace "+secret.Namespace, "Warning", "SecretNotCreated", instance, "", "")
	} else {
		r.setInstanceStatus(ctx, "Secret "+secret.Name+" correctly created in namespace "+secret.Namespace, "Normal", "SecretCreated", instance, "", "")
	}

	// create Service to expose the vm
	service := instance_creation.ForgeService(name, namespace, globalOwnerReference)
	if err = instance_creation.CreateOrUpdate(ctx, r.Client, service); err != nil {
		r.setInstanceStatus(ctx, "Could not create service "+service.Name+" in namespace "+service.Namespace, "Warning", "ServiceNotCreated", instance, "", "")
		return err
	}
	r.setInstanceStatus(ctx, "Service "+service.Name+" correctly created in namespace "+service.Namespace, "Normal", "ServiceCreated", instance, "", "")

	urlUUID := uuid.New().String()
	// create Ingress to manage the service
	ingress := instance_creation.ForgeIngress(name, namespace, &service, urlUUID, r.WebsiteBaseURL, globalOwnerReference)
	if err = instance_creation.CreateOrUpdate(ctx, r.Client, ingress); err != nil {
		r.setInstanceStatus(ctx, "Could not create ingress "+ingress.Name+" in namespace "+ingress.Namespace, "Warning", "IngressNotCreated", instance, "", "")
		return err
	}
	r.setInstanceStatus(ctx, "Ingress "+ingress.Name+" correctly created in namespace "+ingress.Namespace, "Normal", "IngressCreated", instance, "", "")

	if err = r.createOAUTHlogic(name, instance, namespace, globalOwnerReference, urlUUID); err != nil {
		return err
	}

	// create VirtualMachineInstance
	vmi, err := instance_creation.CreateVirtualMachineInstance(name, namespace, environment, instance.Name, secret.Name, globalOwnerReference)
	if err != nil {
		return err
	}
	if err := instance_creation.CreateOrUpdate(ctx, r.Client, vmi); err != nil {
		r.setInstanceStatus(ctx, "Could not create vmi "+vmi.Name+" in namespace "+vmi.Namespace, "Warning", "VmiNotCreated", instance, "", "")
		return err
	}
	r.setInstanceStatus(ctx, "VirtualMachineInstance "+vmi.Name+" correctly created in namespace "+vmi.Namespace, "Normal", "VmiCreated", instance, "", "")
	go r.getVmiStatus(ctx, environment.GuiEnabled, &service, &ingress, instance, vmi, vmStart)
	return nil
}

func (r *InstanceReconciler) createOAUTHlogic(name string, instance *crownlabsv1alpha2.Instance, namespace string, ownerRef []metav1.OwnerReference, urlUUID string) error {
	ctx := context.TODO()

	// create Service for oauth2
	oauthService := instance_creation.ForgeOauth2Service(name, namespace, ownerRef)
	if err := instance_creation.CreateOrUpdate(ctx, r.Client, oauthService); err != nil {
		r.setInstanceStatus(ctx, "Could not create service "+oauthService.Name+" in namespace "+oauthService.Namespace, "Warning", "Oauth2ServiceNotCreated", instance, "", "")
		return err
	}
	r.setInstanceStatus(ctx, "Service "+oauthService.Name+" correctly created in namespace "+oauthService.Namespace, "Normal", "Oauth2ServiceCreated", instance, "", "")

	// create Ingress to manage the oauth2 service
	oauthIngress := instance_creation.ForgeOauth2Ingress(name, namespace, &oauthService, urlUUID, r.WebsiteBaseURL, ownerRef)
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &oauthIngress, func() error {
		r.setInstanceStatus(ctx, "Ingress "+oauthIngress.Name+" correctly created in namespace "+oauthIngress.Namespace, "Normal", "Oauth2IngressCreated", instance, "", "")
		return ctrl.SetControllerReference(instance, &oauthIngress, r.Scheme)
	}); err != nil {
		return err
	}
	r.setInstanceStatus(ctx, "Could not create ingress "+oauthIngress.Name+" in namespace "+oauthIngress.Namespace, "Warning", "Oauth2IngressNotCreated", instance, "", "")

	// create Deployment for oauth2
	oauthDeploy := instance_creation.ForgeOauth2Deployment(name, namespace, urlUUID, r.Oauth2ProxyImage, r.OidcClientSecret, r.OidcProviderURL, ownerRef)
	if err := instance_creation.CreateOrUpdate(ctx, r.Client, oauthDeploy); err != nil {
		r.setInstanceStatus(ctx, "Could not create deployment "+oauthDeploy.Name+" in namespace "+oauthDeploy.Namespace, "Warning", "Oauth2DeployNotCreated", instance, "", "")
		return err
	}
	r.setInstanceStatus(ctx, "Deployment "+oauthDeploy.Name+" correctly created in namespace "+oauthDeploy.Namespace, "Normal", "Oauth2DeployCreated", instance, "", "")

	return nil
}
