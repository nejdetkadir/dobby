package commands

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"os/exec"
	"strconv"
)

func ManageProcess() *cli.Command {
	return &cli.Command{
		Name:    "process",
		Aliases: []string{"pr"},
		Usage:   "Manage Process",
		Subcommands: []*cli.Command{
			{
				Name:    "kill",
				Usage:   "Kill a process with using port",
				Aliases: []string{"k"},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return errors.New("❌ you need to provide a port to kill the process")
					}

					if err := killProcessByPort(c.Args().First()); err != nil {
						return err
					} else {
						fmt.Println("✅ process has been killed")
					}

					return nil
				},
			},
		},
	}
}

func killProcessByPort(port string) error {
	numericPort, err := strconv.Atoi(port)

	if err != nil {
		return errors.New("❌ port should be numeric")
	}

	execCmd := fmt.Sprintf("lsof -i tcp:%d -t | head -n 1", numericPort)
	cmd := exec.Command("sh", "-c", execCmd)

	pid, err := cmd.Output()

	if err != nil {
		return errors.New("❌ process not found with the given port")
	}

	if len(pid) == 0 {
		return errors.New("❌ process not found with the given port")
	}

	execCmd = fmt.Sprintf("kill -9 %s", string(pid))
	cmd = exec.Command("sh", "-c", execCmd)

	if err = cmd.Run(); err != nil {
		return errors.New("❌ process could not be killed")
	}

	return nil
}
