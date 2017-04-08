package main

import (
	"log"
	"net"
	"time"

	"github.com/daidodo/overlord/inner"
	"github.com/golang/protobuf/proto"
)

func main() {
	// init log
	log.SetPrefix("[overlord_agent]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	// init shm
	ns, err := inner.Attach(true)
	if err != nil {
		log.Fatalf("Cannot attach shm: %v\n", err)
	}
	for _, n := range ns {
		n.Reset()
	}
	// init net
	conn, err := net.Dial("udp", "collector.overlord.com:9527")
	if err != nil {
		log.Fatalf("Cannot init network: %v\n", err)
	}
	// loop
	log.Println("program started")
	for {
		time.Sleep(2 * time.Second)
		report := &Report{}
		for _, n := range ns {
			if n.Attr == 0 {
				break
			}
			a := Report_Node{Attr: proto.Uint32(n.Attr), Value: proto.Uint64(n.Value)}
			report.Attrs = append(report.Attrs, &a)
		}
		msg, err := proto.Marshal(report)
		if err != nil {
			log.Printf("Cannot marshal report: %v, report=%v\n", err, report)
			continue
		}
		if len(msg) < 1 {
			log.Printf("msg has 0 length for report=%v\n", report)
			continue
		}
		n, err := conn.Write(msg)
		if err != nil {
			log.Printf("Send msg (size=%v) failed: %v\n", len(msg), err)
			continue
		}
		if n != len(msg) {
			log.Printf("%v bytes were sent for msg (size=%v)\n", n, len(msg))
			continue
		}
		//log.Printf("msg (size=%v) were sent to %v", len(msg), conn.RemoteAddr())
	}
	// exit
	log.Fatalln("program exit!")
}
