package main

import (
	"log"
	"net"

	"github.com/daidodo/overlord/inner"
	"github.com/golang/protobuf/proto"
)

func main() {
	// init log
	log.SetPrefix("[overlord_collector]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	// init net
	addr, err := net.ResolveUDPAddr("udp", "localhost:9527")
	if err != nil {
		log.Fatalf("Cannot resolve network addr: %v\n", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Cannot create udp socket: %v\n", err)
	}
	defer conn.Close()
	log.Print("program started")
	// loop
	for buf := new([65536]byte); ; {
		n, addr, err := conn.ReadFromUDP(buf[:])
		if err != nil {
			log.Printf("ReadFromUDP failed: %v", err)
			continue
		}
		if n < 1 {
			log.Printf("ReadFromUDP() returns %v from %v", n, addr)
			continue
		}
		report := &inner.AgentReport{}
		if err = proto.Unmarshal(buf[:n], report); err != nil {
			log.Printf("Invalid data from %v: %v", addr, err)
			continue
		}
		go process(report, addr)
	}
	log.Fatal("program exited")
}

func process(report *inner.AgentReport, addr *net.UDPAddr) {
	log.Printf("process report=%v from %v", report, addr)
}
