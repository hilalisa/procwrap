package main

import (
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

type config struct {
	executable              string
	hasArgs                 bool
	args                    []string
	outputDebug             bool
	restartOnFailure        bool
	restartPauseMs          int
	maxLogSizeMb            int
	truncateIntervalSeconds int
	logFile                 string
	maxLogAgeDays           int
	maxLogBackups           int
	fatalLogMsgPattern      string
	hasFatalLogMsgPattern   bool
	hasTimeformat           bool
	timeFormat              string
}

func main() {

	viper := viper.New()
	viper.SetConfigFile("./procwrap.toml")
	viper.ReadInConfig()

	conf := config{
		executable:              viper.GetString("executable"),
		hasArgs:                 viper.IsSet("args"),
		args:                    viper.GetStringSlice("args"),
		outputDebug:             viper.GetBool("outputDebug"),
		restartOnFailure:        viper.IsSet("restartPauseMs"),
		restartPauseMs:          viper.GetInt("restartPauseMs"),
		maxLogSizeMb:            viper.GetInt("maxLogSizeMb"),
		truncateIntervalSeconds: viper.GetInt("truncateIntervalSeconds"),
		logFile:                 viper.GetString("logFile"),
		maxLogAgeDays:           viper.GetInt("maxLogAgeDays"),
		maxLogBackups:           viper.GetInt("maxLogBackups"),
		fatalLogMsgPattern:      viper.GetString("fatalLogMsgPattern"),
		hasFatalLogMsgPattern:   viper.IsSet("fatalLogMsgPattern"),
		hasTimeformat:           viper.IsSet("timeformat"),
		timeFormat:              viper.GetString("timeformat"),
	}

	var cmd *exec.Cmd
	if conf.hasArgs {
		cmd = exec.Command(conf.executable, conf.args...)
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

	if err != nil {

		if conf.hasFatalLogMsgPattern {
			timeUtc := time.Now().UTC()
			var timeStr string

			if conf.hasTimeformat {
				timeStr = timeUtc.Format(conf.timeFormat)
			} else {
				timeStr = timeUtc.String()
			}

			hostAddress, _ := os.Hostname()

			fatalMessage := strings.Replace(conf.fatalLogMsgPattern, "$dateTimeUtc", timeStr, -1)
			fatalMessage = strings.Replace(fatalMessage, "$hostIpAddress", hostAddress, -1)
			fatalMessage = strings.Replace(fatalMessage, "$error", err.Error(), -1)
			stdErrWriter.Write([]byte(fatalMessage + "\r\n"))
		}

		if conf.restartOnFailure {
			logger.Close()
			time.Sleep(time.Duration(conf.restartPauseMs) * time.Millisecond)
			main()
		}
	}
}
