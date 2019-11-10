package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

var help = cli.HelpCommand("Display help information")
var logPrefix = ""
var showTimeInLog = true
var serviceName = "gologD.service"

type argT struct {
	cli.Helper
}

var root = &cli.Command{
	Argv: func() interface{} { return new(argT) },
	Fn: func(ctx *cli.Context) error {
		fmt.Println("Usage: gologD <install/start/stop/push> []")
		return nil
	},
}

func main() {
	if err := cli.Root(root,
		cli.Tree(help),
		cli.Tree(pushCMD),
		cli.Tree(startCMD),
		cli.Tree(installCMD),
		cli.Tree(stopCMD),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
