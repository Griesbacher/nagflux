package logging

import (
	"github.com/kdar/factorlog"
	"io"
	"os"
)

const logFormat = "%{Date} %{Time} %{Severity}: %{Message}"
const logColors = "%{Color \"white\" \"DEBUG\"}%{Color \"magenta\" \"WARN\"}%{Color \"red\" \"CRITICAL\"}"

var singleLogger *factorlog.FactorLog

//InitLogger Constructor.
func InitLogger(logFile, minSeverity string) {
	var logFormatter factorlog.Formatter
	var targetWriter io.Writer
	var err error
	if logFile == "" {
		logFormatter = factorlog.NewStdFormatter(logColors + logFormat)
		targetWriter = os.Stdout
	} else {
		logFormatter = factorlog.NewStdFormatter(logFormat)
		targetWriter, err = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	}
	if err != nil {
		panic(err)
	}
	singleLogger = factorlog.New(targetWriter, logFormatter)
	singleLogger.SetMinMaxSeverity(factorlog.StringToSeverity(minSeverity), factorlog.StringToSeverity("PANIC"))
}

//GetLogger getsingelton logger
func GetLogger() *factorlog.FactorLog {
	if singleLogger == nil {
		InitLogger("", "WARN")
	}
	return singleLogger
}

//InitTestLogger creates logger for testing
func InitTestLogger() {
	singleLogger = factorlog.New(os.Stderr, factorlog.NewStdFormatter(""))
}
