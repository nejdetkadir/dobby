package main

import (
	"dobby/commands"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	registeredCommands := []*cli.Command{
		commands.ManageRedis(),
		commands.ManageMSSQL(),
		commands.ManagePSQL(),
		commands.ManageProxyman(),
		commands.ManagerRandom(),
		commands.ManageProcess(),
		commands.ManageRabbitMQ(),
		commands.ManageElasticsearch(),
		commands.ManageMongoDB(),
		commands.ManageLocalStack(),
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
