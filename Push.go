package main

import (
	"fmt"
	"os"

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

		if len(config.Token) != 64 {
			LogInfo("You need to enter a valid token")
			os.Exit(1)
			return nil
		}

		if len(config.Files) == 0 {
			LogInfo("No logfile configured. Nothing to do")
			os.Exit(1)
			return nil
		}

		data.Validate()

		filesToWatch := data.MergeWithConfig(*config)
		for _, file := range filesToWatch {
			fmt.Println(file.File)
		}

		data.Save()

		return nil
	},
}
