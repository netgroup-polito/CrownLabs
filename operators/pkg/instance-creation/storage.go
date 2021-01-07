package instance_creation

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetWebdavCredentials(ctx context.Context, c client.Client, secretName, namespace string, username, password *string) error {
	sec := corev1.Secret{}
	nsdName := types.NamespacedName{
		Namespace: namespace,
		Name:      secretName,
	}
	if err := c.Get(ctx, nsdName, &sec); err != nil {
		return err
	}

	var ok bool
	var userBytes, passBytes []byte
	if userBytes, ok = sec.Data["username"]; !ok {
		klog.Error(nil, "Unable to find username in webdav secret "+secretName)
	} else {
		*username = string(userBytes)
	}
	if passBytes, ok = sec.Data["password"]; !ok {
		klog.Error(nil, "Unable to find password in webdav secret"+secretName)
	} else {
		*password = string(passBytes)
	}
	return nil
}
