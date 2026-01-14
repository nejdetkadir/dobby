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

const ElasticsearchImage = "docker.elastic.co/elasticsearch/elasticsearch:9.2.4"

func ManageElasticsearch() *cli.Command {
	return &cli.Command{
		Name:    "elasticsearch",
		Aliases: []string{"es"},
		Usage:   "Manage Elasticsearch containers",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start an Elasticsearch container",
				Action: func(c *cli.Context) error {
					if err := startElasticsearchContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop an Elasticsearch container",
				Action: func(c *cli.Context) error {
					if err := stopElasticsearchContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the Elasticsearch container",
				Action: func(c *cli.Context) error {
					if elasticsearchContainerExists() {
						fmt.Println("✅ elasticsearch container is running")
					} else {
						fmt.Println("❌ elasticsearch container is not running")
					}

					return nil
				},
			},
			{
				Name:  "url",
				Usage: "Get connection strings for Elasticsearch",
				Action: func(c *cli.Context) error {
					fmt.Println("http://localhost:9200")

					return nil
				},
			},
		},
	}
}

func elasticsearchContainerExists() bool {
	return docker.GetRunningContainerByImage(ElasticsearchImage) != nil
}

func startElasticsearchContainer() error {
	if elasticsearchContainerExists() {
		return errors.New("❌ elasticsearch container is already running")
	}

	portBinding := nat.PortMap{
		"9200/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "9200",
			},
		},
		"9300/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "9300",
			},
		},
	}

	containerConfig := &container.Config{
		Image: ElasticsearchImage,
		Env: []string{
			"discovery.type=single-node",
			"xpack.security.enabled=false",
			"ES_JAVA_OPTS=-Xms512m -Xmx512m",
		},
		ExposedPorts: nat.PortSet{
			"9200/tcp": struct{}{},
			"9300/tcp": struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBinding,
	}

	dockerClient := docker.Client

	out, err := dockerClient.ImagePull(context.Background(), ElasticsearchImage, image.PullOptions{})

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

	fmt.Println("✅ elasticsearch container started successfully")

	return nil
}

func stopElasticsearchContainer() error {
	dockerClient := docker.Client
	runningElasticsearchContainer := docker.GetRunningContainerByImage(ElasticsearchImage)

	if runningElasticsearchContainer == nil {
		return errors.New("❌ elasticsearch container is not running")
	}

	if err := dockerClient.ContainerStop(context.Background(), runningElasticsearchContainer.ID, container.StopOptions{}); err != nil {
		return fmt.Errorf("❌ error stopping container: %v", err)
	}

	fmt.Println("✅ elasticsearch container stopped successfully")

	return nil
}
