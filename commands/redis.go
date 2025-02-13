package commands

import (
	"context"
	"dobby/docker"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/go-connections/nat"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
)

const RedisImage = "redis:7.4"

func ManageRedis() *cli.Command {
	return &cli.Command{
		Name:    "redis",
		Aliases: []string{"r"},
		Usage:   "Manage Redis containers",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a Redis container",
				Action: func(c *cli.Context) error {
					if err := startRedisContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop a Redis container",
				Action: func(c *cli.Context) error {
					if err := stopRedisContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the Redis container",
				Action: func(c *cli.Context) error {
					if redisContainerExists() {
						fmt.Println("✅ redis container is running")
					} else {
						fmt.Println("❌ redis container is not running")
					}

					return nil
				},
			},
			{
				Name:  "url",
				Usage: "Get connection strings for Redis",
				Action: func(c *cli.Context) error {
					fmt.Println("redis://localhost:6379")

					return nil
				},
			},
		},
	}
}

func redisContainerExists() bool {
	return docker.GetRunningContainerByImage(RedisImage) != nil
}

func startRedisContainer() error {
	if redisContainerExists() {
		return errors.New("❌ redis container is already running")
	}

	portBinding := nat.PortMap{
		"6379/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "6379",
			},
		},
	}

	containerConfig := &container.Config{
		Image: RedisImage,
		ExposedPorts: nat.PortSet{
			"6379/tcp": struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBinding,
	}

	dockerClient := docker.Client

	out, err := dockerClient.ImagePull(context.Background(), RedisImage, image.PullOptions{})

	if err != nil {
		return fmt.Errorf("❌ error pulling image: %v", err)
	}

	defer func(out io.ReadCloser) {
		err := out.Close()
		if err != nil {
			log.Fatalf("❌ error closing image pull response: %v", err)
		}
	}(out)

	_, err = io.Copy(os.Stdout, out)
	if err != nil {
		return fmt.Errorf("❌ error copying image pull response: %v", err)
	}

	resp, err := dockerClient.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, "")

	if err != nil {
		return fmt.Errorf("❌ error creating container: %v", err)
	}

	if err = dockerClient.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("❌ error starting container: %v", err)
	}

	fmt.Println("✅ redis container started successfully")

	return nil
}

func stopRedisContainer() error {
	dockerClient := docker.Client
	runningRedisContainer := docker.GetRunningContainerByImage(RedisImage)

	if runningRedisContainer == nil {
		return errors.New("❌ redis container is not running")
	}

	if err := dockerClient.ContainerStop(context.Background(), runningRedisContainer.ID, container.StopOptions{}); err != nil {
		return fmt.Errorf("❌ error stopping container: %v", err)
	}

	fmt.Println("✅ redis container stopped successfully")

	return nil
}
