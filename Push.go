package main

import (
	"github.com/mkideal/cli"
)

type pushT struct {
	cli.Helper
}

var pushCMD = &cli.Command{
	Name:    "push",
	Aliases: []string{"p", "push"},
	Desc:    "push new logs",
	Argv:    func() interface{} { return new(pushT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*pushT)
		_ = argv

		data, err := checkData()
		if err != nil {
			LogCritical("Couldn't load data: " + err.Error())
			return nil
		}

		config, err := checkConfig()
		if err != nil {
			LogCritical("Couldn't load config: " + err.Error())
			return nil
		}

		_, _ = data, config
		return nil
	},
}
