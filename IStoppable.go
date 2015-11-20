package main

//Stoppable represents every daemonlike struct which can be stopped
type Stoppable interface {
	Stop()
}
