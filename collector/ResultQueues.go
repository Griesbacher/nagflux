package collector

import "github.com/griesbacher/nagflux/data"

type ResultQueues map[data.Target]chan Printable
