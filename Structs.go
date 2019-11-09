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

//PushSyslogRequest request to push
type PushSyslogRequest struct {
	Token     string         `json:"t"`
	StartTime int64          `json:"st"`
	Syslogs   []*SyslogEntry `json:"lgs"`
}
