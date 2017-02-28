package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

func worker(config *WorkerConfig) {
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

type WorkerConfig struct {
	Policy   string
	Notify   bool
	Rollback bool
	Intranet bool
}

func main() {
	policy := flag.String("policy", "auto", "policy updating containers (auto, hook, web)")
	notify := flag.Bool("notify", false, "notify about fail")
	rollback := flag.Bool("rollback", false, "rollback after fail (only for containers which was runned)")
	intranet := flag.Bool("intranet", false, "puller's network")

	config := WorkerConfig{*policy, *notify, *rollback, *intranet}

	if config.Policy == "auto" {
		ticker := time.NewTicker(600 * time.Second)
		quit := make(chan struct{})
		for {
			select {
			case <-ticker.C:
				worker(&config)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}
}
