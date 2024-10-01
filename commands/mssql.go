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

const (
	MssqlImage    = "mcr.microsoft.com/mssql/server:2019-latest"
	MssqlPassword = "Parselmouth1$"
)

func ManageMSSQL() *cli.Command {
	return &cli.Command{
		Name:    "mssql",
		Aliases: []string{"ms"},
		Usage:   "Manage MSSQL containers",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a MSSQL container",
				Action: func(c *cli.Context) error {
					if err := startMssqlContainer(); err != nil {
						return errors.New(fmt.Sprintf("🛑 Error starting MSSQL container: %v", err))
					}

					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop a MSSQL container",
				Action: func(c *cli.Context) error {
					if err := stopMssqlContainer(); err != nil {
						return errors.New(fmt.Sprintf("🛑 Error stopping MSSQL container: %v", err))
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the MSSQL container",
				Action: func(c *cli.Context) error {
					if mssqlContainerExists() {
						fmt.Println("MSSQL container is running ✅")
					} else {
						fmt.Println("MSSQL container is not running ❌")
					}

					return nil
				},
			},
			{
				Name:  "url",
				Usage: "Get connection strings for MSSQL",
				Action: func(c *cli.Context) error {
					fmt.Println(mssqlConnectionStrings())

					return nil
				},
			},
		},
	}
}

func mssqlContainerExists() bool {
	return docker.GetRunningContainerByImage(MssqlImage) != nil
}

func startMssqlContainer() error {
	if mssqlContainerExists() {
		return errors.New("mssql container is already running")
	}

	portBinding := nat.PortMap{
		"1433/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "1433",
			},
		},
	}

	containerConfig := &container.Config{
		Image: MssqlImage,
		Env: []string{
			"ACCEPT_EULA=Y",
			fmt.Sprintf("MSSQL_SA_PASSWORD=%s", MssqlPassword),
		},
		ExposedPorts: nat.PortSet{
			"1433/tcp": struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBinding,
	}

	dockerClient := docker.Client

	out, err := dockerClient.ImagePull(context.Background(), MssqlImage, image.PullOptions{})

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

	fmt.Println("MSSQL container started successfully 🚀")

	return nil
}

func stopMssqlContainer() error {
	dockerClient := docker.Client
	runningMssqlContainer := docker.GetRunningContainerByImage(MssqlImage)

	if runningMssqlContainer == nil {
		return errors.New("mssql container is not running")
	}

	if err := dockerClient.ContainerStop(context.Background(), runningMssqlContainer.ID, container.StopOptions{}); err != nil {
		return err
	}

	fmt.Println("MSSQL container stopped successfully ✋🏻")

	return nil
}

func mssqlConnectionStrings() string {
	formats := []string{
		fmt.Sprintf("Server=localhost,1433;Database=master;User Id=sa;Password=%s;TrustServerCertificate=true", MssqlPassword),
	}

	return strings.Join(formats[:], "\n")
}
