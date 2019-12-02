package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/hpcloud/tail"
	"github.com/mkideal/cli"
)

type pushT struct {
	cli.Helper
	Verbose int `cli:"v,verbose" usage:"Show more log output"`
}

var verbose int

var pushCMD = &cli.Command{
	Name:    "push",
	Aliases: []string{"p", "push"},
	Desc:    "push new logs",
	Argv:    func() interface{} { return new(pushT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*pushT)
		verbose = argv.Verbose
		if verbose > 1 {
			LogInfo("Starting with verboselevel: " + strconv.Itoa(verbose))
		}
		initLoggerFiles(logPrefix)

		data, config, er := validateFiles()
		if er {
			os.Exit(1)
			return nil
		}

		config.Validate()
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

func validateFiles() (data *Data, config *Config, erro bool) {
	erro = false
	data, err := checkData()
	if err != nil {
		LogCritical("Couldn't load data: " + err.Error())
		return nil, nil, true
	}

	config, err = checkConfig()
	if err != nil {
		LogCritical("Couldn't load config: " + err.Error())
		return nil, nil, true
	}

	if len(config.Token) != 24 {
		LogInfo("You need to enter a valid token before")
		os.Exit(1)
		return nil, nil, true
	}

	if len(config.Files) == 0 {
		LogInfo("No logfile configured. Nothing to do")
		os.Exit(1)
		return nil, nil, true
	}
	return
}

func runFileWatcher(config *Config, data *Data, filesToWatch []WatchedFile) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	for _, file := range filesToWatch {
		go watchFile(config, data, file, watcher)
	}
	for {
		time.Sleep(1 * time.Second)
	}
}

func watchFile(config *Config, data *Data, file WatchedFile, watcher *fsnotify.Watcher) {
	_, er := os.Stat(file.File)
	if er != nil {
		return
	}
	var fd *FileData
	var confD *FileConfig
	for i, filec := range config.Files {
		if filec.File == file.FileData.FileName {
			confD = &config.Files[i]
			for _, k := range config.GlobalKeyBlacklist {
				if len(strings.Trim(k, " ")) > 0 {
					confD.KeyBlacklist = append(confD.KeyBlacklist, k)
				}
			}
		}
	}

	for i, filedata := range data.Files {
		if filedata.FileName == file.FileData.FileName {
			fd = &data.Files[i]
		}
	}

	sFilterMode := strings.Trim(confD.FilterMode, " ")
	if sFilterMode != "and" && sFilterMode != "or" {
		if len(sFilterMode) > 0 {
			LogInfo("Wrong filtermode for \"" + fd.FileName + "\"! Using \"and\"")
		} else {
			LogInfo("No filtermode set for \"" + fd.FileName + "\"! Using \"and\"")
		}
	}

	t, err := tail.TailFile(file.File, tail.Config{Follow: true, ReOpen: true, Poll: true})
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return
	}
	started := true
	var lastLine int64
	var lines []logLineData
	for line := range t.Lines {
		since := fd.LastLogTime
		prepared, tima, timelen, err := ParselogTime(file.File, line.Text)
		if prepared == nil || err != nil {
			if err != nil {
				LogError(err.Error())
			}
			continue
		}
		b := tima.Unix() < since
		if started {
			b = tima.Unix() <= since
		}
		if b {
			continue
		}
		if lastLine > 0 {
			if lastLine+1 > time.Now().Unix() {
				started = false
			}
		}
		lastLine = time.Now().Unix()
		if started {
			lines = append(lines, logLineData{
				prepared: prepared,
				tim:      tima,
				timelen:  timelen,
				line:     line.Text,
			})
		} else {
			if len(lines) > 0 {
				processLines(lines, fd, confD, config, data)
				lines = []logLineData{}
			} else {
				processLines([]logLineData{logLineData{
					prepared: prepared,
					tim:      tima,
					timelen:  timelen,
					line:     line.Text,
				}}, fd, confD, config, data)
			}
		}
	}
}

type logLineData struct {
	line     string
	prepared []string
	tim      time.Time
	timelen  int
}

