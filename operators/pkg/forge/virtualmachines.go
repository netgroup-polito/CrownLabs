package forge

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	virtv1 "kubevirt.io/client-go/api/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1beta1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	urlDockerPrefix = "docker://"
	// nolint:gosec // The constant refers to the name of a secret, and it is not a secret itself.
	cdiSecretName = "registry-credentials-cdi"
)

// VMReadinessProbe forges the readiness probe for a given VM environment.
func VMReadinessProbe(environment *clv1alpha2.Environment) *virtv1.Probe {
	port := ServiceSSHPort
	if environment.GuiEnabled {
		port = ServiceNoVNCPort
	}

	return &virtv1.Probe{
		InitialDelaySeconds: 10,
		PeriodSeconds:       2,
		FailureThreshold:    5,
		Handler: virtv1.Handler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt(port),
			},
		},
	}
}

// DataVolumeTemplate forges the DataVolume template associated with a given environment.
func DataVolumeTemplate(name string, environment *clv1alpha2.Environment) virtv1.DataVolumeTemplateSpec {
	return virtv1.DataVolumeTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: cdiv1beta1.DataVolumeSpec{
			Source: cdiv1beta1.DataVolumeSource{
				Registry: &cdiv1beta1.DataVolumeSourceRegistry{
					URL:       urlDockerPrefix + environment.Image,
					SecretRef: cdiSecretName,
				},
			},
			PVC: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: environment.Resources.Disk,
					},
				},
			},
		},
	}
}
