package utils

import (
	"github.com/sirupsen/logrus"
	"github.com/x-cray/logrus-prefixed-formatter"
	"os"
)

var (
	logger *logrus.Logger
	logEntry *logrus.Entry
)

func init() {
	logger = logrus.New()
	customFormatter := new(prefixed.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.ForceColors = true
	customFormatter.ForceFormatting  = true
	customFormatter.FullTimestamp = true
	logger.SetFormatter(customFormatter)
	logEntry = logger.WithField("prefix", "[AUTH SERVER]")

	if debug := os.Getenv("DEBUG"); debug == "true" {
		logger.Level = logrus.DebugLevel
	}

}

func SetLoggerDebugMode(debug bool) {
	if debug {
		logger.Level = logrus.DebugLevel
		logEntry = logger.WithField("prefix", "[AUTH SERVER]")
	} else {
		logger.Level = logrus.InfoLevel
		logEntry = logger.WithField("prefix", "[AUTH SERVER]")
	}
}

func SetEnvDebug() {
	if debug, found := os.LookupEnv("debug"); found && debug == "true" {
		logger.Level = logrus.DebugLevel
		logEntry = logger.WithField("prefix", "[AUTH SERVER]")
	}
}


func AuthLoggerEntry() *logrus.Entry {
	return logEntry
}

func AuthLogger() *logrus.Logger {
	return logger
}



type GormLogger struct {}

func (*GormLogger) Print(v ...interface{}) {
	log := AuthLogger()
	if v[0] == "sql" {
		log.WithFields(logrus.Fields{"prefix": "[DATABASE INTERNAL]"}).Debug(v[3])
	}
	if v[0] == "log" {
		log.WithFields(logrus.Fields{"prefix": "[DATABASE]"}).Debug(v[2])
	}
}