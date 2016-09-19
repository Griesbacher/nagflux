package main

import (
	"github.com/kdar/factorlog"
	"os"
)

func main() {
	frmt := `%{Color "red" "ERROR"}%{Color "yellow" "WARN"}%{Color "green" "INFO"}%{Color "cyan" "DEBUG"}%{Color "blue" "TRACE"}[%{Date} %{Time}] [%{SEVERITY}:%{File}:%{Line}] %{Message}%{Color "reset"}`
	log := factorlog.New(os.Stdout, factorlog.NewStdFormatter(frmt))
	log.Error("Severity: Error occurred")
	log.Warn("Severity: Warning!!!")
	log.Info("Severity: I have some info for you")
	log.Debug("Severity: Debug what?")
	log.Trace("Severity: Tracing your IP... muauahaha")
}
