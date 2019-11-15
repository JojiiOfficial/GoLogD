package main

import (
	"encoding/hex"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/blake2b"
	_ "golang.org/x/crypto/blake2b"
)

type timeLen struct {
	len, timeid int
}

//PrepareLine prepares a line for parsing
func PrepareLine(line string) []string {
	replaced := strings.ReplaceAll(strings.ReplaceAll(line, "   ", " "), "  ", " ")
	return strings.Split(replaced, " ")
}

var timeFormatCache = make(map[string]*timeLen)

func parseStamp(src []string) (time.Time, error) {
	if len(src) > 2 {
		return time.ParseInLocation(time.Stamp, src[0]+" "+src[1]+" "+src[2], time.Now().Location())
	}
	return time.Now(), errors.New("not enough data")
}

func parseNginx(src []string) (time.Time, error) {
	if len(src) > 1 {
		return time.Parse("_2/Jan/2006:15:04:05 -0700", src[0]+" "+src[1])
	}
	return time.Now(), errors.New("not enough data")
}

func detectTF(src []string) (*timeLen, time.Time, error) {
	var err error
	var t time.Time
	if len(src) >= 3 {
		t, err = parseStamp(src)
		if err == nil {
			return &timeLen{len: 3, timeid: 1}, t, nil
		}
	}
	if len(src) > 1 {
		t, err = parseNginx(src)
		if err == nil {
			return &timeLen{len: 2, timeid: 2}, t, nil
		}
	}
	return nil, t, err
}

func getTimeFormat(file string, src []string) (time.Time, int, error) {
	tf, has := timeFormatCache[file]
	if !has {
		dtf, t, err := detectTF(src)
		if err != nil {
			return time.Now(), dtf.len, err
		}
		timeFormatCache[file] = dtf
		return t, 0, nil
	}

	if tf.len <= len(src) {
		if tf.timeid == 1 {
			time, err := parseStamp(src)
			return time, tf.len, err
		} else if tf.timeid == 2 {
			time, err := parseNginx(src)
			return time, tf.len, err
		}
	}
	return time.Now(), 0, errors.New("Couldn't parse time")
}

//ParselogTime parses only the time value from message
func ParselogTime(file, line string) (prepared []string, tim time.Time, timeLen int, err error) {
	prepared = PrepareLine(line)
	tim, timeLen, err = getTimeFormat(file, prepared)
	if tim.Year() == 0 {
		tim = tim.AddDate(time.Now().Year(), 0, 0)
	}
	return
}

var regexBins = make(map[string]*regexp.Regexp)

//ParseSyslogMessage parses a message from syslog
func ParseSyslogMessage(splitted []string, tim time.Time, line string, fileconfig *FileConfig, startTime int64) *SyslogEntry {
	sFilterMode := strings.Trim(fileconfig.FilterMode, " ")
	hitFilter := false

	filterMode := 1
	if sFilterMode == "or" {
		filterMode = 0
	}

	logentry := &SyslogEntry{}

	logentry.Date = (int)(tim.Unix() - startTime)

	logentry.Hostname = splitted[3]

	if len(fileconfig.HostnameFilter) > 0 {
		if mr := logRegexMatch(logentry.Hostname, fileconfig.HostnameFilter); !mr && filterMode == 1 {
			return &SyslogEntry{}
		} else if mr && filterMode == 0 {
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

	if isOwnLogEntry(logentry.Tag) {
		return &SyslogEntry{}
	}

	if len(fileconfig.TagFilter) > 0 {
		if mr := logRegexMatch(logentry.Tag, fileconfig.TagFilter); !mr && filterMode == 1 {
			return &SyslogEntry{}
		} else if mr && filterMode == 0 {
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
		if mr := isInIntArray(logentry.LogLevel, fileconfig.LogLevelFilter); !mr && filterMode == 1 {
			return &SyslogEntry{}
		} else if mr && filterMode == 0 {
			hitFilter = true
		}
	}

	for i := start; i < len(splitted); i++ {
		logentry.Message += splitted[i] + " "
	}

	if isOwnLogEntry(logentry.Message) {
		return &SyslogEntry{}
	}

	if len(fileconfig.MessageFilter) > 0 {
		if mr := logRegexMatch(logentry.Message, fileconfig.MessageFilter); !mr && filterMode == 1 {
			return &SyslogEntry{}
		} else if mr && filterMode == 0 {
			hitFilter = true
		}
	}
	if filterMode == 0 && !hitFilter && (len(fileconfig.HostnameFilter) > 0 || len(fileconfig.TagFilter) > 0 || len(fileconfig.LogLevelFilter) > 0 || len(fileconfig.MessageFilter) > 0) {
		return &SyslogEntry{}
	}
	if len(fileconfig.KeyBlacklist) > 0 {
		for _, key := range fileconfig.KeyBlacklist {
			if len(strings.Trim(key, " ")) == 0 {
				continue
			}
			if strings.Contains(line, key) {
				return &SyslogEntry{}
			}
		}
	}

	return logentry
}

func isOwnLogEntry(src string) bool {
	src = strings.ToLower(src)
	own := strings.ToLower(serviceName)
	return src == own || strings.Contains(src, own)
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

func parseCustomLogMessage(splitted []string, timea time.Time, fileconfig *FileConfig, line string, timelen int, startTime int64) *CustomLogEntry {
	if len(splitted) < timelen+1 {
		LogError("Log too short to parse")
		return nil
	}

	sFilterMode := strings.Trim(fileconfig.FilterMode, " ")
	hitFilter := false
	filterMode := 1
	if sFilterMode == "or" {
		filterMode = 0
	}
	logentry := &CustomLogEntry{}
	logentry.Date = (int)(timea.Unix() - startTime)
	start := timelen
	if fileconfig.ParseSource {
		start++
		logentry.Source = splitted[timelen]
	}
	for i := start; i < len(splitted); i++ {
		logentry.Message += splitted[i] + " "
	}

	if isOwnLogEntry(logentry.Message) {
		return &CustomLogEntry{}
	}

	if len(fileconfig.MessageFilter) > 0 {
		if mr := logRegexMatch(logentry.Message, fileconfig.MessageFilter); !mr && filterMode == 1 {
			return &CustomLogEntry{}
		} else if mr && filterMode == 0 {
			hitFilter = true
		}
	}

	if len(fileconfig.SourceFilter) > 0 && fileconfig.ParseSource {
		if mr := logRegexMatch(logentry.Source, fileconfig.SourceFilter); !mr && filterMode == 1 {
			return &CustomLogEntry{}
		} else if mr && filterMode == 0 {
			hitFilter = true
		}
	}

	if filterMode == 0 && !hitFilter && (len(fileconfig.SourceFilter) > 0 || len(fileconfig.MessageFilter) > 0) {
		return &CustomLogEntry{}
	}
	if len(fileconfig.KeyBlacklist) > 0 {
		for _, key := range fileconfig.KeyBlacklist {
			if len(strings.Trim(key, " ")) == 0 {
				continue
			}
			if strings.Contains(line, key) {
				return &CustomLogEntry{}
			}
		}
	}
	return logentry
}
