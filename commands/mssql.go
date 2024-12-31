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
						return err
					}

					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop a MSSQL container",
				Action: func(c *cli.Context) error {
					if err := stopMssqlContainer(); err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Check the status of the MSSQL container",
				Action: func(c *cli.Context) error {
					if mssqlContainerExists() {
						fmt.Println("✅ MSSQL container is running")
					} else {
						fmt.Println("❌ MSSQL container is not running")
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
			{
				Name:  "db:create",
				Usage: "Create a database in MSSQL",
				Action: func(c *cli.Context) error {
					if !mssqlContainerExists() {
						return errors.New("❌ you need to start the MSSQL container first before creating a database")
					}

					if c.NArg() == 0 {
						return errors.New("❌ please provide a database name")
					}

					dbName := c.Args().First()
					if err := createMSSQLDatabase(dbName); err != nil {
						return err
					} else {
						fmt.Println("✅ database created successfully")
					}

					return nil
				},
			},
			{
				Name:  "db:drop",
				Usage: "Drop a database in MSSQL",
				Action: func(c *cli.Context) error {
					if !mssqlContainerExists() {
						return errors.New("❌ you need to start the MSSQL container first before dropping a database")
					}

					if c.NArg() == 0 {
						return errors.New("❌ please provide a database name")
					}

					dbName := c.Args().First()
					if err := dropMSSQLDatabase(dbName); err != nil {
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

func mssqlContainerExists() bool {
	return docker.GetRunningContainerByImage(MssqlImage) != nil
}

func startMssqlContainer() error {
	if mssqlContainerExists() {
		return errors.New("❌ MSSQL container is already running")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("❌ error getting user home directory: %v", err)
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

	volumePath := homeDir + "/docker_volumes/mssql_data"

	hostConfig := &container.HostConfig{
		PortBindings: portBinding,
		Binds:        []string{volumePath + ":/var/opt/mssql"},
	}

	dockerClient := docker.Client

	out, err := dockerClient.ImagePull(context.Background(), MssqlImage, image.PullOptions{})

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

	fmt.Println("✅ MSSQL container started successfully")

	return nil
}

func stopMssqlContainer() error {
	dockerClient := docker.Client
	runningMssqlContainer := docker.GetRunningContainerByImage(MssqlImage)

	if runningMssqlContainer == nil {
		return errors.New("❌ MSSQL container is not running")
	}

	if err := dockerClient.ContainerStop(context.Background(), runningMssqlContainer.ID, container.StopOptions{}); err != nil {
		return fmt.Errorf("❌ error stopping container: %v", err)
	}

	fmt.Println("✅ MSSQL container stopped successfully")

	return nil
}

func mssqlConnectionStrings() string {
	formats := []string{
		fmt.Sprintf("Server=localhost,1433;Database=master;User Id=sa;Password=%s;TrustServerCertificate=true", MssqlPassword),
	}

	return strings.Join(formats[:], "\n")
}

func createMSSQLDatabase(dbName string) error {
	execCmd := fmt.Sprintf("CREATE DATABASE %s", dbName)
	cmd := exec.Command("sqlcmd", "-S", "localhost,1433", "-U", "sa", "-P", MssqlPassword, "-Q", execCmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ error running command: %v", err)
	}

	return nil
}

func dropMSSQLDatabase(dbName string) error {
	execCmd := fmt.Sprintf("DROP DATABASE %s", dbName)
	cmd := exec.Command("sqlcmd", "-S", "localhost,1433", "-U", "sa", "-P", MssqlPassword, "-Q", execCmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("❌ error running command: %v", err)
	}

	return nil
}
