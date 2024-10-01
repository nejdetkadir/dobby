package commands

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"os/exec"
)

func ManageProxyman() *cli.Command {
	return &cli.Command{
		Name:    "proxyman",
		Aliases: []string{"px"},
		Usage:   "Manage Proxyman",
		Subcommands: []*cli.Command{
			{
				Name:  "terminal",
				Usage: "Open Proxyman terminal",
				Action: func(c *cli.Context) error {
					if err := openProxymanTerminal(); err != nil {
						return errors.New(fmt.Sprintf("🛑 Error opening Proxyman terminal: %v", err))
					}

					return nil
				},
			},
		},
	}
}

func openProxymanTerminal() error {
	execCmd := "set -a && source $HOME/.proxyman/proxyman_env_automatic_setup.sh && set +a"

	cmd := exec.Command("bash", "-c", execCmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
