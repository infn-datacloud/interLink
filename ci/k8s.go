package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dagger.io/dagger"
)

func NewK8sInstance(ctx context.Context, client *dagger.Client) *K8sInstance {
	return &K8sInstance{
		ctx:         ctx,
		client:      client,
		container:   nil,
		configCache: client.CacheVolume("k3s_config"),
	}
}

type K8sInstance struct {
	ctx         context.Context
	client      *dagger.Client
	container   *dagger.Container
	configCache *dagger.CacheVolume
}

func (k *K8sInstance) start() error {
	// create k3s service container
	k3s := k.client.Pipeline("k3s init").Container().
		From("rancher/k3s").
		WithMountedCache("/etc/rancher/k3s", k.configCache).
		WithMountedTemp("/etc/lib/cni").
		WithMountedTemp("/var/lib/kubelet").
		WithMountedTemp("/var/lib/rancher/k3s").
		WithMountedTemp("/var/log").
		WithEntrypoint([]string{"sh", "-c"}).
		WithExec([]string{"k3s server --disable traefik --disable metrics-server"}, dagger.ContainerWithExecOpts{InsecureRootCapabilities: true}).
		WithExposedPort(6443).AsService()

	k.container = k.client.Container().
		From("bitnami/kubectl").
		WithMountedCache("/cache/k3s", k.configCache).
		WithServiceBinding("k3s", k3s).
		WithEnvVariable("CACHE", time.Now().String()).
		WithUser("root").
		WithExec([]string{"cp", "/cache/k3s/k3s.yaml", "/.kube/config"}, dagger.ContainerWithExecOpts{SkipEntrypoint: true}).
		WithExec([]string{"chown", "1001:0", "/.kube/config"}, dagger.ContainerWithExecOpts{SkipEntrypoint: true}).
		WithUser("1001").
		WithEntrypoint([]string{"sh", "-c"})

	if err := k.waitForNodes(); err != nil {
		return fmt.Errorf("failed to start k8s: %v", err)
	}
	return nil
}

func (k *K8sInstance) kubectl(command string) (string, error) {
	return k.exec("kubectl", fmt.Sprintf("kubectl %v", command))
}

func (k *K8sInstance) exec(name, command string) (string, error) {
	return k.container.Pipeline(name).Pipeline(command).
		WithEnvVariable("CACHE", time.Now().String()).
		WithExec([]string{command}).
		Stdout(k.ctx)
}

func (k *K8sInstance) waitForNodes() (err error) {
	maxRetries := 5
	retryBackoff := 30 * time.Second
	for i := 0; i < maxRetries; i++ {
		time.Sleep(retryBackoff)
		kubectlGetNodes, err := k.kubectl("get nodes -o wide")
		if err != nil {
			fmt.Println(fmt.Errorf("could not fetch nodes: %v", err))
			continue
		}
		if strings.Contains(kubectlGetNodes, "Ready") {
			return nil
		}
		fmt.Println("waiting for k8s to start:", kubectlGetNodes)
	}
	return fmt.Errorf("k8s took too long to start")
}
