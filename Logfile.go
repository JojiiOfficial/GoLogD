package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var files = make(map[string]*os.File)
var fileSize = make(map[string]int64)
var flock = sync.RWMutex{}

//ParseLogFile parses a logfile
func ParseLogFile(file string, since int64, callb func([]string, time.Time, int, string)) {
	if verbose > 2 {
		LogInfo("File: " + file)
	}
	flock.RLock()
	wasNil := false
	if _, ok := files[file]; !ok {
		wasNil = true
		var err error
		f, err := os.Open(file)
		if err != nil {
			LogCritical("Couldn't open " + file)
			return
		}
		files[file] = f
		dat, err := os.Stat(file)
		if err == nil {
			fileSize[file] = dat.Size()
		} else if verbose > 1 {
			LogError("Error getting stat of file: " + err.Error())
		}
	}
	dat, _ := os.Stat(file)
	if verbose > 2 {
		LogInfo("New fs: " + strconv.FormatInt(dat.Size(), 10) + " - old fs: " + strconv.FormatInt(fileSize[file], 10) + " was nil: " + strconv.FormatBool(wasNil))
	}
	if !wasNil {
		if (dat.Size()) < fileSize[file] || (dat.Size() == 0 && fileSize[file] == 0) {
			LogInfo("file truncated!")
			f, err := os.Open(file)
			if err != nil {
				LogCritical("Couldn't open " + file)
				return
			}
			files[file] = f
		}
	}

	fileSize[file] = dat.Size()
	flock.RUnlock()

	scanner := bufio.NewScanner(files[file])
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.Trim(line, " ")) == 0 {
			continue
		}
		prepared, tima, timelen, err := ParselogTime(file, line)
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
		callb(prepared, tima, timelen, line)
	}

	if err := scanner.Err(); err != nil {
		LogError("Error scanning: " + err.Error())
	}
}
