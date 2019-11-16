package main

// ------------------ REST structs ----------------------

//SyslogEntry a log entry in the syslog
type SyslogEntry struct {
	Date     int    `json:"d"`
	Hostname string `json:"h"`
	Tag      string `json:"t"`
	PID      int    `json:"p"`
	LogLevel int    `json:"l"`
	Message  string `json:"m"`
}

//CustomLogEntry a log entry from a custom file
type CustomLogEntry struct {
	Date     int    `json:"d"`
	Message  string `json:"m"`
	Source   string `json:"s"`
	Hostname string `json:"h"`
	Tag      string `json:"t,omitempty"`
}

// ------------------ LOCAL structs ----------------------

//WatchedFile struct with obj from Config and Data
type WatchedFile struct {
	File           string
	HostnameFilter []string
	TagFilter      []string
	LogLevelFilter []int
	MessageFilter  []string
	FileData       *FileData
}

//PushLogsRequest request to push syslog
type PushLogsRequest struct {
	Token     string      `json:"t"`
	StartTime int64       `json:"st"`
	Logs      interface{} `json:"lgs"`
}
