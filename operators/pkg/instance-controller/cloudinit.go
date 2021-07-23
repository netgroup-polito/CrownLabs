package instance_controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	instance_creation "github.com/netgroup-polito/CrownLabs/operators/pkg/instance-creation"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

const (
	webdavSecretUsernameKey = "username"
	webdavSecretPasswordKey = "password"
)

// EnforceCloudInitSecret enforces the creation/update of a secret containing the cloud-init configuration,
// based on the information retrieved for the tenant object and its associated WebDav credentials.
func (r *InstanceReconciler) EnforceCloudInitSecret(ctx context.Context, instance *crownlabsv1alpha2.Instance) error {
	log := ctrl.LoggerFrom(ctx)

	// Retrieve the WebDav credentials.
	namespacedName := forge.NamespaceName(instance)
	secretName := types.NamespacedName{Namespace: namespacedName.Namespace, Name: r.WebdavSecretName}
	user, password, err := r.getWebDavCredentials(ctx, secretName)
	if err != nil {
		log.Error(err, "unable to get webdav credentials", "secret", secretName)
		return err
	}
	log.V(utils.LogDebugLevel).Info("webdav credentials correctly retrieved", "secret", secretName)

	// Retrieve the public keys
	var publicKeys []string
	if err = instance_creation.GetPublicKeys(ctx, r.Client, instance.Spec.Tenant, instance.Spec.Template, &publicKeys); err != nil {
		log.Error(err, "unable to get public keys")
		return err
	}
	log.V(utils.LogDebugLevel).Info("public keys correctly retrieved")

	// Create the cloud-init secret
	secret := instance_creation.CreateCloudInitSecret(namespacedName.Name, namespacedName.Namespace, user, password, r.NextcloudBaseURL, publicKeys)
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &secret, func() error {
		return ctrl.SetControllerReference(instance, &secret, r.Scheme)
	})
	if err != nil {
		log.Error(err, "failed to enforce cloud-init secret", "secret", klog.KObj(&secret))
		return err
	}

	log.V(utils.FromResult(res)).Info("cloud-init secret enforced", "secret", klog.KObj(&secret), "result", res)
	return nil
}

// getWebDavCredentials extracts the credentials (i.e. username and password)
// required to mount the MyDrive disk of a given tenant from the associated
// secret.
func (r *InstanceReconciler) getWebDavCredentials(ctx context.Context, secretName types.NamespacedName) (username, password string, err error) {
	secret := corev1.Secret{}
	if err = r.Get(ctx, secretName, &secret); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed to retrieve secret", "secret", secretName)
		return
	}

	var ok bool
	var userBytes, passBytes []byte

	if userBytes, ok = secret.Data[webdavSecretUsernameKey]; !ok {
		err = fmt.Errorf("cannot find %v key in secret", webdavSecretUsernameKey)
		ctrl.LoggerFrom(ctx).Error(err, "failed to retrieve credentials from secret", "secret", secretName)
		return
	}

	if passBytes, ok = secret.Data[webdavSecretPasswordKey]; !ok {
		err = fmt.Errorf("cannot find %v key in secret", webdavSecretPasswordKey)
		ctrl.LoggerFrom(ctx).Error(err, "failed to retrieve credentials from secret", "secret", secretName)
		return
	}

	return string(userBytes), string(passBytes), nil
}
