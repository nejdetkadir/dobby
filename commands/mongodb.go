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

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/go-connections/nat"
	"github.com/urfave/cli/v2"
)

const MongoDBImage = "mongo:8.0"

func ManageMongoDB() *cli.Command {
	return &cli.Command{
		Name:    "mongodb",
		Aliases: []string{"mongo"},
		Usage:   "Manage MongoDB containers",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a MongoDB container",
				Action: func(c *cli.Context) error {
					if err := startMongoDBContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop a MongoDB container",
				Action: func(c *cli.Context) error {
					if err := stopMongoDBContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the MongoDB container",
				Action: func(c *cli.Context) error {
					if mongoDBContainerExists() {
						fmt.Println("✅ mongodb container is running")
					} else {
						fmt.Println("❌ mongodb container is not running")
					}

					return nil
				},
			},
			{
				Name:  "url",
				Usage: "Get connection strings for MongoDB",
				Action: func(c *cli.Context) error {
					fmt.Println("mongodb://admin:admin123@localhost:27017/")
					fmt.Println("mongodb://admin:admin123@localhost:27017/?authSource=admin")

					return nil
				},
			},
			{
				Name:  "db:create",
				Usage: "Create a new database",
				Action: func(c *cli.Context) error {
					if !mongoDBContainerExists() {
						return errors.New("❌ you need to start the MongoDB container first before creating a database")
					}

					if c.NArg() == 0 {
						return errors.New("❌ please provide a database name")
					}

					dbName := c.Args().First()
					if err := createMongoDBDatabase(dbName); err != nil {
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
					if !mongoDBContainerExists() {
						return errors.New("❌ you need to start the MongoDB container first before dropping a database")
					}

					if c.NArg() == 0 {
						return errors.New("❌ please provide a database name")
					}

					dbName := c.Args().First()
					if err := dropMongoDBDatabase(dbName); err != nil {
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

func mongoDBContainerExists() bool {
	return docker.GetRunningContainerByImage(MongoDBImage) != nil
}

func startMongoDBContainer() error {
	if mongoDBContainerExists() {
		return errors.New("❌ mongodb container is already running")
	}

	portBinding := nat.PortMap{
		"27017/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "27017",
			},
		},
	}

	containerConfig := &container.Config{
		Image: MongoDBImage,
		Env: []string{
			"MONGO_INITDB_ROOT_USERNAME=admin",
			"MONGO_INITDB_ROOT_PASSWORD=admin123",
		},
		ExposedPorts: nat.PortSet{
			"27017/tcp": struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBinding,
	}

	dockerClient := docker.Client

	out, err := dockerClient.ImagePull(context.Background(), MongoDBImage, image.PullOptions{})

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

	fmt.Println("✅ mongodb container started successfully")

	return nil
}

func stopMongoDBContainer() error {
	dockerClient := docker.Client
	runningMongoDBContainer := docker.GetRunningContainerByImage(MongoDBImage)

	if runningMongoDBContainer == nil {
		return errors.New("❌ mongodb container is not running")
	}

	if err := dockerClient.ContainerStop(context.Background(), runningMongoDBContainer.ID, container.StopOptions{}); err != nil {
		return fmt.Errorf("❌ error stopping container: %v", err)
	}

	fmt.Println("✅ mongodb container stopped successfully")

	return nil
}

func createMongoDBDatabase(dbName string) error {
	execCmd := fmt.Sprintf("mongosh \"mongodb://admin:admin123@localhost:27017/?authSource=admin\" --eval \"use %s; db.createCollection('init'); db.init.drop();\"", dbName)
	cmd := exec.Command("sh", "-c", execCmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ error running command: %v", err)
	}

	return nil
}

func dropMongoDBDatabase(dbName string) error {
	execCmd := fmt.Sprintf("mongosh \"mongodb://admin:admin123@localhost:27017/?authSource=admin\" --eval \"use %s; db.dropDatabase();\"", dbName)
	cmd := exec.Command("sh", "-c", execCmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ error running command: %v", err)
	}

	return nil
}
