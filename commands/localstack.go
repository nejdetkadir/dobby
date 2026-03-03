package commands

import (
	"context"
	"dobby/docker"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/urfave/cli/v2"
)

const LocalStackImage = "localstack/localstack"

func ManageLocalStack() *cli.Command {
	return &cli.Command{
		Name:    "localstack",
		Aliases: []string{"ls"},
		Usage:   "Manage LocalStack containers",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a LocalStack container",
				Action: func(c *cli.Context) error {
					if err := startLocalStackContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop a LocalStack container",
				Action: func(c *cli.Context) error {
					if err := stopLocalStackContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the LocalStack container",
				Action: func(c *cli.Context) error {
					if localStackContainerExists() {
						fmt.Println("✅ localstack container is running")
					} else {
						fmt.Println("❌ localstack container is not running")
					}

					return nil
				},
			},
			{
				Name:  "url",
				Usage: "Get connection strings for LocalStack",
				Action: func(c *cli.Context) error {
					fmt.Println("http://localhost:4566")

					return nil
				},
			},
		},
	}
}

func localStackContainerExists() bool {
	return docker.GetRunningContainerByImage(LocalStackImage) != nil
}

func startLocalStackContainer() error {
	if localStackContainerExists() {
		return errors.New("❌ localstack container is already running")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("❌ error getting home directory: %v", err)
	}

	localStackDataDir := filepath.Join(homeDir, "docker_volumes", "localstack_data")
	if err := os.MkdirAll(localStackDataDir, 0755); err != nil {
		return fmt.Errorf("❌ error creating data directory: %v", err)
	}

	portBinding := nat.PortMap{
		"4566/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "4566",
			},
		},
	}

	containerConfig := &container.Config{
		Image: LocalStackImage,
		Env: []string{
			"SERVICES=s3,sqs",
			"DEBUG=1",
		},
		ExposedPorts: nat.PortSet{
			"4566/tcp": struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBinding,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: localStackDataDir,
				Target: "/var/lib/localstack",
			},
			{
				Type:   mount.TypeBind,
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
			},
		},
	}

	dockerClient := docker.Client

	out, err := dockerClient.ImagePull(context.Background(), LocalStackImage, image.PullOptions{})

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

	fmt.Println("✅ localstack container started successfully")

	return nil
}

func stopLocalStackContainer() error {
	dockerClient := docker.Client
	runningLocalStackContainer := docker.GetRunningContainerByImage(LocalStackImage)

	if runningLocalStackContainer == nil {
		return errors.New("❌ localstack container is not running")
	}

	if err := dockerClient.ContainerStop(context.Background(), runningLocalStackContainer.ID, container.StopOptions{}); err != nil {
		return fmt.Errorf("❌ error stopping container: %v", err)
	}

	fmt.Println("✅ localstack container stopped successfully")

	return nil
}
