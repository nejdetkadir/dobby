package commands

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"os/exec"
	"strconv"
)

func ManagerRandom() *cli.Command {
	return &cli.Command{
		Name:    "random",
		Aliases: []string{"ra"},
		Usage:   "Create random data",
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return errors.New("❌ you need to pass length of random string")
			}

			length := c.Args().First()

			if length == "" {
				return errors.New("❌ you need to pass length of random string")
			}

			randomString, err := generateRandomString(length)
			if err != nil {
				return err
			}

			fmt.Println("✅ Random string generated successfully")
			fmt.Println(randomString)

			return nil
		},
	}
}

func generateRandomString(length string) (string, error) {
	_, err := strconv.Atoi(length)
	if err != nil {
		return "", errors.New("❌ length must be a number")
	}

	execCmd := exec.Command("openssl", "rand", "-base64", length)
	output, err := execCmd.Output()

	if err != nil {
		return "", errors.New("❌ failed to generate random string")
	}

	return string(output), nil
}
