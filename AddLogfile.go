package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/mkideal/cli"
)

type addLogfileT struct {
	cli.Helper
	File      string `cli:"*f,file" usage:"Specify a logfile to add to log"`
	Overwrite bool   `cli:"o,overwrite" usage:"Overwrite an existing logfile" dft:"false"`
}

var addLogCMD = &cli.Command{
	Name:    "addFile",
	Aliases: []string{"af", "addfile", "addlog", "al"},
	Desc:    "adds a new logfile to config",
	Argv:    func() interface{} { return new(addLogfileT) },
	Fn: func(ct *cli.Context) error {
		argv := ct.Argv().(*addLogfileT)
		_, err := os.Stat(argv.File)
		if err != nil {
			fmt.Println("Log doesn't exists!")
			return nil
		}
		config, err := checkConfig()
		if err != nil {
			fmt.Println("You neet to fix your config before")
			return nil
		}

		for _, f := range config.Files {
			if f.File == argv.File {
				if argv.Overwrite {
					fmt.Println("Warning! You will change a logfile that is already configured!")
					break
				}
				fmt.Println("Logfile already in config. Use -o if you want to overwrite it")
				return nil
			}
		}
		reader := bufio.NewReader(os.Stdin)
		addLog(argv.File, reader, config, argv.Overwrite)
		return nil
	},
}

func addLog(logfile string, reader *bufio.Reader, config *Config, overwrite bool) {
	var lts string
	var fileconfig *FileConfig

	if !overwrite {
		fileconfig = &FileConfig{}
	} else {
		for i, fc := range config.Files {
			if fc.File == logfile {
				fileconfig = &config.Files[i]
				break
			}
		}
		if fileconfig == nil {
			fileconfig = &FileConfig{}
			fmt.Println("Logfile not found in config. Ignoring -o")
			overwrite = false
		}
	}
	fileconfig.File = logfile

	for i, typ := range LogTypes {
		lts += strconv.Itoa(i) + " = " + typ + ", "
	}

	lts = lts[:len(lts)-2]
	fmt.Println("Enter \"A\" to abort")
	var curr string
	if overwrite {
		curr = " (" + fileconfig.LogType + ")> "
	}
	i, txt := WaitForMessage("Which kind of logfile do you have selected ["+lts+"]>"+curr, reader)
	if i != 1 && !overwrite || (overwrite && i == -1) {
		return
	}
	if i == 1 {
		logType, err := strconv.Atoi(strings.Trim(txt, " "))
		if err != nil {
			fmt.Println("Not an int!")
			return
		}
		if logType > len(LogTypes)-1 || logType < 0 {
			fmt.Println("Out of range!")
			return
		}

		fileconfig.LogType = LogTypes[logType]
	}

	curr = ""
	if overwrite {
		curr = "(e) > "
		if fileconfig.FilterMode == "or" {
			curr = "(p) > "
		}
	}
	i, txt = WaitForMessage("Does a log need to fit (e)verything or only a (p)art of this filter [e/p]> "+curr, reader)
	if i == -1 || (i == 1 && (txt != "p" && txt != "e")) {
		return
	}
	if i == 1 {
		if txt == "e" {
			fileconfig.FilterMode = "and"
		} else {
			fileconfig.FilterMode = "or"
		}
	} else if !overwrite {
		fileconfig.FilterMode = "and"
	}

	curr = ""
	if overwrite {
		if len(fileconfig.TagFilter) > 0 {
			curr += "';' to delete. (" + regexFilterToKeywords(fileconfig.TagFilter) + ") > "
		}
	}
	i, txt = WaitForMessage("Specify the keyords to filter in the logtag separated by commas > "+curr, reader)
	if i == -1 {
		return
	} else if i == 1 {
		if overwrite {
			if txt == ";" {
				fileconfig.TagFilter = nil
			} else {
				fileconfig.TagFilter = []string{
					keyordsToRegexFilter(txt),
				}
			}
		} else {
			fileconfig.TagFilter = append(fileconfig.TagFilter, keyordsToRegexFilter(txt))
		}
	}

	curr = ""
	if overwrite {
		if len(fileconfig.MessageFilter) > 0 {
			curr += "';' to delete. (" + regexFilterToKeywords(fileconfig.MessageFilter) + ") > "
		}
	}
	i, txt = WaitForMessage("Specify the keyords to filter in logmessage separated by commas > "+curr, reader)
	if i == -1 {
		return
	} else if i == 1 {
		if overwrite {
			if txt == ";" {
				fileconfig.MessageFilter = nil
			} else {
				fileconfig.MessageFilter = []string{
					keyordsToRegexFilter(txt),
				}
			}
		} else {
			fileconfig.MessageFilter = append(fileconfig.MessageFilter, keyordsToRegexFilter(txt))
		}
	}

	curr = ""
	if overwrite {
		if len(fileconfig.KeyBlacklist) > 0 {
			curr += "; to delete. ("
			for _, kw := range fileconfig.KeyBlacklist {
				curr += kw + ","
			}
			if len(curr) > 2 {
				curr = curr[:len(curr)-1]
			}
			curr += ") > "
		}
	}
	i, txt = WaitForMessage("Specify keywords to blacklist separated by comma > "+curr, reader)
	if i == -1 {
		return
	}
	if i == 1 {
		if txt == ";" {
			fileconfig.KeyBlacklist = nil
		} else {
			fileconfig.KeyBlacklist = strings.Split(txt, ",")
		}
	}

	if !overwrite {
		config.Files = append(config.Files, *fileconfig)
	}
	err := config.Save()
	if err != nil {
		fmt.Println("Error saving config: " + err.Error())
	} else {
		fmt.Println("Config saved successfully")
	}
}

func keyordsToRegexFilter(input string) string {
	a := "(?i)("
	for _, f := range strings.Split(input, ",") {
		e := strings.Trim(f, " ")
		if len(e) == 0 {
			continue
		}
		a += e + "|"
	}
	a = a[:len(a)-1] + ")"
	return a
}

func regexFilterToKeywords(inputs []string) string {
	var a string
	for _, input := range inputs {
		input = strings.ReplaceAll(input, "(?i)", "")
		if mtch, _ := regexp.MatchString("(?i)[(]([a-z]*[|]*)*[)]", input); mtch {
			a += strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(input, ")", ""), "(", ""), "|", ",")
		} else {
			a += input + ","
		}
	}
	if strings.HasSuffix(a, ",") {
		a = a[:len(a)-1]
	}
	return a
}
