package commands

import (
	"context"
	"dobby/docker"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/go-connections/nat"
	"github.com/urfave/cli/v2"
)

const PostGISImage = "imresamu/postgis:17-3.5-alpine3.22"

var postgisFormats = []string{
	"postgresql://postgres:metamorphmagus@localhost:5433/postgres",
	"Host=localhost;Port=5433;Database=postgres;User ID=postgres;Password=metamorphmagus",
}

func ManagePostGIS() *cli.Command {
	return &cli.Command{
		Name:    "postgis",
		Aliases: []string{"pg"},
		Usage:   "Manage PostGIS containers",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a PostGIS container",
				Action: func(c *cli.Context) error {
					if err := startPostGISContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop a PostGIS container",
				Action: func(c *cli.Context) error {
					if err := stopPostGISContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the PostGIS container",
				Action: func(c *cli.Context) error {
					if postgisContainerExists() {
						fmt.Println("✅ PostGIS container is running")
					} else {
						fmt.Println("❌ PostGIS container is not running")
					}

					return nil
				},
			},
			{
				Name:  "url",
				Usage: "Get connection strings for PostGIS",
				Action: func(c *cli.Context) error {
					fmt.Println(postgisConnectionStrings())

					return nil
				},
			},
			{
				Name:  "db:create",
				Usage: "Create a new PostGIS-enabled database",
				Action: func(c *cli.Context) error {
					if !postgisContainerExists() {
						return errors.New("❌ you need to start the PostGIS container first before creating a database")
					}

					if c.NArg() == 0 {
						return errors.New("❌ please provide a database name")
					}

					dbName := c.Args().First()
					if err := createPostGISDatabase(dbName); err != nil {
						return err
					} else {
						fmt.Println("✅ database created with PostGIS extensions enabled")
					}

					return nil
				},
			},
			{
				Name:  "db:drop",
				Usage: "Drop an existing database",
				Action: func(c *cli.Context) error {
					if !postgisContainerExists() {
						return errors.New("❌ you need to start the PostGIS container first before dropping a database")
					}

					if c.NArg() == 0 {
						return errors.New("❌ please provide a database name")
					}

					dbName := c.Args().First()
					if err := dropPostGISDatabase(dbName); err != nil {
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

func postgisContainerExists() bool {
	return docker.GetRunningContainerByImage(PostGISImage) != nil
}

func startPostGISContainer() error {
	if postgisContainerExists() {
		return errors.New("❌ PostGIS container already running")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("❌ error getting user home directory: %v", err)
	}

	portBinding := nat.PortMap{
		"5432/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "5433",
			},
		},
	}

	containerConfig := &container.Config{
		Image: PostGISImage,
		Env: []string{
			"POSTGRES_PASSWORD=metamorphmagus",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=postgres",
		},
		ExposedPorts: nat.PortSet{
			"5432/tcp": struct{}{},
		},
	}

	volumePath := homeDir + "/docker_volumes/postgis_data"

	hostConfig := &container.HostConfig{
		PortBindings: portBinding,
		Binds:        []string{volumePath + ":/var/lib/postgresql/data"},
	}

	dockerClient := docker.Client

	out, err := dockerClient.ImagePull(context.Background(), PostGISImage, image.PullOptions{})

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

	fmt.Println("✅ PostGIS container started successfully")

	return nil
}

func stopPostGISContainer() error {
	dockerClient := docker.Client
	runningContainer := docker.GetRunningContainerByImage(PostGISImage)

	if runningContainer == nil {
		return errors.New("❌ PostGIS container is not running")
	}

	if err := dockerClient.ContainerStop(context.Background(), runningContainer.ID, container.StopOptions{}); err != nil {
		return fmt.Errorf("❌ error stopping container: %v", err)
	}

	fmt.Println("✅ PostGIS container stopped successfully")

	return nil
}

func postgisConnectionStrings() string {
	return strings.Join(postgisFormats[:], "\n")
}

func createPostGISDatabase(dbName string) error {
	createDB := fmt.Sprintf("psql \"%s\" -c \"CREATE DATABASE %s;\"", postgisFormats[0], dbName)
	cmd := exec.Command("sh", "-c", createDB)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ error creating database: %v", err)
	}

	dbConnStr := fmt.Sprintf("postgresql://postgres:metamorphmagus@localhost:5433/%s", dbName)
	enableExtensions := fmt.Sprintf(
		"psql \"%s\" -c \"CREATE EXTENSION IF NOT EXISTS postgis CASCADE; CREATE EXTENSION IF NOT EXISTS postgis_topology CASCADE; CREATE EXTENSION IF NOT EXISTS postgis_tiger_geocoder CASCADE;\"",
		dbConnStr,
	)
	extCmd := exec.Command("sh", "-c", enableExtensions)

	extCmd.Stdout = os.Stdout
	extCmd.Stderr = os.Stderr

	if err := extCmd.Run(); err != nil {
		return fmt.Errorf("❌ error enabling PostGIS extensions: %v", err)
	}

	return nil
}

func dropPostGISDatabase(dbName string) error {
	execCmd := fmt.Sprintf("psql \"%s\" -c \"DROP DATABASE %s;\"", postgisFormats[0], dbName)
	cmd := exec.Command("sh", "-c", execCmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ error running command: %v", err)
	}

	return nil
}
