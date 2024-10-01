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
	"strings"
)

const PsqlImage = "postgres:latest"

func ManagePSQL() *cli.Command {
	return &cli.Command{
		Name:    "psql",
		Aliases: []string{"ps"},
		Usage:   "Manage PSQL containers",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a PSQL container",
				Action: func(c *cli.Context) error {
					if err := startPSQLContainer(); err != nil {
						return errors.New(fmt.Sprintf("🛑 Error starting PSQL container: %v", err))
					}

					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop a PSQL container",
				Action: func(c *cli.Context) error {
					if err := stopPSQLContainer(); err != nil {
						return errors.New(fmt.Sprintf("🛑 Error stopping PSQL container: %v", err))
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the PSQL container",
				Action: func(c *cli.Context) error {
					if psqlContainerExists() {
						fmt.Println("PSQL container is running ✅")
					} else {
						fmt.Println("PSQL container is not running ❌")
					}

					return nil
				},
			},
			{
				Name:  "url",
				Usage: "Get connection strings for PSQL",
				Action: func(c *cli.Context) error {
					fmt.Println(psqlConnectionStrings())

					return nil
				},
			},
		},
	}
}

func psqlContainerExists() bool {
	return docker.GetRunningContainerByImage(PsqlImage) != nil
}

func startPSQLContainer() error {
	if psqlContainerExists() {
		return errors.New("PSQL container already running")
	}

	portBinding := nat.PortMap{
		"5432/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "5432",
			},
		},
	}

	containerConfig := &container.Config{
		Image: PsqlImage,
		Env: []string{
			"POSTGRES_PASSWORD=metamorphmagus",
		},
		ExposedPorts: nat.PortSet{
			"5432/tcp": struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBinding,
	}

	dockerClient := docker.Client

	out, err := dockerClient.ImagePull(context.Background(), PsqlImage, image.PullOptions{})

	if err != nil {
		return err
	}

	defer func(out io.ReadCloser) {
		err := out.Close()
		if err != nil {
			log.Fatalf("Error closing image pull response: %v", err)
		}
	}(out)

	_, err = io.Copy(os.Stdout, out)
	if err != nil {
		return err
	}

	resp, err := dockerClient.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, "")

	if err != nil {
		return err
	}

	if err := dockerClient.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	fmt.Println("PSQL container started successfully ✅")

	return nil
}

func stopPSQLContainer() error {
	dockerClient := docker.Client
	runningContainer := docker.GetRunningContainerByImage(PsqlImage)

	if runningContainer == nil {
		return errors.New("PSQL container is not running")
	}

	if err := dockerClient.ContainerStop(context.Background(), runningContainer.ID, container.StopOptions{}); err != nil {
		return err
	}

	fmt.Println("PSQL container stopped successfully ✋🏻")

	return nil
}

func psqlConnectionStrings() string {
	formats := []string{
		"postgresql://postgres:metamorphmagus@localhost:5432/postgres",
		"Host=localhost;Port=5432;Database=postgres;User ID=postgres",
	}

	return strings.Join(formats[:], "\n")
}
