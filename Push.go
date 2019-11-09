package main

import (
	"github.com/mkideal/cli"
)

type pushT struct {
	cli.Helper
	BackupIPtables bool `cli:"t,iptables" usage:"Update iptables" dft:"false"`
	BackupIPset    bool `cli:"s,ipset" usage:"Update ipset" dft:"true"`
}

var pushCMD = &cli.Command{
	Name:    "push",
	Aliases: []string{"p", "push"},
	Desc:    "push new logs",
	Argv:    func() interface{} { return new(pushT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*pushT)
		_ = argv
		return nil
	},
}
