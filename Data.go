package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
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
