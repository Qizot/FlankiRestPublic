package websocketchat

import (
	"github.com/sirupsen/logrus"
	"github.com/x-cray/logrus-prefixed-formatter"
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
	logger.Level = logrus.DebugLevel
	logEntry = logger.WithField("prefix","[CHAT]")
}

func Logger() *logrus.Entry {
	return logEntry
}

func RawLogger() *logrus.Logger {
	return logger
}
