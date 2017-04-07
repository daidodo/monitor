package main

import (
	"log"
	"time"

	"github.com/daidodo/overlord/inner"
)

func main() {
	// init log
	log.SetPrefix("[overlord_agent]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	// init shm
	ns, err := inner.Attach(true)
	if err != nil {
		log.Panicf("Cannot attach shm: %v\n", err)
	}
	for _, n := range ns {
		n.Reset()
	}
	// loop
	log.Println("program started")
	for {
		time.Sleep(60 * time.Second)
		for _, n := range ns {
			if n.Attr == 0 {
				break
			}
			// TODO: gather all [attr, value] pairs and send to server
		}
	}
	// exit
	log.Fatalln("program exit!")
}
