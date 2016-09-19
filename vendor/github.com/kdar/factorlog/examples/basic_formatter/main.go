package main

import (
	"github.com/kdar/factorlog"
	"os"
)

func main() {
	log := factorlog.New(os.Stdout, factorlog.NewStdFormatter("%{Date} %{Time} %{File}:%{Line} %{Message}"))
	log.Println("Basic formatter")
}
