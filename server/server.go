package main

import (
	"flag"
	"sync"

	"github.com/gitferry/zeitgeber"
	"github.com/gitferry/zeitgeber/log"
)

var algorithm = flag.String("algorithm", "bcb", "synchronizer algorithm")
var id = flag.String("id", "", "ID of the node")
var simulation = flag.Bool("sim", false, "simulation mode")
var isByz = flag.Bool("isByz", false, "this is a Byzantine node")

func replica(id zeitgeber.ID, isByz bool) {
	log.Infof("node %v starting...", id)
	zeitgeber.NewReplica(id, *algorithm, isByz).Run()
}

func main() {
	zeitgeber.Init()

	if *simulation {
		var wg sync.WaitGroup
		wg.Add(1)
		zeitgeber.Simulation()
		for id := range zeitgeber.GetConfig().Addrs {
			isByz := false
			if id.Node() <= zeitgeber.GetConfig().ByzNo {
				isByz = true
			}
			go replica(id, isByz)
		}
		wg.Wait()
	} else {
		replica(zeitgeber.ID(*id), *isByz)
	}
}