package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
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
	TagFilter      []string `json:"tagFilter,omitempty"`
	LogLevelFilter []int    `json:"logLevelFilter,omitempty"`
	KeyBlacklist   []string `json:"termsToIgnore,omitempty"`
	MessageFilter  []string `json:"messageFilter,omitempty"`
	SourceFilter   []string `json:"sourceFilter,omitempty"`
	ParseSource    bool     `json:"parseSource,omitempty"`
}

//LogTypes all supported Logtypes
var LogTypes = []string{
	Syslog,
	Custom,
}

const (
	//Syslog syslog conf
	Syslog = "syslog"
	//Custom conf
	Custom = "custom"
)

func checkConfig() (configa *Config, err error) {
	defaultConfig := &Config{
		Token:              "",
		GlobalKeyBlacklist: []string{},
		Files: []FileConfig{
			FileConfig{
				LogType:    "syslog",
				File:       "/var/log/syslog",
				FilterMode: "or",
				KeyBlacklist: []string{
					"success",
				},
				TagFilter: []string{
					"(?i)(cron|systemd)",
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

//Save saves config
func (config *Config) Save() error {
	b, err := json.Marshal(config)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	json.Indent(&out, b, "", "\t")

	return ioutil.WriteFile(configFile, out.Bytes(), 0600)
}

//Validate removes empty fields
func (config *Config) Validate() {
	for h := 0; h < len(config.Files); h++ {
		var fields = []*[]string{
			&config.Files[h].HostnameFilter,
			&config.Files[h].TagFilter,
			&config.Files[h].MessageFilter,
		}
		for i := 0; i < len(fields); i++ {
			var cl []string
			for j := 0; j < len(*fields[i]); j++ {
				cf := strings.Trim((*fields[i])[j], " ")

				if len(cf) > 0 {
					cl = append(cl, cf)
				}
			}
			*fields[i] = cl
		}
	}
}
