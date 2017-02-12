package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"io"
	"os"
)

func main() {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		reader, err := cli.ImagePull(context.Background(), container.Image, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		io.Copy(os.Stdout, reader)

		config, err := cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			panic(err)
		}

		isRunning := container.State == "running"

		networkConfig := network.NetworkingConfig{
			EndpointsConfig: config.NetworkSettings.Networks,
		}

		err = cli.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			panic(err)
		}

		container, err := cli.ContainerCreate(context.Background(), config.Config, config.HostConfig, &networkConfig, config.Name)
		if err != nil {
			panic(err)
		}
		fmt.Println(container)

		if isRunning {
			err = cli.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
			if err != nil {
				panic(err)
			}
		}
	}

	fmt.Println("done")
}
