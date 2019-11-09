package main

import (
	"strconv"
	"strings"
	"time"
)

//PrepareLine prepares a line for parsing
func PrepareLine(line string) []string {
	replaced := strings.ReplaceAll(strings.ReplaceAll(line, "   ", " "), "  ", " ")
	return strings.Split(replaced, " ")
}

//ParseSyslogTime parses only the time value from message
func ParseSyslogTime(line string) (prepared []string, tim time.Time, err error) {
	prepared = PrepareLine(line)
	tim, err = time.Parse(time.Stamp, prepared[0]+" "+prepared[1]+" "+prepared[2])
	if err != nil {
		return
	}
	tim = tim.AddDate(time.Now().Year(), 0, 0)
	return
}

//ParseSyslogMessage parses a message from syslog
func ParseSyslogMessage(line string, startTime int64) *SyslogEntry {
	logentry := &SyslogEntry{}

	splitted, tim, err := ParseSyslogTime(line)
	if err != nil {
		//TODO log
		return logentry
	}
	logentry.Date = (int)(startTime - tim.Unix())
	logentry.Hostname = splitted[3]
	tag := strings.Split(splitted[4], "[")
	if len(tag) == 2 {
		pid, err := strconv.Atoi(strings.Split(tag[1], "]")[0])
		if err == nil {
			logentry.PID = pid
		}
	} else {
		logentry.Tag = tag[0]
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
	for i := start; i < len(splitted); i++ {
		logentry.Message += splitted[i] + " "
	}

	return logentry
}
