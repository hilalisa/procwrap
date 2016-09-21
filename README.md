### procwrap

procwrap is a command that does the following:

- start a process (any executable)
- start a HTTP server to return status code 200 if the process is running or 500 (see healthCheckPort in the config)
- proxy stdout & stderr of the child process to the stdout & stderr of the wrapper as well as a log file
- log the error on child process failure to stderr and log file
- restart child process on failure (optional)
- truncate log file (uses gopkg.in/natefinch/lumberjack.v2)

####command line args

- v : enables verbose output 
- p : process definition file, e.g. procwrap.toml (uses github.com/spf13/viper so can be toml or yml)

Example process definition file:

```
  executable="ping"
  args=["127.0.0.1"]
  restartPauseMs=2000
  maxLogSizeMb=64
  logFile="log.txt"
  maxLogAgeDays=28
  maxLogBackups=2
  fatalLogMsgPattern="{\"TimeUtc\": \"$dateTimeUtc\",\"ServiceKey\": \"MyServiceName\",\"Title\": \"MyService FATAL ERROR: $error\",\"HostAddress\": \"$hostIpAddress\"}"
  timeFormat="2006-01-02 15:04:05"
  healthCheckPort=4004
  ```
  
Note the replacement tokens in the fatalLogMsgPattern:

- $dateTimeUtc
- $hostIpAddress
- $error


