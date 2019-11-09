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
	GlobalKeyBlacklist []string     `json:"termsToIgnore"`
	Files              []FileConfig `json:"LogFiles"`
}

//FileConfig config for each (log) file
type FileConfig struct {
	File           string   `json:"logfile"`
	RegexWhitelist []string `json:"regexWhitelist"`
	RegexBlacklist []string `json:"regexBlacklist"`
}

func checkConfig() (config *Config, err error) {
	config = &Config{
		Token:              "",
		GlobalKeyBlacklist: []string{},
		Files: []FileConfig{
			FileConfig{
				File:           "/var/log/syslog",
				RegexWhitelist: []string{},
				RegexBlacklist: []string{},
			},
		},
	}
	_, err = os.Stat(configFile)
	if err != nil {
		f, err := os.Create(configFile)
		if err != nil {
			return nil, err
		}
		sdata, err := json.Marshal(config)
		var out bytes.Buffer
		json.Indent(&out, sdata, "", "\t")

		_, err = f.Write(out.Bytes())
		if err != nil {
			return nil, err
		}
		return config, nil
	}
	dat, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(dat, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (config *Config) save() error {
	b, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFile, b, 0600)
}
