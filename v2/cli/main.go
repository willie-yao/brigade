package main

import (
	"fmt"
	"os"

	"github.com/brigadecore/brigade/v2/internal/signals"
	"github.com/brigadecore/brigade/v2/internal/version"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "Brigade"
	app.Usage = "Event Driven Scripting for Kubernetes"
	app.Version = fmt.Sprintf(
		"%s -- commit %s",
		version.Version(),
		version.Commit(),
	)
	app.Commands = []*cli.Command{
		eventCommand,
		loginCommand,
		logoutCommand,
		projectCommand,
		rolesCommands,
		serviceAccountCommand,
		userCommand,
		termCommand,
	}
	fmt.Println()
	if err := app.RunContext(signals.Context(), os.Args); err != nil {
		fmt.Printf("\n%s\n\n", err)
		os.Exit(1)
	}
	fmt.Println()
}
