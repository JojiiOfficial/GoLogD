package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
)

//ParseSysLogFile parses a syslogFile
func ParseSysLogFile(file string, since int64) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	start := time.Now()
	lines := 0
	for scanner.Scan() {
		if lines >= 1000 {
			break
		}
		logE := ParseSyslogMessage(scanner.Text())
		fmt.Println(logE)
		lines++
	}
	fmt.Println("took ", time.Now().Sub(start), "for ", lines, "lines")

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
