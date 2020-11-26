package controllers

import (
	"context"
	"github.com/google/uuid"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instance-creation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func (r *LabInstanceReconciler) CreateEnvironment(labInstance *crownlabsv1alpha2.Instance, labTemplate *crownlabsv1alpha2.Template, namespace string, name string, VMstart time.Time) error {
	var user, password string
	// this is added so that all resources created for this Instance are destroyed when the Instance is deleted
	b := true
	labiOwnerRef := []metav1.OwnerReference{
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
	secret := instance_creation.CreateCloudInitSecret(name, namespace, user, password, r.NextcloudBaseUrl)
	secret.SetOwnerReferences(labiOwnerRef)
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, secret); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create secret "+secret.Name+" in namespace "+secret.Namespace, "Warning", "SecretNotCreated", labInstance, "", "")
	} else {
		setLabInstanceStatus(r, ctx, log, "Secret "+secret.Name+" correctly created in namespace "+secret.Namespace, "Normal", "SecretCreated", labInstance, "", "")
	}

	// create Service to expose the vm
	service := instance_creation.CreateService(name, namespace)
	service.SetOwnerReferences(labiOwnerRef)
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, service); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create service "+service.Name+" in namespace "+service.Namespace, "Warning", "ServiceNotCreated", labInstance, "", "")
		return err
	} else {
		setLabInstanceStatus(r, ctx, log, "Service "+service.Name+" correctly created in namespace "+service.Namespace, "Normal", "ServiceCreated", labInstance, "", "")
	}

	urlUUID := uuid.New().String()
	// create Ingress to manage the service
	ingress := instance_creation.CreateIngress(name, namespace, service, urlUUID, r.WebsiteBaseUrl)
	ingress.SetOwnerReferences(labiOwnerRef)
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, ingress); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create ingress "+ingress.Name+" in namespace "+ingress.Namespace, "Warning", "IngressNotCreated", labInstance, "", "")
		return err
	} else {
		setLabInstanceStatus(r, ctx, log, "Ingress "+ingress.Name+" correctly created in namespace "+ingress.Namespace, "Normal", "IngressCreated", labInstance, "", "")
	}

	if err := r.createOAUTHLogic(name, labInstance, namespace, labiOwnerRef, urlUUID); err != nil {
		return err
	}

	// create VirtualMachineInstance
	vmi, err := instance_creation.CreateVirtualMachineInstance(name, namespace, labTemplate.Spec.EnvironmentList[0], labInstance.Name, secret.Name)
	if err != nil {
		return err
	}
	vmi.SetOwnerReferences(labiOwnerRef)
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, vmi); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create vmi "+vmi.Name+" in namespace "+vmi.Namespace, "Warning", "VmiNotCreated", labInstance, "", "")
		return err
	} else {
		setLabInstanceStatus(r, ctx, log, "VirtualMachineInstance "+vmi.Name+" correctly created in namespace "+vmi.Namespace, "Normal", "VmiCreated", labInstance, "", "")
	}
	go getVmiStatus(r, ctx, log, labTemplate.Spec.EnvironmentList[0].GuiEnabled, service, ingress, labInstance, vmi, VMstart)
	return nil
}

func (r *LabInstanceReconciler) createOAUTHLogic(name string, labInstance *crownlabsv1alpha2.Instance, namespace string, labiOwnerRef []metav1.OwnerReference, urlUUID string) error {
	ctx := context.TODO()
	log := r.Log
	// create Service for oauth2
	oauthService := instance_creation.CreateOauth2Service(name, namespace)
	oauthService.SetOwnerReferences(labiOwnerRef)
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, oauthService); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create service "+oauthService.Name+" in namespace "+oauthService.Namespace, "Warning", "Oauth2ServiceNotCreated", labInstance, "", "")
		return err
	} else {
		setLabInstanceStatus(r, ctx, log, "Service "+oauthService.Name+" correctly created in namespace "+oauthService.Namespace, "Normal", "Oauth2ServiceCreated", labInstance, "", "")
	}

	// create Ingress to manage the oauth2 service
	oauthIngress := instance_creation.CreateOauth2Ingress(name, namespace, oauthService, urlUUID, r.WebsiteBaseUrl)
	oauthIngress.SetOwnerReferences(labiOwnerRef)
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, oauthIngress); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create ingress "+oauthIngress.Name+" in namespace "+oauthIngress.Namespace, "Warning", "Oauth2IngressNotCreated", labInstance, "", "")
		return err
	} else {
		setLabInstanceStatus(r, ctx, log, "Ingress "+oauthIngress.Name+" correctly created in namespace "+oauthIngress.Namespace, "Normal", "Oauth2IngressCreated", labInstance, "", "")
	}

	// create Deployment for oauth2
	oauthDeploy := instance_creation.CreateOauth2Deployment(name, namespace, urlUUID, r.Oauth2ProxyImage, r.OidcClientSecret, r.OidcProviderUrl)
	oauthDeploy.SetOwnerReferences(labiOwnerRef)
	if err := instance_creation.CreateOrUpdate(r.Client, ctx, log, oauthDeploy); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create deployment "+oauthDeploy.Name+" in namespace "+oauthDeploy.Namespace, "Warning", "Oauth2DeployNotCreated", labInstance, "", "")
		return err
	} else {
		setLabInstanceStatus(r, ctx, log, "Deployment "+oauthDeploy.Name+" correctly created in namespace "+oauthDeploy.Namespace, "Normal", "Oauth2DeployCreated", labInstance, "", "")
	}
	return nil
}
