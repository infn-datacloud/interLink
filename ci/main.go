package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	"github.com/containerd/containerd/log"
	"github.com/sirupsen/logrus"
)

func BuildImage(ctx context.Context, client *dagger.Client, dockerfile string, imageName string) (string, error) {

	contextDir := client.Host().Directory(".")

	log.G(ctx).Info(fmt.Sprintf("Building image %s from %s", imageName, dockerfile))
	return contextDir.
		DockerBuild(dagger.DirectoryDockerBuildOpts{
			Dockerfile: dockerfile,
		}).
		Publish(ctx, fmt.Sprintf(imageName))
}

func main() {
	ctx := context.Background()
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.DebugLevel)
	log.L = logger.WithContext(ctx)
	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Build VK
	//ref, err := BuildImage(ctx, client, "./docker/Dockerfile.vk", "dciangot/vk:latest")
	//fmt.Printf("Published image to :%s\n", ref)

	// Initialize k3s server
	k8s := NewK8sInstance(ctx, client)
	if err = k8s.start(); err != nil {
		panic(err)
	}

	pods, err := k8s.kubectl("get pods -A -o wide")
	if err != nil {
		panic(err)
	}
	fmt.Println(pods)
}
