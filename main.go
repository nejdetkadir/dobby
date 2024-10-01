package main

import (
	"dobby/commands"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	registeredCommands := []*cli.Command{
		commands.ManageRedis(),
		commands.ManageMSSQL(),
		commands.ManagePSQL(),
		commands.ManageProxyman(),
	}

	app := &cli.App{
		Name:    "Dobby CLI",
		Version: "1.0.0",
		Authors: []*cli.Author{
			{
				Name:  "nejdetkadir",
				Email: "nejdetkadir.550@gmail.com",
			},
		},
		Commands: registeredCommands,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
