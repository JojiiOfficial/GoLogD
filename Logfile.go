package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

var f *os.File
var files = make(map[string]*os.File)

//ParseSysLogFile parses a syslogFile
func ParseSysLogFile(file string, since int64) {
	wasNil := false
	if _, ok := files[file]; !ok {
		wasNil = true
		var err error
		f, err = os.Open(file)
		if err != nil {
			LogCritical("Couldn't open " + file)
			return
		}
		files[file] = f
	}

	scanner := bufio.NewScanner(files[file])
	for scanner.Scan() {
		line := scanner.Text()
		prepared, tima, _ := ParseSyslogTime(line)
		b := tima.Unix() < since
		if wasNil {
			b = tima.Unix() <= since
		}
		if b {
			continue
		}
		logE := ParseSyslogMessage(prepared, tima, line, since)
		fmt.Println(*logE)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
