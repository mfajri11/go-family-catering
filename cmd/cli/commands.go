package cmd

import (
	"family-catering/config"
	"family-catering/internal/app"
	"family-catering/pkg/db/migration"
	"family-catering/pkg/logger"
	"fmt"
	"os"

	cli "github.com/urfave/cli/v2"
)

var cmds []*cli.Command

func init() {
	RegisterCommands(
		version(),
		migrate(),
		rollbacks(),
		step(),
		drop(),
		run(),
		start())
}

func RegisterCommands(args ...*cli.Command) {
	cmds = append(cmds, args...)
}

func version() *cli.Command {
	command := &cli.Command{
		Name:        "version",
		Description: "show the current version of the app",
		Action: func(c *cli.Context) error {
			fmt.Println(config.Cfg().App.Version)
			return nil
		},
	}
	return command
}

func migrate() *cli.Command {
	command := &cli.Command{
		Name:        "migrations",
		Description: "migrate all the way up from active schema",
		Action: func(c *cli.Context) error {
			return migration.Up()
		},
	}
	return command
}

func rollbacks() *cli.Command {
	command := &cli.Command{
		Name:        "rollbacks",
		Description: "migrate all the way down from active schema",
		Action: func(c *cli.Context) error {
			return migration.Down()
		},
	}
	return command
}

func step() *cli.Command {
	command := &cli.Command{
		Name:        "step",
		Description: "migrate n step up/down from active schema (if n > 0 it will migrate up, otherwise is down)",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "n", Usage: "number of step relative to current schema version"},
		},
		Action: func(c *cli.Context) error {
			return migration.Step(c.Int("n"))
		},
	}
	return command
}

func drop() *cli.Command {
	command := &cli.Command{
		Name:        "drop",
		Description: "delete database",
		Action: func(c *cli.Context) error {
			return migration.Drop()
		},
	}
	return command
}

func run() *cli.Command {
	command := &cli.Command{
		Name:        "run",
		Description: "running the application",
		Action: func(c *cli.Context) error {
			return app.Run()
		},
	}
	return command
}

func start() *cli.Command {
	command := &cli.Command{
		Name:        "start",
		Description: "running the application and migrate from current active schema all the way up",
		Action: func(c *cli.Context) error {
			err := migration.Up()
			if err != nil {
				err = fmt.Errorf("cli: error migrate up: %w", err)
				logger.Error(err, "error migrate up")
				return err
			}
			return app.Run()
		},
	}
	return command
}

func Execute() error {
	app := cli.NewApp()
	app.Name = "family-catering CLI app"
	app.Description = "Interfacing utility for managing catering"
	app.Commands = cmds
	return app.Run(os.Args)
}
