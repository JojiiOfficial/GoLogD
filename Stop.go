package main

import (
	"github.com/JojiiOfficial/SystemdGoService"

	"github.com/mkideal/cli"
)

var stopCMD = &cli.Command{
	Name:    "stop",
	Aliases: []string{"stop"},
	Desc:    "stops the deamon",
	Fn: func(ct *cli.Context) error {
		err := setDeamonStatus(SystemdGoService.Stop)
		if err != nil {
			LogError("Error: " + err.Error())
		} else {
			LogInfo("Stopped successfully")
		}
		return nil
	},
}
