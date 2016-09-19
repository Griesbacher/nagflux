package main

import (
	"github.com/kdar/factorlog"
	"os"
)

func main() {
	log := factorlog.New(os.Stdout, factorlog.NewStdFormatter(`%{Color "magenta"}[%{Date} %{Time}] %{Color "cyan"}[%{SEVERITY}:%{File}:%{Line}] %{Color "yellow"}%{Message}%{Color "reset"}`))
	log.Println("Color")
}
