package commands

import (
	"context"
	"dobby/docker"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/go-connections/nat"
	"github.com/urfave/cli/v2"
)

const KafkaImage = "apache/kafka:3.9.0"

func ManageKafka() *cli.Command {
	return &cli.Command{
		Name:    "kafka",
		Aliases: []string{"ka"},
		Usage:   "Manage Kafka containers",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a Kafka container",
				Action: func(c *cli.Context) error {
					if err := startKafkaContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop a Kafka container",
				Action: func(c *cli.Context) error {
					if err := stopKafkaContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the Kafka container",
				Action: func(c *cli.Context) error {
					if kafkaContainerExists() {
						fmt.Println("✅ kafka container is running")
					} else {
						fmt.Println("❌ kafka container is not running")
					}

					return nil
				},
			},
			{
				Name:  "url",
				Usage: "Get connection strings for Kafka",
				Action: func(c *cli.Context) error {
					fmt.Println("localhost:9092")

					return nil
				},
			},
		},
	}
}

func kafkaContainerExists() bool {
	return docker.GetRunningContainerByImage(KafkaImage) != nil
}

func startKafkaContainer() error {
	if kafkaContainerExists() {
		return errors.New("❌ kafka container is already running")
	}

	portBinding := nat.PortMap{
		"9092/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "9092",
			},
		},
	}

	containerConfig := &container.Config{
		Image: KafkaImage,
		Env: []string{
			"KAFKA_NODE_ID=1",
			"KAFKA_PROCESS_ROLES=broker,controller",
			"KAFKA_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093",
			"KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092",
			"KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
			"KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS=0",
		},
		ExposedPorts: nat.PortSet{
			"9092/tcp": struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBinding,
	}

	dockerClient := docker.Client

	out, err := dockerClient.ImagePull(context.Background(), KafkaImage, image.PullOptions{})

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

	fmt.Println("✅ kafka container started successfully")

	return nil
}

func stopKafkaContainer() error {
	dockerClient := docker.Client
	runningKafkaContainer := docker.GetRunningContainerByImage(KafkaImage)

	if runningKafkaContainer == nil {
		return errors.New("❌ kafka container is not running")
	}

	if err := dockerClient.ContainerStop(context.Background(), runningKafkaContainer.ID, container.StopOptions{}); err != nil {
		return fmt.Errorf("❌ error stopping container: %v", err)
	}

	fmt.Println("✅ kafka container stopped successfully")

	return nil
}
