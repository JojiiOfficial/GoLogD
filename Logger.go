package main

import (
	"log"
	"os"
)

//Syslogger logs to syslog
var Syslogger = log.New(os.Stdout, logPrefix, 0)

var logFile = "/var/log/gologger.log"
var lf *os.File

func initLoggerFiles(prefix string) {
	var err error
	lf, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		Syslogger.Printf("Can't open logfile! Exiting...")
		os.Exit(1)
		return
	}
	lf.Truncate(0)
	log.SetOutput(lf)
	log.SetPrefix(prefix)
}

//Log logs log
func Log(level int, msg string) {
	s, err := os.Stat(logFile)
	if err != nil {
		Syslogger.Println("File was Truncated")
		initLoggerFiles(logPrefix)
	}
	if s != nil && s.Size() >= 500000000 {
		Syslogger.Println("Log too big! Truncating...")
		lf.Truncate(0)
		initLoggerFiles(logPrefix)
	}
	log.Printf(
		"%s %s",
		logTypeToString(level),
		msg,
	)
}

//LogCritical logs a very critical error
func LogCritical(msg string) {
	Log(LogCrit, msg)
}

//LogError logs error message
func LogError(msg string) {
	Log(LogErr, msg)
}

//LogInfo logs info message
func LogInfo(msg string) {
	Log(LogInf, msg)
}

//PrintFInfo fprints info
func PrintFInfo(format string, data ...interface{}) {
	go log.Printf(format, data...)
}

const (
	//LogInf log
	LogInf = 1
	//LogErr error log
	LogErr = 2
	//LogCrit critical error log
	LogCrit = 3
)

func logTypeToString(logType int) string {
	switch logType {
	case LogInf:
		{
			return "[info]"
		}
	case LogErr:
		{
			return "[!error!]"
		}
	case LogCrit:
		{
			return "[*!Critical!*]"
		}
	default:
		return "[ ]"
	}
}
