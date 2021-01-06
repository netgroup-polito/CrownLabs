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

func (r *LabInstanceReconciler) CreateVMEnvironment(labInstance *crownlabsv1alpha2.Instance, environment *crownlabsv1alpha2.Environment, namespace, name string, vmStart time.Time) error {
	var user, password string
	// this is added so that all resources created for this Instance are destroyed when the Instance is deleted
	b := true
	globalOwnerReference := []metav1.OwnerReference{
		{
			APIVersion:         labInstance.APIVersion,
			Kind:               labInstance.Kind,
			Name:               labInstance.Name,
			UID:                labInstance.UID,
			BlockOwnerDeletion: &b,
		},
	}
	ctx := context.TODO()
	err := instance_creation.GetWebdavCredentials(ctx, r.Client, r.WebdavSecretName, labInstance.Namespace, &user, &password)
	if err != nil {
		klog.Error("unable to get Webdav Credentials")
		klog.Error(err)
	} else {
		klog.Info("Webdav secrets obtained. Getting public keys. " + labInstance.Name)
	}
	var publicKeys []string
	if err = instance_creation.GetPublicKeys(ctx, r.Client, labInstance.Spec.Tenant, labInstance.Spec.Template, &publicKeys); err != nil {
		klog.Error("unable to get public keys")
		klog.Error(err)
	} else {
		klog.Info("Public keys obtained. Building cloud-init script. " + labInstance.Name)
	}
	secret := instance_creation.CreateCloudInitSecret(name, namespace, user, password, r.NextcloudBaseURL, publicKeys, globalOwnerReference)
	if err = instance_creation.CreateOrUpdate(ctx, r.Client, secret); err != nil {
		r.setLabInstanceStatus(ctx, "Could not create secret "+secret.Name+" in namespace "+secret.Namespace, "Warning", "SecretNotCreated", labInstance, "", "")
	} else {
		r.setLabInstanceStatus(ctx, "Secret "+secret.Name+" correctly created in namespace "+secret.Namespace, "Normal", "SecretCreated", labInstance, "", "")
	}

	// create Service to expose the vm
	service := instance_creation.ForgeService(name, namespace, globalOwnerReference)
	if err = instance_creation.CreateOrUpdate(ctx, r.Client, service); err != nil {
		r.setLabInstanceStatus(ctx, "Could not create service "+service.Name+" in namespace "+service.Namespace, "Warning", "ServiceNotCreated", labInstance, "", "")
		return err
	}
	r.setLabInstanceStatus(ctx, "Service "+service.Name+" correctly created in namespace "+service.Namespace, "Normal", "ServiceCreated", labInstance, "", "")

	urlUUID := uuid.New().String()
	// create Ingress to manage the service
	ingress := instance_creation.ForgeIngress(name, namespace, &service, urlUUID, r.WebsiteBaseURL, globalOwnerReference)
	if err = instance_creation.CreateOrUpdate(ctx, r.Client, ingress); err != nil {
		r.setLabInstanceStatus(ctx, "Could not create ingress "+ingress.Name+" in namespace "+ingress.Namespace, "Warning", "IngressNotCreated", labInstance, "", "")
		return err
	}
	r.setLabInstanceStatus(ctx, "Ingress "+ingress.Name+" correctly created in namespace "+ingress.Namespace, "Normal", "IngressCreated", labInstance, "", "")

	if err = r.createOAUTHlogic(name, labInstance, namespace, globalOwnerReference, urlUUID); err != nil {
		return err
	}

	// create VirtualMachineInstance
	vmi, err := instance_creation.CreateVirtualMachineInstance(name, namespace, environment, labInstance.Name, secret.Name, globalOwnerReference)
	if err != nil {
		return err
	}
	if err := instance_creation.CreateOrUpdate(ctx, r.Client, vmi); err != nil {
		r.setLabInstanceStatus(ctx, "Could not create vmi "+vmi.Name+" in namespace "+vmi.Namespace, "Warning", "VmiNotCreated", labInstance, "", "")
		return err
	}
	r.setLabInstanceStatus(ctx, "VirtualMachineInstance "+vmi.Name+" correctly created in namespace "+vmi.Namespace, "Normal", "VmiCreated", labInstance, "", "")
	go r.getVmiStatus(ctx, environment.GuiEnabled, &service, &ingress, labInstance, vmi, vmStart)
	return nil
}

func (r *LabInstanceReconciler) createOAUTHlogic(name string, labInstance *crownlabsv1alpha2.Instance, namespace string, labiOwnerRef []metav1.OwnerReference, urlUUID string) error {
	ctx := context.TODO()

	// create Service for oauth2
	oauthService := instance_creation.ForgeOauth2Service(name, namespace, labiOwnerRef)
	if err := instance_creation.CreateOrUpdate(ctx, r.Client, oauthService); err != nil {
		r.setLabInstanceStatus(ctx, "Could not create service "+oauthService.Name+" in namespace "+oauthService.Namespace, "Warning", "Oauth2ServiceNotCreated", labInstance, "", "")
		return err
	}
	r.setLabInstanceStatus(ctx, "Service "+oauthService.Name+" correctly created in namespace "+oauthService.Namespace, "Normal", "Oauth2ServiceCreated", labInstance, "", "")

	// create Ingress to manage the oauth2 service
	oauthIngress := instance_creation.ForgeOauth2Ingress(name, namespace, &oauthService, urlUUID, r.WebsiteBaseURL, labiOwnerRef)
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &oauthIngress, func() error {
		r.setLabInstanceStatus(ctx, "Ingress "+oauthIngress.Name+" correctly created in namespace "+oauthIngress.Namespace, "Normal", "Oauth2IngressCreated", labInstance, "", "")
		return ctrl.SetControllerReference(labInstance, &oauthIngress, r.Scheme)
	}); err != nil {
		return err
	}
	r.setLabInstanceStatus(ctx, "Could not create ingress "+oauthIngress.Name+" in namespace "+oauthIngress.Namespace, "Warning", "Oauth2IngressNotCreated", labInstance, "", "")

	// create Deployment for oauth2
	oauthDeploy := instance_creation.ForgeOauth2Deployment(name, namespace, urlUUID, r.Oauth2ProxyImage, r.OidcClientSecret, r.OidcProviderURL, labiOwnerRef)
	if err := instance_creation.CreateOrUpdate(ctx, r.Client, oauthDeploy); err != nil {
		r.setLabInstanceStatus(ctx, "Could not create deployment "+oauthDeploy.Name+" in namespace "+oauthDeploy.Namespace, "Warning", "Oauth2DeployNotCreated", labInstance, "", "")
		return err
	}
	r.setLabInstanceStatus(ctx, "Deployment "+oauthDeploy.Name+" correctly created in namespace "+oauthDeploy.Namespace, "Normal", "Oauth2DeployCreated", labInstance, "", "")

	return nil
}
