# GoLogD
GoLogD is a logging daemon for the gologging centralized logging system. It parses logs and pushes them to the [GoLogServer](https://github.com/JojiiOfficial/GoLogServer)

# Logtypes
Currently following logs are supported:<br>
- <b>systemd</b> (syslog/authlog/etc...)
- every file starting with a timestamp in each row (<b>custom</b> logfiles)

# Install
Install go 1.13, clone this repository. Then run
```go
go get
go build -o goLogD
```
to compile it. Then run
```go
 ./goLogD push
 ```
 to let the daemon automatically create a config file. You need to change following options:<br>
`host`        The host of the [GoLogServer](https://github.com/JojiiOfficial/GoLogServer)
`token`       The token for the daemon (24bytes!)(Need to be added in the 'User' table from the server)<br>
<br>
You can run `./goLogD push` again if you want to check if the config is filled correctly. If the daemon keeps running everything is ok. If not have a look at `/var/log/gologger.log`<br>
<b>Note:</b> You can run `./goLogD install` to create a systemd service automatically.
# Config
You can add/edit a logfile in the config using `./goLogD addFile -f /var/log/auth.log`. But using this tool you can only set a few options. Here are all config options:<br>
#### Global
`ignoreCert` (bool) ignore invalid TLS certificates<br>
`termsToIgnore` (string array) Don't push a log if it contains at least one of the given keywords (globally)<br>
`LogFiles` (logfile array) Contains options for the files to parse:<br>
#### Logfiles
`file` (string) The logfile (eg /var/log/syslog)<br>
`logType` (string) The type of log (see [Logtypes](https://github.com/JojiiOfficial/GoLogD#logtypes). The keywords for the config are bold)<br>
`filterMode` ("and"/"or") to specify if the given filter must match completely or only a partialy<br>
`hostnameFilter` (string array) filter by hostname(s). Takes a regex.<br>
`messageFilter` (string array) filter by a message(s). Takes a regex.<br>
`tagFilter` (string array) filter by tag(s). Takes a regex.<br>
`logLevesFilter` (int array) filter by loglevels(s).<br>
`termsToIgnore` (string array) don't push logentries containing one (or multiple) keys
`sourceFilter` (string array) filter by sources(s). Takes a regex.<br>
`parseSource` (bool) use second word in custom log as 'source'. You can filter by the source later on.<br>
`customTag` (string) overwrite or set logTag for<br>

## Special options
These options only work for the given logtype:<br>
Syslog:
- Hostname
- Tag

Custom:
- Source
- ParseSource

## Example config
```json
{
	"token": "abcdefghijklmnopqrstuvwxyz",
	"host": "http://192.168.3.11:8081",
	"ignoreCert": false,
	"termsToIgnore": [],
	"LogFiles": [
		{
			"logfile": "/var/log/syslog",
			"logType": "syslog",
			"filterMode": "or",
			"tagFilter": [
				"(?i)(cron|systemd|certbot)"
			],
			"termsToIgnore": [
				"success"
			],
		},
		{
			"logfile": "/var/log/auth.log",
			"logType": "syslog",
			"filterMode": "or",
			"tagFilter": [
				"(?i)(ssh|sudo)"
			],
			"termsToIgnore": [
				"success"
			]
		},
		{
			"logfile": "/var/log/nginx/access.log",
			"logType": "custom",
			"filterMode": "or",
			"parseSource":true,
			"customTag":"revproxy",
		}
	]
}

```
1. Pushes logs from the syslog if the tag says 'cron','systemd' or 'certbot' and the line doesn't contain 'success'.<br>
2. Pushes logs from the auth.log if the tag says 'ssh' or 'sudo' and the line doesn't contain 'success'.<br>
2. Pushes all logs from the access.log (nginx) parsing the second word as src (the IP) and 'revproxy' as a custom tag.<br>
