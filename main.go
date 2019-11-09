package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

var help = cli.HelpCommand("Display help information")
var logPrefix = ""
var showTimeInLog = true

type argT struct {
	cli.Helper
}

var root = &cli.Command{
	Argv: func() interface{} { return new(argT) },
	Fn: func(ctx *cli.Context) error {
		fmt.Println("Usage: goLogd <push> []")
		return nil
	},
}

func main() {
	if err := cli.Root(root,
		cli.Tree(help),
		cli.Tree(pushCMD),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
