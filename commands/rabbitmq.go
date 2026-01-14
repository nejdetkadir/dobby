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

const RabbitMQImage = "rabbitmq:3-management"

func ManageRabbitMQ() *cli.Command {
	return &cli.Command{
		Name:    "rabbitmq",
		Aliases: []string{"rmq"},
		Usage:   "Manage RabbitMQ containers",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a RabbitMQ container",
				Action: func(c *cli.Context) error {
					if err := startRabbitMQContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop a RabbitMQ container",
				Action: func(c *cli.Context) error {
					if err := stopRabbitMQContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the RabbitMQ container",
				Action: func(c *cli.Context) error {
					if rabbitMQContainerExists() {
						fmt.Println("✅ rabbitmq container is running")
					} else {
						fmt.Println("❌ rabbitmq container is not running")
					}

					return nil
				},
			},
			{
				Name:  "url",
				Usage: "Get connection strings for RabbitMQ",
				Action: func(c *cli.Context) error {
					fmt.Println("amqp://admin:admin123@localhost:5672/")
					fmt.Println("Management UI: http://localhost:15672")

					return nil
				},
			},
		},
	}
}

func rabbitMQContainerExists() bool {
	return docker.GetRunningContainerByImage(RabbitMQImage) != nil
}

func startRabbitMQContainer() error {
	if rabbitMQContainerExists() {
		return errors.New("❌ rabbitmq container is already running")
	}

	portBinding := nat.PortMap{
		"5672/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "5672",
			},
		},
		"15672/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "15672",
			},
		},
	}

	containerConfig := &container.Config{
		Image: RabbitMQImage,
		Env: []string{
			"RABBITMQ_DEFAULT_USER=admin",
			"RABBITMQ_DEFAULT_PASS=admin123",
		},
		ExposedPorts: nat.PortSet{
			"5672/tcp":  struct{}{},
			"15672/tcp": struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBinding,
	}

	dockerClient := docker.Client

	out, err := dockerClient.ImagePull(context.Background(), RabbitMQImage, image.PullOptions{})

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

	fmt.Println("✅ rabbitmq container started successfully")

	return nil
}

func stopRabbitMQContainer() error {
	dockerClient := docker.Client
	runningRabbitMQContainer := docker.GetRunningContainerByImage(RabbitMQImage)

	if runningRabbitMQContainer == nil {
		return errors.New("❌ rabbitmq container is not running")
	}

	if err := dockerClient.ContainerStop(context.Background(), runningRabbitMQContainer.ID, container.StopOptions{}); err != nil {
		return fmt.Errorf("❌ error stopping container: %v", err)
	}

	fmt.Println("✅ rabbitmq container stopped successfully")

	return nil
}
