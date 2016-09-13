package main

import (
	"encoding/json"
	"flag"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type config struct {
	Executable            string
	HasArgs               bool
	Args                  []string
	OutputDebug           bool
	RestartOnFailure      bool
	RestartPauseMs        int
	MaxLogSizeMb          int
	LogFile               string
	MaxLogAgeDays         int
	MaxLogBackups         int
	FatalLogMsgPattern    string
	HasFatalLogMsgPattern bool
	HasTimeformat         bool
	TimeFormat            string
}

func main() {

	configFile := flag.String("p", "./procwrap.toml", "process definition file, TOML format, typically ./procwrap.toml")
	verboseOutput := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	if *verboseOutput {
		log.Printf("Starting procwrap using def: %s", *configFile)
	}

	viper := viper.New()
	viper.SetConfigFile(*configFile)
	viper.ReadInConfig()

	conf := config{
		Executable:            viper.GetString("executable"),
		HasArgs:               viper.IsSet("args"),
		Args:                  viper.GetStringSlice("args"),
		OutputDebug:           viper.GetBool("outputDebug"),
		RestartOnFailure:      viper.IsSet("restartPauseMs"),
		RestartPauseMs:        viper.GetInt("restartPauseMs"),
		MaxLogSizeMb:          viper.GetInt("maxLogSizeMb"),
		LogFile:               viper.GetString("logFile"),
		MaxLogAgeDays:         viper.GetInt("maxLogAgeDays"),
		MaxLogBackups:         viper.GetInt("maxLogBackups"),
		FatalLogMsgPattern:    viper.GetString("fatalLogMsgPattern"),
		HasFatalLogMsgPattern: viper.IsSet("fatalLogMsgPattern"),
		HasTimeformat:         viper.IsSet("timeformat"),
		TimeFormat:            viper.GetString("timeformat"),
	}

	if *verboseOutput {
		jsonBytes, err := json.MarshalIndent(conf, "", " ")

		if err != nil {
			log.Println(err.Error())
		}
		log.Println(string(jsonBytes))
	}

	var cmd *exec.Cmd
	if conf.HasArgs {
		cmd = exec.Command(conf.Executable, conf.Args...)
	} else {
		cmd = exec.Command(conf.Executable)
	}

	logger := &lumberjack.Logger{
		Filename:   conf.LogFile,
		MaxSize:    conf.MaxLogSizeMb,
		MaxBackups: conf.MaxLogBackups,
		MaxAge:     conf.MaxLogAgeDays,
	}

	stdOutWriter := io.MultiWriter(os.Stdout, logger)
	stdErrWriter := io.MultiWriter(os.Stderr, logger)
	cmd.Stdout = stdOutWriter
	cmd.Stderr = stdErrWriter

	if *verboseOutput {
		log.Println("Starting executable")
	}

	err := cmd.Run()

	if err != nil {

		if conf.HasFatalLogMsgPattern {
			timeUtc := time.Now().UTC()
			var timeStr string

			if conf.HasTimeformat {
				timeStr = timeUtc.Format(conf.TimeFormat)
			} else {
				timeStr = timeUtc.String()
			}

			hostAddress, _ := os.Hostname()

			fatalMessage := strings.Replace(conf.FatalLogMsgPattern, "$dateTimeUtc", timeStr, -1)
			fatalMessage = strings.Replace(fatalMessage, "$hostIpAddress", hostAddress, -1)
			fatalMessage = strings.Replace(fatalMessage, "$error", err.Error(), -1)
			stdErrWriter.Write([]byte(fatalMessage + "\r\n"))
		}

		if conf.RestartOnFailure {
			logger.Close()
			time.Sleep(time.Duration(conf.RestartPauseMs) * time.Millisecond)
			main()
		}
	} else {
		if *verboseOutput {
			log.Println("Executable terminated without error")
		}
	}
}
