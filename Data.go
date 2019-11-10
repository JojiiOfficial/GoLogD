package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

var dataFile = "./data.json"

//Data deamon data
type Data struct {
	LastPush int64      `json:"lastPush"`
	Files    []FileData `json:"files"`
}

//FileData data for each File
type FileData struct {
	FileName    string `json:"file"`
	LastLogTime int64  `json:"lastLogTime"`
	LogType     string `json:"logtype"`
}

func checkData() (data *Data, err error) {
	data = &Data{
		LastPush: 0,
		Files:    []FileData{},
	}
	st, err := os.Stat(dataFile)
	if err != nil || st.Size() == 0 {
		f, err := os.Create(dataFile)
		if err != nil {
			return nil, err
		}
		sdata, err := json.Marshal(data)
		_, err = f.Write(sdata)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	dat, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(dat, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

//Save saves data struct
func (data *Data) Save() error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dataFile, b, 0600)
}

//Validate removes all not existing files from data
func (data *Data) Validate() {
	var files []FileData
	for _, file := range data.Files {
		logFile := file.FileName
		if _, err := os.Stat(logFile); err != nil {
			LogError("Logfile \"" + logFile + "\" Doesn't exists!")
			continue
		}
		files = append(files, file)
	}
	data.Files = files
}

//MergeWithConfig adds files from config to data and removes FileData from Data if logfile is not in config anymore
func (data *Data) MergeWithConfig(config Config) (watchedFiles []WatchedFile) {
	var dataFiles []FileData
	for _, configFile := range config.Files {
		if fileData := data.hasFile(configFile.File); fileData != (FileData{}) {
			if watchedFilesHasLog(watchedFiles, configFile.File) {
				LogCritical("Logfile \"" + configFile.File + "\" is configured twice!")
				os.Exit(1)
				return
			}
			if !validateLogType(configFile.LogType) {
				LogError("Logtype \"" + configFile.LogType + "\" is not supported! Ignoring logfile \"" + fileData.FileName + "\"")
				continue
			}
			fileData.LogType = configFile.LogType
			watchedFiles = append(watchedFiles, WatchedFile{
				File:           configFile.File,
				HostnameFilter: configFile.HostnameFilter,
				TagFilter:      configFile.TagFilter,
				LogLevelFilter: configFile.LogLevelFilter,
				MessageFilter:  configFile.MessageFilter,
				FileData:       &fileData,
			})
			dataFiles = append(dataFiles, fileData)
		} else {
			if watchedFilesHasLog(watchedFiles, configFile.File) {
				LogCritical("Logfile \"" + configFile.File + "\" is configured twice!")
				os.Exit(1)
				return
			}
			if _, err := os.Stat(configFile.File); err != nil {
				LogError("Logfile doesn't exist \"" + configFile.File + "\"")
				continue
			}
			if !validateLogType(configFile.LogType) {
				LogError("Logtype \"" + configFile.LogType + "\" is not supported! Ignoring logfile \"" + fileData.FileName + "\"")
				continue
			}
			fileData := &FileData{
				FileName:    configFile.File,
				LastLogTime: time.Now().Unix(),
				LogType:     configFile.LogType,
			}
			dataFiles = append(dataFiles, *fileData)

			watchedFiles = append(watchedFiles, WatchedFile{
				File:           configFile.File,
				HostnameFilter: configFile.HostnameFilter,
				TagFilter:      configFile.TagFilter,
				LogLevelFilter: configFile.LogLevelFilter,
				MessageFilter:  configFile.MessageFilter,
				FileData:       fileData,
			})
		}
	}
	data.Files = dataFiles
	return watchedFiles
}

func (data *Data) hasFile(sFile string) (fileData FileData) {
	for _, file := range data.Files {
		if file.FileName == sFile {
			return file
		}
	}
	return FileData{}
}

func validateLogType(logtype string) bool {
	for _, logT := range LogTypes {
		if logT == logtype {
			return true
		}
	}
	return false
}

func watchedFilesHasLog(watchedFiles []WatchedFile, logFile string) bool {
	for _, wf := range watchedFiles {
		if wf.File == logFile {
			return true
		}
	}
	return false
}
