package logging

import (
	"github.com/kdar/factorlog"
	"io"
	"os"
)

const logFormat = "%{Date} %{Time} %{Severity}: %{SafeMessage}"
const logColors = "%{Color \"white\" \"DEBUG\"}%{Color \"magenta\" \"WARN\"}%{Color \"red\" \"CRITICAL\"}"

var singleLogger *factorlog.FactorLog = nil

//Logger Constructor.
func InitLogger(logFile, minSeverity string) {
	var logFormatter factorlog.Formatter
	var targetWriter io.Writer
	var err error
	if logFile == "" {
		logFormatter = factorlog.NewStdFormatter(logColors + logFormat)
		targetWriter = os.Stdout
	} else {
		logFormatter = factorlog.NewStdFormatter(logFormat)
		if _, err := os.Stat(logFile); err == nil {
			targetWriter, err = os.Create(logFile)
		} else {
			targetWriter, err = os.Open(logFile)
		}
	}
	if err != nil {
		panic(err)
	}
	singleLogger = factorlog.New(targetWriter, logFormatter)
	singleLogger.SetMinMaxSeverity(factorlog.StringToSeverity(minSeverity), factorlog.StringToSeverity("PANIC"))
}

//Singelton logger
func GetLogger() *factorlog.FactorLog {
	if singleLogger == nil {
		InitLogger("", "WARN")
	}
	return singleLogger
}
