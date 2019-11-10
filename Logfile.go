package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"
)

var f *os.File
var files = make(map[string]*os.File)
var fileSize = make(map[string]int64)
var flock = sync.RWMutex{}

//ParseSysLogFile parses a syslogFile
func ParseSysLogFile(file string, fileConfig *FileConfig, since int64) []*SyslogEntry {
	wasNil := false
	if _, ok := files[file]; !ok {
		wasNil = true
		var err error
		f, err = os.Open(file)
		if err != nil {
			LogCritical("Couldn't open " + file)
			return nil
		}
		flock.RLock()
		files[file] = f
		flock.RUnlock()
		dat, _ := os.Stat(file)
		fileSize[file] = dat.Size()
	}
	if !wasNil {
		dat, _ := os.Stat(file)
		if dat.Size() < fileSize[file] {
			LogInfo("file truncated!")
			fileSize[file] = dat.Size()
			var err error
			f, err = os.Open(file)
			if err != nil {
				LogCritical("Couldn't open " + file)
				return nil
			}
			flock.RLock()
			files[file] = f
			flock.RUnlock()
		}
	}

	syslogEntries := []*SyslogEntry{}
	scanner := bufio.NewScanner(files[file])
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.Trim(line, " ")) == 0 {
			continue
		}
		prepared, tima, err := ParseSyslogTime(line)
		if prepared == nil || err != nil {
			if err != nil {
				LogError(err.Error())
			}
			continue
		}
		b := tima.Unix() < since
		if wasNil {
			b = tima.Unix() <= since
		}
		if b {
			continue
		}
		loge := ParseSyslogMessage(prepared, tima, line, fileConfig, since)
		if loge != nil && *loge != (SyslogEntry{}) {
			syslogEntries = append(syslogEntries, loge)
		} else if loge == nil {
			LogInfo("Couldn't parse " + file)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return syslogEntries
}
