package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

//ParseSysLogFile parses a syslogFile
func ParseSysLogFile(file string, since int64) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		prepared, tima, _ := ParseSyslogTime(line)
		if tima.Unix() <= since {
			continue
		}
		logE := ParseSyslogMessage(prepared, tima, line, since)
		fmt.Println(*logE)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