func processLines(lines []logLineData, filedata *FileData, fileConfig *FileConfig, config *Config, data *Data) {
	startTime := time.Now()
	if fileConfig.LogType == Syslog {
		syslogEntries := []*SyslogEntry{}
		for _, line := range lines {
			loge := ParseSyslogMessage(&line, fileConfig, filedata.LastLogTime)
			if loge != nil && *loge != (SyslogEntry{}) {
				syslogEntries = append(syslogEntries, loge)
			} else if loge == nil {
				LogInfo("Couldn't parse " + filedata.FileName)
			}
		}
		handleSyslogChange(syslogEntries, startTime, config, filedata, data)
	} else if fileConfig.LogType == Custom {
		customLogEntries := []*CustomLogEntry{}
		for _, line := range lines {
			loge := parseCustomLogMessage(&line, fileConfig, filedata.LastLogTime)
			if loge != nil && *loge != (CustomLogEntry{}) {
				customLogEntries = append(customLogEntries, loge)
			} else if loge == nil {
				LogInfo("Couldn't parse " + filedata.FileName)
			}
		}
		handleCustomlogChange(customLogEntries, startTime, config, filedata, data)
	}
}
func handleCustomlogChange(logs []*CustomLogEntry, start time.Time, config *Config, fd *FileData, data *Data) {
	if len(logs) > 0 {
		duration := time.Since(start)
		if duration > 500*time.Millisecond {
			LogInfo("Duration: " + duration.String())
		}
		err := pushlogs(config, fd.LastLogTime, logs, "custom")
		if err != nil {
			LogError("Error reporting: " + err.Error())
			if errCounter > 20 {
				LogCritical("More than 20 errors in a row! Stopping service! look at your configuration")
				os.Exit(1)
				return
			}
		} else {
			fd.LastLogTime = time.Now().Unix()
			data.Save()
		}
	}
}

func handleSyslogChange(logs []*SyslogEntry, start time.Time, config *Config, filedata *FileData, data *Data) {
	if len(logs) > 0 {
		duration := time.Since(start)
		if duration > 500*time.Millisecond {
			LogInfo("Duration: " + duration.String())
		}
		err := pushlogs(config, filedata.LastLogTime, logs, "syslog")
		if err != nil {
			LogError("Error reporting: " + err.Error())
			if errCounter > 20 {
				LogCritical("More than 20 errors in a row! Stopping service! look at your configuration")
				os.Exit(1)
				return
			}
		} else {
			filedata.LastLogTime = time.Now().Unix()
			data.Save()
		}
	}
}

func firelogChange(file WatchedFile, fd *FileData, data *Data, fileConfig *FileConfig, config *Config) {
	//start := time.Now()
	//if fileConfig.LogType == Syslog {
	//logs := ParseSysLogFile(file.File, fileConfig, fd.LastLogTime)
	//for _, i := range logs {
	//LogInfo(i.Message)
	//}
	//if len(logs) > 0 {
	//duration := time.Since(start)
	//if duration > 500*time.Millisecond {
	//LogInfo("Duration: " + duration.String())
	//}
	//err := pushlogs(config, fd.LastLogTime, logs, "syslog")
	//if err != nil {
	//LogError("Error reporting: " + err.Error())
	//if errCounter > 20 {
	//LogCritical("More than 20 errors in a row! Stopping service! look at your configuration")
	//os.Exit(1)
	//return
	//}
	//} else {
	//fd.LastLogTime = time.Now().Unix()
	//data.Save()
	//}
	//}
	//} else if fileConfig.LogType == Custom {
	//logs := parseCustomLogfile(file.File, fileConfig, fd.LastLogTime)
	//for _, a := range logs {
	//LogInfo(a.Message)
	//}
	//if len(logs) > 0 {
	//duration := time.Since(start)
	//if duration > 500*time.Millisecond {
	//LogInfo("Duration: " + duration.String())
	//}
	//err := pushlogs(config, fd.LastLogTime, logs, "custom")
	//if err != nil {
	//LogError("Error reporting: " + err.Error())
	//if errCounter > 20 {
	//LogCritical("More than 20 errors in a row! Stopping service! look at your configuration")
	//os.Exit(1)
	//return
	//}
	//} else {
	//fd.LastLogTime = time.Now().Unix()
	//data.Save()
	//}
	//}
	//}
}

var errCounter = 0

func pushlogs(config *Config, startTime int64, logs interface{}, logType string) error {
	plr := PushLogsRequest{
		Token:     config.Token,
		StartTime: startTime,
		Logs:      logs,
	}
	d, err := json.Marshal(plr)
	if err != nil {
		return err
	}
	resp, err := request(config.Host, "/glog/push/logs/"+logType, d, config.IgnoreCert)
	if err != nil {
		errCounter++
		return err
	}
	errCounter = 0
	if strings.Trim(strings.ReplaceAll(resp, "\n", ""), " ") != "1" {
		LogInfo("resp: " + resp)
	}
	return nil
}
