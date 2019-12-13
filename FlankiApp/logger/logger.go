package logger

import (
	//"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"sync"
)

var loggerInstance *logrus.Logger
var loggerOnce sync.Once

func initialize(logger *logrus.Logger) *logrus.Logger {
	logger = logrus.New()
	customFormatter := new(prefixed.TextFormatter)
	customFormatter.TimestampFormat = "02-01-2006 15:04:05"
	customFormatter.ForceColors = true
	customFormatter.ForceFormatting  = true
	customFormatter.FullTimestamp = true

	logger.SetFormatter(customFormatter)
	logger.SetReportCaller(true)

	//filenameHook := filename.NewHook()
	//filenameHook.Field = "custom_source_field" // Customize source field name
	//logger.AddHook(filenameHook)

	if debug := os.Getenv("DEBUG"); debug == "true" {
		logger.Level = logrus.DebugLevel
	}

	return logger
}


func GetGlobalLogger() *logrus.Logger {
	loggerOnce.Do(func() {
		loggerInstance = initialize(loggerInstance)
	})
	return loggerInstance
}

func SetGlobalDebug(debug bool) {
	if debug {
		loggerInstance.Level = logrus.DebugLevel
	} else {
		loggerInstance.Level = logrus.InfoLevel
	}
}
