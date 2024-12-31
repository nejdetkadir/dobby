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
	"os/exec"
	"strings"
)

const PsqlImage = "postgres:17.2"

var formats = []string{
	"postgresql://postgres:metamorphmagus@localhost:5432/postgres",
	"Host=localhost;Port=5432;Database=postgres;User ID=postgres;Password=metamorphmagus",
}

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
						return err
					}

					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop a PSQL container",
				Action: func(c *cli.Context) error {
					if err := stopPSQLContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the PSQL container",
				Action: func(c *cli.Context) error {
					if psqlContainerExists() {
						fmt.Println("✅ PSQL container is running")
					} else {
						fmt.Println("❌ PSQL container is not running")
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
			{
				Name:  "db:create",
				Usage: "Create a new database",
				Action: func(c *cli.Context) error {
					if !psqlContainerExists() {
						return errors.New("❌ you need to start the PSQL container first before creating a database")
					}

					if c.NArg() == 0 {
						return errors.New("❌ please provide a database name")
					}

					dbName := c.Args().First()
					if err := createPSQLDatabase(dbName); err != nil {
						return err
					} else {
						fmt.Println("✅ database created successfully")
					}

					return nil
				},
			},
			{
				Name:  "db:drop",
				Usage: "Drop an existing database",
				Action: func(c *cli.Context) error {
					if !psqlContainerExists() {
						return errors.New("❌ you need to start the PSQL container first before dropping a database")
					}

					if c.NArg() == 0 {
						return errors.New("❌ please provide a database name")
					}

					dbName := c.Args().First()
					if err := dropPSQLDatabase(dbName); err != nil {
						return err
					} else {
						fmt.Println("✅ database dropped successfully")
					}

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
		return errors.New("❌ PSQL container already running")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("❌ error getting user home directory: %v", err)
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
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=postgres",
		},
		ExposedPorts: nat.PortSet{
			"5432/tcp": struct{}{},
		},
	}

	volumePath := homeDir + "/docker_volumes/psql_data"

	hostConfig := &container.HostConfig{
		PortBindings: portBinding,
		Binds:        []string{volumePath + ":/var/lib/postgresql/data"},
	}

	dockerClient := docker.Client

	out, err := dockerClient.ImagePull(context.Background(), PsqlImage, image.PullOptions{})

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

	fmt.Println("✅ PSQL container started successfully")

	return nil
}

func stopPSQLContainer() error {
	dockerClient := docker.Client
	runningContainer := docker.GetRunningContainerByImage(PsqlImage)

	if runningContainer == nil {
		return errors.New("❌ PSQL container is not running")
	}

	if err := dockerClient.ContainerStop(context.Background(), runningContainer.ID, container.StopOptions{}); err != nil {
		return fmt.Errorf("❌ error stopping container: %v", err)
	}

	fmt.Println("✅ PSQL container stopped successfully")

	return nil
}

func psqlConnectionStrings() string {
	return strings.Join(formats[:], "\n")
}

func createPSQLDatabase(dbName string) error {
	execCmd := fmt.Sprintf("psql \"%s\" -c \"CREATE DATABASE %s;\"", formats[0], dbName)
	cmd := exec.Command("sh", "-c", execCmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ error running command: %v", err)
	}

	return nil
}

func dropPSQLDatabase(dbName string) error {
	execCmd := fmt.Sprintf("psql \"%s\" -c \"DROP DATABASE %s;\"", formats[0], dbName)
	cmd := exec.Command("sh", "-c", execCmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ error running command: %v", err)
	}

	return nil
}
