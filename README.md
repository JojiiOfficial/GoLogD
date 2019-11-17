# GoLogD
GoLogD is a logging daemon for the gologging centralized logging system. It parses logs and pushes them to the [GoLogServer](https://github.com/JojiiOfficial/GoLogServer)

# Logtypes
Currently following logs are supported:<br>
- systemd (syslog/authlog/etc...)
- every file starting with a timestamp in each row (custom logfiles)

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
`token`       The token for the daemon (Need to be added in the 'User' table from the server)<br>
<br>
You can run `./goLogD push` again if you want to check if the config is filled correctly. If the daemon keeps running everything is ok. If not have a look at `/var/log/gologger.log`<br>
<b>Note:</b> You can run `./goLogD install` to create a systemd service automatically.
# Config
\<to be continued\>
