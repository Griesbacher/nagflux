package main

import (
	"os"

	"github.com/kdar/factorlog"
	"github.com/kdar/factorlog-contrib/glog"
)

func main() {
	log := factorlog.New(os.Stdout, glog.NewGlogFormatter())
	log.Print("Glog formatter")
}
