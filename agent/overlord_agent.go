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
	defer conn.Close()
	// loop
	log.Println("program started")
	for {
		time.Sleep(60 * time.Second)
		report := &inner.AgentReport{}
		// interfaces & ips
		if ifs, err := net.Interfaces(); err == nil {
			for _, i := range ifs {
				if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
					continue
				}
				var ip net.IP
				as, err := i.Addrs()
				if err != nil {
					log.Printf("Cannot get addrs for interface %v: %v", i.Name, err)
					continue
				}
				for _, a := range as {
					ipn, ok := a.(*net.IPNet)
					if !ok {
						continue
					}
					if ip4 := ipn.IP.To4(); ip4 != nil {
						ip = ip4
						break
					} else if ip == nil {
						ip = ipn.IP
					}
				}
				if ip == nil {
					continue
				}
				a := &inner.AgentReport_Addr{Mac: i.HardwareAddr, Ip: ip}
				report.Addrs = append(report.Addrs, a)
			}
		} else {
			log.Printf("Cannot get interfaces: %v", err)
		}
		// attrs
		for _, n := range ns {
			if n.Attr == 0 {
				break
			}
			// atomic operation
			a := &inner.AgentReport_Node{Attr: proto.Uint32(n.Attr), Value: proto.Uint64(n.Value)}
			report.Attrs = append(report.Attrs, a)
		}
		msg, err := proto.Marshal(report)
		if err != nil {
			log.Printf("Cannot marshal report: %v, report=%v\n", err, report)
			continue
		}
		if len(msg) < 1 {
			log.Printf("msg has 0 size for report=%v\n", report)
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
		// debug
		log.Printf("msg (size=%v) were sent to %v, report=%v", len(msg), conn.RemoteAddr(), report)
	}
	// exit
	log.Fatalln("program exit!")
}
