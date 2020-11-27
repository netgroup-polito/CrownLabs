package instance_creation

import (
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetWebdavCredentials(c client.Client, ctx context.Context, log logr.Logger, secretName string, namespace string, username *string, password *string) error {
	sec := corev1.Secret{}
	nsdName := types.NamespacedName{
		Namespace: namespace,
		Name:      secretName,
	}
	if err := c.Get(ctx, nsdName, &sec); err == nil {
		var ok bool
		var userBytes, passBytes []byte
		if userBytes, ok = sec.Data["username"]; !ok {
			log.Error(nil, "Unable to find username in webdav secret "+secretName)
		} else {
			*username = string(userBytes)
		}
		if passBytes, ok = sec.Data["password"]; !ok {
			log.Error(nil, "Unable to find password in webdav secret"+secretName)
		} else {
			*password = string(passBytes)
		}
		return nil
	} else {
		return err
	}
}
