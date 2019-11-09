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
}

func checkData() (data *Data, err error) {
	data = &Data{
		LastPush: 0,
		Files:    []FileData{},
	}
	_, err = os.Stat(dataFile)
	if err != nil {
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
		if fileData := data.hasFile(configFile.File); fileData != nil {
			watchedFiles = append(watchedFiles, WatchedFile{
				File:           configFile.File,
				RegexWhitelist: configFile.RegexWhitelist,
				RegexBlacklist: configFile.RegexBlacklist,
				LastLogTime:    fileData.LastLogTime,
			})
			dataFiles = append(dataFiles, *fileData)
		} else {
			if _, err := os.Stat(configFile.File); err != nil {
				LogError("Logfile doesn't exist \"" + configFile.File + "\"")
				continue
			}
			watchedFiles = append(watchedFiles, WatchedFile{
				File:           configFile.File,
				RegexWhitelist: configFile.RegexWhitelist,
				RegexBlacklist: configFile.RegexBlacklist,
				LastLogTime:    time.Now().Unix(),
			})
			dataFiles = append(dataFiles, FileData{
				FileName:    configFile.File,
				LastLogTime: time.Now().Unix(),
			})
		}
	}
	data.Files = dataFiles
	return watchedFiles
}

func (data *Data) hasFile(sFile string) (fileData *FileData) {
	for _, file := range data.Files {
		if file.FileName == sFile {
			return &file
		}
	}
	return nil
}
