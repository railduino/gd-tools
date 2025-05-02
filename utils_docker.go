package main

import (
	"context"
	"time"

	"github.com/docker/docker/client"
)

func IsDockerAvailable(timeout time.Duration) bool {
	cli, err := client.NewClientWithOpts(
		client.WithHostFromEnv(),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return false
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = cli.Ping(ctx)
	return err == nil
}

func AnyContainerRunning(names []string, timeout time.Duration) (bool, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHostFromEnv(),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return false, err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, name := range names {
		container, err := cli.ContainerInspect(ctx, name)
		if err != nil {
			if client.IsErrNotFound(err) {
				continue
			}
			return false, err
		}

		if container.State != nil && container.State.Running {
			return true, nil
		}
	}

	return false, nil
}

func AllContainersRunning(names []string, timeout time.Duration) (bool, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHostFromEnv(),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return false, err
	}
	defer cli.Close()

	deadline := time.Now().Add(timeout)
	ctx := context.Background()

	for time.Now().Before(deadline) {
		allRunning := true

		for _, name := range names {
			container, err := cli.ContainerInspect(ctx, name)
			if err != nil {
				if client.IsErrNotFound(err) {
					allRunning = false
					break
				}
				return false, err
			}

			if container.State == nil || !container.State.Running {
				allRunning = false
				break
			}
		}

		if allRunning {
			return true, nil
		}

		time.Sleep(1 * time.Second)
	}

	return false, nil // nicht alle liefen innerhalb des Timeouts
}

func AllContainersStopped(names []string) (bool, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHostFromEnv(),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return false, err
	}
	defer cli.Close()

	ctx := context.Background()

	for _, name := range names {
		container, err := cli.ContainerInspect(ctx, name)
		if err != nil {
			if client.IsErrNotFound(err) {
				continue // Container existiert nicht → okay
			}
			return false, err
		}

		if container.State != nil && container.State.Running {
			return false, nil // mindestens einer läuft
		}
	}

	return true, nil
}
