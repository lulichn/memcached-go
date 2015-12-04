package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/lulichn/memcached-go/command"
	"os"
)

var GlobalFlags = []cli.Flag{
	cli.StringSliceFlag {
		Name:  "host",
		//Value: []string{"localhost:11211"},
		Usage: "HostName:Port",
	},
	cli.BoolFlag {
		Name:  "es",
		Usage: "ElastiCache",
	},
}

var Commands = []cli.Command{
	{
		Name:   "list",
		Usage:  "List cached items",
		Action: command.CmdList,
		Flags:  []cli.Flag{},
	},
	{
		Name:   "get",
		Usage:  "Get cahced item",
		Action: command.CmdGet,
		Flags:  []cli.Flag{},
	},
	{
		Name:   "delete",
		Usage:  "Delete cached item",
		Action: command.CmdDelete,
		Flags:  []cli.Flag{},
	},
	{
		Name:   "set",
		Usage:  "Set item",
		Action: command.CmdSet,
		Flags:  []cli.Flag{},
	},
	{
		Name:   "stats",
		Usage:  "Stats",
		Action: command.CmdStats,
		Flags:  []cli.Flag{},
	},
}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
