package main

import (
	"log"
	"net"
	"sync/atomic"
	"time"

	sigar "github.com/cloudfoundry/gosigar"
	"github.com/daidodo/overlord/attr"
	"github.com/daidodo/overlord/inner"
	"github.com/golang/protobuf/proto"
)

const kSleep = 10 * time.Second

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
	// system stats
	go systemStats()
	// loop
	log.Println("program started")
	for {
		time.Sleep(kSleep)
		report := &inner.AgentReport{}
		// addrs
		if ifs, err := net.Interfaces(); err == nil {
			for _, i := range ifs {
				if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
					continue
				}
				as, err := i.Addrs()
				if err != nil {
					log.Printf("Cannot get addrs for interface %v: %v", i.Name, err)
					continue
				}
				var ips []string
				for _, a := range as {
					ipn, ok := a.(*net.IPNet)
					if !ok {
						continue
					}
					ip := ipn.IP
					if ip4 := ip.To4(); ip4 != nil {
						ip = ip4
					}
					ips = append(ips, ip.String())
				}
				if len(ips) == 0 {
					continue
				}
				mac := i.HardwareAddr.String()
				a := &inner.AgentReport_Addr{Mac: &mac, Ips: ips}
				report.Addrs = append(report.Addrs, a)
			}
		} else {
			log.Printf("Cannot get interfaces: %v", err)
		}
		// attrs
		for i, n := range ns {
			if n.Attr == 0 {
				break
			}
			if n.Value == 0 {
				continue
			}
			v := atomic.SwapUint64(&ns[i].Value, 0)
			a := &inner.AgentReport_Node{Attr: proto.Uint32(n.Attr), Value: proto.Uint64(v)}
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
		log.Printf("msg (size=%v) were sent to %v, report=%+v", len(msg), conn.RemoteAddr(), report)
	}
	// exit
	log.Fatalln("program exit!")
}

func systemStats() {
	// attrs
	const kCpuMax = 64
	const (
		_        = iota
		kLoadOne // sys load
		kLoadFive
		kLoadFifteen
		kCpu0 // cpu

		kMemTotal = iota + kCpuMax - 1 // memory
		kMemUsed
		kMemFree
		kSwapTotal // swap
		kSwapUsed
		kSwapFree
	)
	var cpus sigar.CpuList
	if err := cpus.Get(); err != nil {
		log.Printf("Cannot get cpu usage: %v", err)
	}
	time.Sleep(kSleep / 2)
	var s sigar.ConcreteSigar
	for {
		// sys load
		if load, err := s.GetLoadAverage(); err != nil {
			log.Printf("Cannot get system loads: %v", err)
		} else {
			attr.Set(kLoadOne, uint64(load.One*100))
			attr.Set(kLoadFive, uint64(load.Five*100))
			attr.Set(kLoadFifteen, uint64(load.Fifteen*100))
		}
		// cpu usage
		var cur sigar.CpuList
		if err := cur.Get(); err != nil {
			log.Printf("Cannot get cpu usage: %v", err)
		} else {
			for i, cpu := range cur.List {
				c := cpu.Delta(cpus.List[i])
				t, u := c.Total(), uint64(0)
				if t > 0 {
					u = (t - c.Idle) * 10000 / t
				}
				attr.Set(uint32(kCpu0+i), u)
			}
			cpus = cur
		}
		// memory usage
		if mem, err := s.GetMem(); err != nil {
			log.Printf("Cannot get memory info: %v", err)
		} else {
			attr.Set(kMemTotal, mem.Total)
			attr.Set(kMemUsed, mem.ActualUsed)
			attr.Set(kMemFree, mem.ActualFree)
		}
		// swap usage
		if swap, err := s.GetSwap(); err != nil {
			log.Printf("Cannot get swap info: %v", err)
		} else {
			attr.Set(kSwapTotal, swap.Total)
			attr.Set(kSwapUsed, swap.Used)
			attr.Set(kSwapFree, swap.Free)
		}

		time.Sleep(kSleep)
	}
}
