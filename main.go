package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

var help = cli.HelpCommand("Display help information")
var showTimeInLog = true
var serviceName = "goLogD"
var logPrefix = serviceName + " "

type argT struct {
	cli.Helper
}

var root = &cli.Command{
	Argv: func() interface{} { return new(argT) },
	Fn: func(ctx *cli.Context) error {
		fmt.Println("Usage: ./gologd <install/start/stop/push/addFile> []")
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
		cli.Tree(addLogCMD),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
