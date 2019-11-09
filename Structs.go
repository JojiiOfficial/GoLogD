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
