package main

import (
	"github.com/spf13/viper"
	"os/exec"
	"io"
	"os"
	"time"
	"strings"
	"gopkg.in/natefinch/lumberjack.v2"
	"fmt"
)

type config struct {
	executable              string
	hasArgs                 bool
	args                    string
	outputDebug             bool
	restartOnFailure        bool
	restartPauseMs          int
	maxLogSizeMb            int
	truncateIntervalSeconds int
	logFile                 string
	maxLogAgeDays           int
	maxLogBackups           int
	fatalLogMsgPattern string
}


func main() {

	viper := viper.New()
	viper.SetConfigFile("./procwrap.toml")
	viper.ReadInConfig()

	conf := config{
		executable: viper.GetString("executable"),
		hasArgs: viper.IsSet("args"),
		args: viper.GetString("args"),
		outputDebug: viper.GetBool("outputDebug"),
		restartOnFailure: viper.IsSet("restartPauseMs"),
		restartPauseMs: viper.GetInt("restartPauseMs"),
		maxLogSizeMb: viper.GetInt("maxLogSizeMb"),
		truncateIntervalSeconds: viper.GetInt("truncateIntervalSeconds"),
		logFile: viper.GetString("logFile"),
		maxLogAgeDays: viper.GetInt("maxLogAgeDays"),
		maxLogBackups: viper.GetInt("maxLogBackups"),
		fatalLogMsgPattern: viper.GetString("fatalLogMsgPattern"),
	}

	var cmd *exec.Cmd
	if conf.hasArgs{
		cmd = exec.Command(conf.executable, strings.Split(conf.args, ",")...)
	} else {
		cmd = exec.Command(conf.executable)
	}

	logger := &lumberjack.Logger{
		Filename:   conf.logFile,
		MaxSize:    conf.maxLogSizeMb,
		MaxBackups: conf.maxLogBackups,
		MaxAge:     conf.maxLogAgeDays,
	}

	stdOutWriter := io.MultiWriter(os.Stdout, logger)
	stdErrWriter := io.MultiWriter(os.Stderr, logger)
	cmd.Stdout = stdOutWriter
	cmd.Stderr = stdErrWriter

	err := cmd.Run()

	if err!=nil {
		stdErrWriter.Write(([]byte(fmt.Sprintf(conf.fatalLogMsgPattern, err.Error()) + "\r\n")))
		if conf.restartOnFailure {
			logger.Close()
			time.Sleep(time.Duration(conf.restartPauseMs) * time.Millisecond)
			main()
		}
	}
}
