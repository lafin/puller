package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"io/ioutil"
	"strings"
)

func main() {
	policyUpdating := flag.String("policy", "all", "policy of updateting container (all, running)")
	// autoUpdating := flag.Bool("auto", false, "auto update by schedule (checking every 10 min)")
	// webUI := flag.Bool("web", false, "run web ui")
	// hookUpdating := flag.Bool("hook", false, "update by hook")
	// notify := flag.Bool("notify", false, "notify about fail")
	// rollback := flag.Bool("rollback", false, "rollback after fail (only for containers which was runned)")
	// internalNetwork := flag.Bool("network", false, "puller's network")

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
		wasRunned := container.State == "running"
		if *policyUpdating == "running" && !wasRunned {
			continue
		}

		reader, err := cli.ImagePull(context.Background(), container.Image, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		output, err := ioutil.ReadAll(reader)
		if err != nil {
			panic(err)
		}
		if strings.Contains(string(output), "Image is up to date") {
			continue
		}

		config, err := cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			panic(err)
		}
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

		if wasRunned {
			err = cli.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
			if err != nil {
				panic(err)
			}
		}
	}

	fmt.Println("done")
}
