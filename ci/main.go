package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"dagger.io/dagger"
)

func NewK8sInstance(ctx context.Context, client *dagger.Client) *K8sInstance {
	return &K8sInstance{
		ctx:         ctx,
		client:      client,
		container:   nil,
		configCache: client.Host().Directory("./kube_config"),
	}
}

type K8sInstance struct {
	ctx         context.Context
	client      *dagger.Client
	container   *dagger.Container
	configCache *dagger.Directory
}

func (k *K8sInstance) start() error {

	k.container = k.client.Container().
		From("bitnami/kubectl").
		WithMountedDirectory("/config/kube", k.configCache).
		WithEnvVariable("CACHE", time.Now().String()).
		WithUser("root").
		WithExec([]string{"cp", "/config/kube/kubeconfig", "/.kube/config"}, dagger.ContainerWithExecOpts{SkipEntrypoint: true}).
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
	retryBackoff := 5 * time.Second
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

func main() {
	ctx := context.Background()

	// create Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	k8s := NewK8sInstance(ctx, client)
	if err = k8s.start(); err != nil {
		panic(err)
	}

	pods, err := k8s.kubectl("get pods -A -o wide")
	if err != nil {
		panic(err)
	}
	fmt.Println(pods)

	services, err := k8s.kubectl("get services -A -o wide")
	if err != nil {
		panic(err)
	}
	fmt.Println(services)

}
