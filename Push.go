package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	_ "github.com/fsnotify/fsnotify"
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

		data.Save()

		if len(filesToWatch) == 0 {
			LogError("No valid logfile configuration found. Exiting...")
			os.Exit(1)
			return nil
		}
		runFileWatcher(config, data, filesToWatch)

		return nil
	},
}

func runFileWatcher(config *Config, data *Data, filesToWatch []WatchedFile) {
	for _, file := range filesToWatch {
		go watchFile(config, data, file)
	}
	for {
		time.Sleep(1 * time.Second)
	}
}

func watchFile(config *Config, data *Data, file WatchedFile) {
	var fd *FileData
	var confD *FileConfig
	for i, filec := range config.Files {
		if filec.File == file.FileData.FileName {
			confD = &config.Files[i]
		}
	}
	for i, filedata := range data.Files {
		if filedata.FileName == file.FileData.FileName {
			fd = &data.Files[i]
		}
	}

	sFilterMode := strings.Trim(confD.FilterMode, " ")
	if len(sFilterMode) > 0 {
		LogInfo("Wrong filtermode for \"" + fd.FileName + "\"! Using \"and\"")
	} else {
		LogInfo("No filtermode set for \"" + fd.FileName + "\"! Using \"and\"")
	}
	fireSyslogChanges(file, fd, data, confD, config)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				_ = event
				if !ok {
					return
				}
				if fsnotify.Write == fsnotify.Write {
					fireSyslogChanges(file, fd, data, confD, config)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(file.File)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func fireSyslogChanges(file WatchedFile, fd *FileData, data *Data, fileConfig *FileConfig, config *Config) {
	start := time.Now()
	logs := ParseSysLogFile(file.File, fileConfig, fd.LastLogTime)
	for _, i := range logs {
		fmt.Println(i)
	}
	fmt.Println("Duration:", time.Now().Sub(start).String())
	err := pushSyslogs(config, fd.LastLogTime, logs)
	if err != nil {
		LogError("Error reporting: " + err.Error())
	} else {
		fd.LastLogTime = time.Now().Unix()
		data.Save()
	}
}

func pushSyslogs(config *Config, startTime int64, logs []*SyslogEntry) error {
	psr := PushSyslogRequest{
		Token:     config.Token,
		StartTime: startTime,
		Syslogs:   logs,
	}
	d, err := json.Marshal(psr)
	if err != nil {
		return err
	}
	resp, err := request(config.Host, "/push/syslog", d, config.IgnoreCert)
	if err != nil {
		return err
	}
	fmt.Println("resp", resp)
	return nil
}
