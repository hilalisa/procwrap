### procwrap

procwrap is a command that does the following:

- start a process (any executable)
- proxy stdout & stderr of the child process to the stdout & stderr of the wrapper as well as a log file
- log the error on child process failure to stderr and log file
- restart child process on failure (optional)
- truncate log file 

see [procwrap.toml](procwrap.toml) for config options



