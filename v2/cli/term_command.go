package main

import (
	"github.com/rivo/tview"
	"github.com/urfave/cli/v2"
)

var termCommand = &cli.Command{
	Name:   "term",
	Usage:  "Start a Brigade text terminal",
	Flags:  []cli.Flag{},
	Action: term,
}

func term(c *cli.Context) error {
	box := tview.NewBox().SetBorder(true).SetTitle("Hello, world!")
	if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
		return err
	}
	return nil
}
