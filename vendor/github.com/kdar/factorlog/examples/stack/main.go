package main

import (
	log "github.com/kdar/factorlog"
)

func call() {
	log.Stack("Stack from func")
}

func main() {
	call()
}
