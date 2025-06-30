package forge

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	ctrl "sigs.k8s.io/controller-runtime"
)

func GetKubeadmClusterName(restConfig *rest.Config, namespace string, name string) (string, error) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return "", nil
	}

	kubeadmConfigMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	clusterConfigStr := kubeadmConfigMap.Data["ClusterConfiguration"]
	clusterConfigMap := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(clusterConfigStr), &clusterConfigMap); err != nil {
		return "", err
	}

	result, ok := clusterConfigMap["clusterName"]
	if !ok {
		return "", errors.Errorf("Field `clusterName` not found in configMap %s/%s", namespace, name)
	}

	clusterName, ok := result.(string)
	if !ok {
		return "", errors.Errorf("Field `clusterName` in configMap %s/%s is %T, not a string", namespace, name, result)
	}

	return clusterName, nil
}

func ConstructInClusterKubeconfig(ctx context.Context, restConfig *rest.Config, namespace string) (*clientcmdapi.Config, error) {
	log := ctrl.LoggerFrom(ctx)

	log.V(2).Info("Constructing kubeconfig file from rest.Config")

	var clusterName string
	// Attempt to get the cluster name from the kubeadm configmap, if it fails then set a default name.
	clusterName, err := GetKubeadmClusterName(restConfig, "kube-system", "kubeadm-config")
	if err != nil {
		log.Error(err, "failed to get cluster name from kubeadm configmap")
		clusterName = "management-cluster"
	}

	userName := "default-user"
	contextName := "default-context"
	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters[clusterName] = &clientcmdapi.Cluster{
		Server: restConfig.Host,
		// Used in regular kubeconfigs.
		CertificateAuthorityData: restConfig.CAData,
		// Used in in-cluster configs.
		CertificateAuthority: restConfig.CAFile,
	}

	contexts := make(map[string]*clientcmdapi.Context)
	contexts[contextName] = &clientcmdapi.Context{
		Cluster:   clusterName,
		Namespace: namespace,
		AuthInfo:  userName,
	}

	authInfos := make(map[string]*clientcmdapi.AuthInfo)
	authInfos[userName] = &clientcmdapi.AuthInfo{
		Token:                 restConfig.BearerToken,
		ClientCertificateData: restConfig.TLSClientConfig.CertData,
		ClientKeyData:         restConfig.TLSClientConfig.KeyData,
	}

	return &clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: contextName,
		AuthInfos:      authInfos,
	}, nil
}

func WriteKubeconfigToFile(ctx context.Context, filePath string, clientConfig clientcmdapi.Config) error {
	log := ctrl.LoggerFrom(ctx)

	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "failed to create directory %s", dir)
		}
	}

	log.V(2).Info("Writing kubeconfig to location", "location", filePath)
	if err := clientcmd.WriteToFile(clientConfig, filePath); err != nil {
		return err
	}

	return nil
}
