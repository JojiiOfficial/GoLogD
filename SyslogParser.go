package main

import (
	"encoding/hex"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/blake2b"
	_ "golang.org/x/crypto/blake2b"
)

//PrepareLine prepares a line for parsing
func PrepareLine(line string) []string {
	replaced := strings.ReplaceAll(strings.ReplaceAll(line, "   ", " "), "  ", " ")
	return strings.Split(replaced, " ")
}

//ParseSyslogTime parses only the time value from message
func ParseSyslogTime(line string) (prepared []string, tim time.Time, err error) {
	prepared = PrepareLine(line)
	tim, err = time.ParseInLocation(time.Stamp, prepared[0]+" "+prepared[1]+" "+prepared[2], time.Now().Location())
	if err != nil {
		return
	}
	tim = tim.AddDate(time.Now().Year(), 0, 0)
	return
}

var regexBins = make(map[string]*regexp.Regexp)

//ParseSyslogMessage parses a message from syslog
func ParseSyslogMessage(splitted []string, tim time.Time, line string, fileconfig *FileConfig, startTime int64) *SyslogEntry {
	hitFilter := false
	logentry := &SyslogEntry{}

	logentry.Date = (int)(tim.Unix() - startTime)

	logentry.Hostname = splitted[3]
	if len(fileconfig.HostnameFilter) > 0 {
		if mr := logRegexMatch(logentry.Hostname, fileconfig.HostnameFilter); !mr && fileconfig.FilterMode == "and" {
			return &SyslogEntry{}
		} else if mr && fileconfig.FilterMode == "or" {
			hitFilter = true
		}
	}

	tag := strings.Split(splitted[4], "[")
	if len(tag) == 2 {
		pid, err := strconv.Atoi(strings.Split(tag[1], "]")[0])
		if err == nil {
			logentry.PID = pid
		}
		logentry.Tag = tag[0]
	} else {
		logentry.Tag = splitted[4]
	}
	if strings.HasSuffix(logentry.Tag, ":") {
		logentry.Tag = logentry.Tag[:len(logentry.Tag)-1]
	}

	if len(fileconfig.TagFilter) > 0 {
		if mr := logRegexMatch(logentry.Tag, fileconfig.TagFilter); !mr && fileconfig.FilterMode == "and" {
			return &SyslogEntry{}
		} else if mr && fileconfig.FilterMode == "or" {
			hitFilter = true
		}
	}
	start := 5

	sp5 := splitted[5]
	if strings.HasPrefix(sp5, "<") && strings.HasSuffix(sp5, ">") {
		start = 6
		switch sp5 {
		case "<info>":
			{
				logentry.LogLevel = 6
			}
		case "<warn>":
			{
				logentry.LogLevel = 4
			}
		}
	}

	if len(fileconfig.LogLevelFilter) > 0 {
		if mr := isInIntArray(logentry.LogLevel, fileconfig.LogLevelFilter); !mr && fileconfig.FilterMode == "and" {
			return &SyslogEntry{}
		} else if mr && fileconfig.FilterMode == "or" {
			hitFilter = true
		}
	}

	for i := start; i < len(splitted); i++ {
		logentry.Message += splitted[i] + " "
	}

	if len(fileconfig.MessageFilter) > 0 {
		if mr := logRegexMatch(logentry.Message, fileconfig.MessageFilter); !mr && fileconfig.FilterMode == "and" {
			return &SyslogEntry{}
		} else if mr && fileconfig.FilterMode == "or" {
			hitFilter = true
		}
	}
	if fileconfig.FilterMode == "or" && !hitFilter {
		return &SyslogEntry{}
	}
	return logentry
}

func isInIntArray(src int, arr []int) bool {
	for _, i := range arr {
		if i == src {
			return true
		}
	}
	return false
}

func logRegexMatch(src string, pattern []string) bool {
	for _, filter := range pattern {
		ekey := blake2b.Sum512([]byte(filter))
		key := hex.EncodeToString(ekey[:])
		patt, ok := regexBins[key]
		if !ok {
			var err error
			regexBins[key], err = regexp.Compile(filter)
			if err != nil {
				LogCritical("Error in regex: " + filter)
				os.Exit(1)
				return false
			}
			patt = regexBins[key]
		}
		if patt.MatchString(src) {
			return true
		}
	}
	return false
}
