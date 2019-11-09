package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
)

var configFile = "config.json"

//Config deamons config file
type Config struct {
	Token              string       `json:"token"`
	Host               string       `json:"host"`
	IgnoreCert         bool         `json:"ignoreCert"`
	GlobalKeyBlacklist []string     `json:"termsToIgnore"`
	Files              []FileConfig `json:"LogFiles"`
}

//FileConfig config for each (log) file
type FileConfig struct {
	File           string   `json:"logfile"`
	LogType        string   `json:"logType"`
	FilterMode     string   `json:"filterMode"`
	HostnameFilter []string `json:"hostnameFilter,omitempty"`
	TagFilter      []string `json:"tagFilner,omitempty"`
	LogLevelFilter []int    `json:"logLevelFilter,omitempty"`
	MessageFilter  []string `json:"MessageFilter,omitempty"`
}

//LogTypes all supported Logtypes
var LogTypes = []string{
	"syslog",
}

func checkConfig() (configa *Config, err error) {
	defaultConfig := &Config{
		Token:              "",
		GlobalKeyBlacklist: []string{},
		Files: []FileConfig{
			FileConfig{
				LogType:    "syslog",
				File:       "/var/log/syslog",
				FilterMode: "or",
				HostnameFilter: []string{
					"(?i)root",
				},
				TagFilter: []string{
					"(?i)gologd",
				},
				LogLevelFilter: []int{},
				MessageFilter: []string{
					"",
				},
			},
		},
	}
	_, err = os.Stat(configFile)
	if err != nil {
		f, err := os.Create(configFile)
		if err != nil {
			return nil, err
		}
		sdata, err := json.Marshal(defaultConfig)
		var out bytes.Buffer
		json.Indent(&out, sdata, "", "\t")

		_, err = f.Write(out.Bytes())
		if err != nil {
			return nil, err
		}
		return defaultConfig, nil
	}
	configa = &Config{}
	dat, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(dat, &configa)
	if err != nil {
		return nil, err
	}
	return configa, nil
}

func (config *Config) save() error {
	b, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFile, b, 0600)
}
