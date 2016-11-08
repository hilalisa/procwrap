package main

import (
	"encoding/json"
	"flag"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
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
	HealthCheckPort       int
	LogFilePerms          string
	LogFilePermsIsSet     bool
}

var (
	configFile       *string
	verboseOutput    *bool
	processHealthy   bool
	conf             config
	lastProcessError error
)

func main() {

	configFile = flag.String("p", "./procwrap.toml", "process definition file, TOML format, typically ./procwrap.toml")
	verboseOutput = flag.Bool("v", false, "Verbose output")
	flag.Parse()

	writeVerbose("Starting procwrap using def: " + *configFile)

	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		log.Printf("process definition file not found: " + *configFile)
	}

	readConfFile()

	if conf.HealthCheckPort > 0 {
		writeVerbose("Starting healthcheck endpoing on port " + strconv.Itoa(conf.HealthCheckPort))
		http.HandleFunc("/", healthCheckHandler)
		go http.ListenAndServe(":"+strconv.Itoa(conf.HealthCheckPort), nil)
	}

	if conf.LogFilePermsIsSet {

		if _, err := os.Stat(conf.LogFile); os.IsNotExist(err) {
			fileMode, err := strconv.Atoi("0" + conf.LogFilePerms)
			if err == nil {
				writeVerbose("Creatimg file " + conf.LogFilePerms)
				err = ioutil.WriteFile(conf.LogFile, make([]byte, 0), os.FileMode(fileMode))
			}
			if err != nil {
				log.Printf("Unable to create log file " + conf.LogFile)
				log.Printf(err.Error())
				return
			}
		}

		writeVerbose("Setting file perms: chmod " + conf.LogFilePerms + " " + conf.LogFile)
		err := exec.Command("chmod", conf.LogFilePerms, conf.LogFile).Run()
		if err != nil {
			log.Printf("Unable to set log file perms " + conf.LogFile)
			log.Printf(err.Error())
			return
		}
	}

	exeProc()
}

func exeProc() {

	readConfFile()

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

	writeVerbose("Starting executable")

	lastProcessError = cmd.Start()

	if lastProcessError == nil {

		processHealthy = true

		lastProcessError = cmd.Wait()
	}

	if lastProcessError != nil {

		processHealthy = false

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
			fatalMessage = strings.Replace(fatalMessage, "$error", lastProcessError.Error(), -1)
			stdErrWriter.Write([]byte(fatalMessage + "\r\n"))
		}

		if conf.RestartOnFailure {
			logger.Close()
			time.Sleep(time.Duration(conf.RestartPauseMs) * time.Millisecond)
			exeProc()
		}
	} else {
		writeVerbose("Executable terminated without error")
	}
}

func writeVerbose(msg string) {
	if *verboseOutput {
		log.Println(msg)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if processHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func readConfFile() {

	viper := viper.New()
	viper.SetConfigFile(*configFile)
	viper.ReadInConfig()

	conf = config{
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
		HealthCheckPort:       viper.GetInt("healthCheckPort"),
		LogFilePermsIsSet:     viper.IsSet("LogFilePerms"),
	}

	if conf.LogFilePermsIsSet {
		conf.LogFilePerms = viper.GetString("LogFilePerms")
	}

}
