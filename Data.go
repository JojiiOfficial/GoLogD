package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var dataFile = "./data.json"

//Data deamon data
type Data struct {
	LastPush int64 `json:"lastPush"`
}

func checkData() (data *Data, err error) {
	data = &Data{}
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
	err = ioutil.WriteFile(dataFile, b, 0600)
	return err
}
