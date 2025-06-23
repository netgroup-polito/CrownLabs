package forge

import (
	"context"
	"fmt"
	"os/exec"

	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	ctrl "sigs.k8s.io/controller-runtime"
)

// cluster visualization

func ClusterVisulizer(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	namespace := instance.Namespace

	port := environment.Visulizer.VisulizerPort
	if port == "" {
		port = "8082"
	}
	name := fmt.Sprintf("%s-instance-visualizer", instance.Name)
	// Helm upgrade --install
	cmd := exec.Command(
		"helm", "upgrade", "--install",
		name, "cluster-api-visualizer",
		"--repo", "https://jont828.github.io/cluster-api-visualizer/charts",
		"-n", namespace,
	)
	if err := cmd.Run(); err != nil {
		log.Error(err, "Helm upgrade --install failed")
		return err
	}
	log.Info("Helm upgrade --install succeeded")

	// Rollout status
	cmd = exec.CommandContext(
		ctx,
		"kubectl", "rollout", "status",
		"deployment", "capi-visualizer",
		"-n", namespace,
	)
	log.Info("Waiting for rollout", "deployment", "capi-visualizer", "namespace", namespace)
	if err := cmd.Run(); err != nil {
		log.Error(err, "Rollout failed")
		return err
	}

	// Start port-forward (non-blocking)
	pp := fmt.Sprintf("%s:%s", port, "8081")
	cmd = exec.CommandContext(
		ctx,
		"kubectl", "port-forward",
		"-n", namespace,
		"service/capi-visualizer",
		pp,
	)
	log.Info("Starting port-forward", "host", "localhost", "port", port)
	if err := cmd.Start(); err != nil {
		log.Error(err, "kubectl port-forward failed to start")
		return err
	}
	return nil
}
