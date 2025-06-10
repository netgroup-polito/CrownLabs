package forge

import (
	"fmt"
	"os"
	"os/exec"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"k8s.io/client-go/tools/clientcmd"
)

func insertKubeConfig(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, host string) error {

	cluster := environment.Cluster
	path := fmt.Sprintf("./kubeconfigs/%s-cluster.kubeconfig", cluster.Name)

	cmd := exec.Command(
		"clusterctl", "get", "kubeconfig", fmt.Sprintf("%s-cluster", cluster.Name),
		"--namespace", instance.Namespace,
	)

	raw, _ := cmd.Output()

	cfg, _ := clientcmd.Load(raw)

	newURL := fmt.Sprintf("https://%s:%s",
		host, environment.Cluster.ClusterNet.NginxPort)

	for _, c := range cfg.Clusters {
		c.Server = newURL
	}

	updated, _ := clientcmd.Write(*cfg)

	return os.WriteFile(path, updated, 0o600)

}
func Insinstallcni(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, host string) error {
	cluster := environment.Cluster
	clustername := fmt.Sprintf("%s-cluster", cluster.Name)
	namespace := instance.Namespace
	cni := cluster.ClusterNet.Cni
	podCIDR := cluster.ClusterNet.Pods
	kubeconfigPath := fmt.Sprintf("./kubeconfigs/%s-cluster.kubeconfig", cluster.Name)
	// "Waiting for cluster to be ready"
	exec.Command(
		"kubectl", "wait",
		"--for=condition=Ready=true",
		"-n", namespace,
		fmt.Sprintf("clusters.cluster.x-k8s.io/%s", clustername),
		"--timeout=2m",
	)
	// insert relative KUBECONFIG files into local folder ./kubeconfigs
	insertKubeConfig(instance, environment, host)
	//Installing CNI on cluster
	switch cni {
	case clv1alpha2.CniCalico:

	case clv1alpha2.CniCilium:
		installCilium(kubeconfigPath, podCIDR)
		waitCilium(kubeconfigPath)
	case clv1alpha2.CniFlannel:

	}
	return nil
}

func installCilium(kubeconfig string, podCIDR string) error {
	args := []string{
		"install",
		"--set", fmt.Sprintf("ipam.operator.clusterPoolIPv4PodCIDRList=%s", podCIDR),
		"--set", "affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key=liqo.io/type",
		"--set", "affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator=DoesNotExist",
		"--set", "encryption.enabled=true",
		"--set", "encryption.type=wireguard",
		"--wait",
	}

	cmd := exec.Command("cilium", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kubeconfig))

	_, _ = cmd.CombinedOutput()

	return nil
}

func waitCilium(kubeconfig string) error {
	cmd := exec.Command("cilium", "status", "--wait")
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kubeconfig))

	_, _ = cmd.CombinedOutput()

	return nil
}
