package main

import (
	log "github.com/kdar/factorlog"
)

func main() {
	log.SetSeverities(log.INFO | log.WARN)
	log.Info("Severity: will print info.")
	log.Warn("Severity: will print warn.")
	log.Debug("Severity: won't print debug.")

	log.SetMinMaxSeverity(log.TRACE, log.INFO)
	log.Trace("Severity: will print trace.")
	log.Debug("Severity: will print debug.")
	log.Info("Severity: will print info.")
	log.Warn("Severity: won't print warn.")
}
