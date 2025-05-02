package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func DockerNetworkExists(name string) (bool, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHostFromEnv(),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return false, err
	}
	defer cli.Close()

	ctx := context.Background()
	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return false, err
	}

	for _, net := range networks {
		if net.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func DockerCreateNetwork(name, driver string) error {
	cli, err := client.NewClientWithOpts(
		client.WithHostFromEnv(),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	defer cli.Close()

	ctx := context.Background()
	_, err = cli.NetworkCreate(ctx, name, types.NetworkCreate{
		Driver: driver,
	})
	return err
}

func DockerRemoveNetworkIfUnused(name string) error {
	cli, err := client.NewClientWithOpts(
		client.WithHostFromEnv(),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	defer cli.Close()

	ctx := context.Background()
	netInfo, err := cli.NetworkInspect(ctx, name, types.NetworkInspectOptions{})
	if err != nil {
		return nil // probabaly already gone ...
	}

	if len(netInfo.Containers) > 0 {
		fmt.Printf("Netzwerk '%s' wird noch genutzt, nicht löschen\n", name)
		return nil
	}

	fmt.Printf("Entferne ungenutztes externes Netzwerk '%s'\n", name)
	return cli.NetworkRemove(ctx, name)
}
