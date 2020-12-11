package instance_controller

import (
	"context"
	"github.com/google/uuid"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instance-creation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

func (r *LabInstanceReconciler) CreateVMEnvironment(labInstance *crownlabsv1alpha2.Instance, environment *crownlabsv1alpha2.Environment, namespace string, name string, vmStart time.Time) error {
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
	log := r.Log
	ctx := context.TODO()
	err := instance_creation.GetWebdavCredentials(r.Client, ctx, log, r.WebdavSecretName, labInstance.Namespace, &user, &password)
	if err != nil {
		log.Error(err, "unable to get Webdav Credentials")
	} else {
		log.Info("Webdav secrets obtained. Building cloud-init script." + labInstance.Name)
	}
	secret := instance_creation.CreateCloudInitSecret(name, namespace, user, password, r.NextcloudBaseUrl, globalOwnerReference)
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, secret); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create secret "+secret.Name+" in namespace "+secret.Namespace, "Warning", "SecretNotCreated", labInstance, "", "")
	} else {
		setLabInstanceStatus(r, ctx, log, "Secret "+secret.Name+" correctly created in namespace "+secret.Namespace, "Normal", "SecretCreated", labInstance, "", "")
	}

	// create Service to expose the vm
	service := instance_creation.CreateService(name, namespace, globalOwnerReference)
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, service); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create service "+service.Name+" in namespace "+service.Namespace, "Warning", "ServiceNotCreated", labInstance, "", "")
		return err
	} else {
		setLabInstanceStatus(r, ctx, log, "Service "+service.Name+" correctly created in namespace "+service.Namespace, "Normal", "ServiceCreated", labInstance, "", "")
	}

	urlUUID := uuid.New().String()
	// create Ingress to manage the service
	ingress := instance_creation.CreateIngress(name, namespace, service, urlUUID, r.WebsiteBaseUrl, globalOwnerReference)
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, ingress); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create ingress "+ingress.Name+" in namespace "+ingress.Namespace, "Warning", "IngressNotCreated", labInstance, "", "")
		return err
	} else {
		setLabInstanceStatus(r, ctx, log, "Ingress "+ingress.Name+" correctly created in namespace "+ingress.Namespace, "Normal", "IngressCreated", labInstance, "", "")
	}

	if err := r.createOAUTHLogic(name, labInstance, namespace, globalOwnerReference, urlUUID); err != nil {
		return err
	}

	// create VirtualMachineInstance
	vmi, err := instance_creation.CreateVirtualMachineInstance(name, namespace, environment, labInstance.Name, secret.Name, globalOwnerReference)
	if err != nil {
		return err
	}
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, vmi); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create vmi "+vmi.Name+" in namespace "+vmi.Namespace, "Warning", "VmiNotCreated", labInstance, "", "")
		return err
	} else {
		setLabInstanceStatus(r, ctx, log, "VirtualMachineInstance "+vmi.Name+" correctly created in namespace "+vmi.Namespace, "Normal", "VmiCreated", labInstance, "", "")
	}
	go getVmiStatus(r, ctx, log, environment.GuiEnabled, service, ingress, labInstance, vmi, vmStart)
	return nil
}

func (r *LabInstanceReconciler) createOAUTHLogic(name string, labInstance *crownlabsv1alpha2.Instance, namespace string, labiOwnerRef []metav1.OwnerReference, urlUUID string) error {
	ctx := context.TODO()
	log := r.Log

	// create Service for oauth2
	oauthService := instance_creation.CreateOauth2Service(name, namespace, labiOwnerRef)
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, oauthService); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create service "+oauthService.Name+" in namespace "+oauthService.Namespace, "Warning", "Oauth2ServiceNotCreated", labInstance, "", "")
		return err
	} else {
		setLabInstanceStatus(r, ctx, log, "Service "+oauthService.Name+" correctly created in namespace "+oauthService.Namespace, "Normal", "Oauth2ServiceCreated", labInstance, "", "")
	}

	// create Ingress to manage the oauth2 service
	oauthIngress := instance_creation.CreateOauth2Ingress(name, namespace, oauthService, urlUUID, r.WebsiteBaseUrl, labiOwnerRef)
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &oauthIngress, func() error {
		return ctrl.SetControllerReference(labInstance, &oauthIngress, r.Scheme)
	}); err != nil {
		return err
	}
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, oauthIngress); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create ingress "+oauthIngress.Name+" in namespace "+oauthIngress.Namespace, "Warning", "Oauth2IngressNotCreated", labInstance, "", "")
		return err
	} else {
		setLabInstanceStatus(r, ctx, log, "Ingress "+oauthIngress.Name+" correctly created in namespace "+oauthIngress.Namespace, "Normal", "Oauth2IngressCreated", labInstance, "", "")
	}

	// create Deployment for oauth2
	oauthDeploy := instance_creation.CreateOauth2Deployment(name, namespace, urlUUID, r.Oauth2ProxyImage, r.OidcClientSecret, r.OidcProviderUrl, labiOwnerRef)
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, oauthDeploy); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create deployment "+oauthDeploy.Name+" in namespace "+oauthDeploy.Namespace, "Warning", "Oauth2DeployNotCreated", labInstance, "", "")
		return err
	} else {
		setLabInstanceStatus(r, ctx, log, "Deployment "+oauthDeploy.Name+" correctly created in namespace "+oauthDeploy.Namespace, "Normal", "Oauth2DeployCreated", labInstance, "", "")
	}
	return nil
}
