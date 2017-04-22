package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/daidodo/overlord/inner"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/proto"
)

var gDb *sql.DB

func main() {
	// init log
	log.SetPrefix("[overlord_collector]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	// setup NOTE: should move to other place.
	setupDB()
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
	log.Printf("process report=%+v from %v", *report, addr)
	if gDb == nil {
		log.Fatalf("DB has NOT been inited")
	}
	// update machines
	var macs []string
	for _, addr := range report.Addrs {
		mac := addr.GetMac()
		if len(mac) < 1 {
			continue
		}
		macs = append(macs, mac)
		for _, ip := range addr.GetIps() {
			if len(ip) < 1 {
				continue
			}
			sql := fmt.Sprintf("insert into machines(ip, mac) values ('%v', '%v') on duplicate key update ip='%[1]v'", ip, mac)
			res, err := gDb.Exec(sql)
			if err != nil {
				log.Printf("Cannot update table 'machines': %v", err)
			}
			if r, _ := res.RowsAffected(); r > 0 {
				log.Printf("find new machine, mac=%v, ip=%v", mac, ip)
			}
		}
	}
	// update attrs
	if len(macs) < 1 {
		log.Printf("No mac address in report=%+v from %v, ignore it", report, addr)
		return
	}
	attrs := report.GetAttrs()
	if len(attrs) < 1 {
		log.Printf("No attrs in report=%+v from %v, ignore it", report, addr)
		return
	}
	sql := bytes.NewBufferString("insert into attrs(mac,attr,value) values ")
	for _, attr := range report.GetAttrs() {
		a, v := attr.GetAttr(), attr.GetValue()
		for _, m := range macs {
			fmt.Fprintf(sql, "('%v',%v,%v),", m, a, v)
		}
	}
	sql.Truncate(sql.Len() - 1) // truncate last ','
	if _, err := gDb.Exec(sql.String()); err != nil {
		log.Printf("Cannot update table 'attrs': %v", err)
	}
}

func setupDB() {
	var err error
	gDb, err = sql.Open("mysql", "root:mysql@(db:3306)/")
	if err != nil {
		log.Fatalf("Cannot connect to DB: %v", err)
	}
	// database 'overlord'
	if _, err = gDb.Exec("use overlord"); err != nil {
		if _, err := gDb.Exec("create database if not exists overlord"); err != nil {
			log.Fatalf("Cannot create database 'overlord': %v", err)
		}
		if _, err = gDb.Exec("use overlord"); err != nil {
			log.Fatalf("Cannot create database 'overlord': %v", err)
		}
		log.Print("database 'overlord' created")
	}
	// table 'machines'
	if _, err = gDb.Exec("select count(*) from machines"); err != nil {
		const sql = `create table if not exists machines(
			ip varchar(64) not null,
			mac varchar(64) not null,
			primary key(ip),
			key(mac)
		)`
		if _, err = gDb.Exec(sql); err != nil {
			log.Fatalf("Cannot create table 'machines': %v", err)
		}
		log.Print("table 'machines' created")
	}
	// table 'attrs'
	if _, err = gDb.Exec("select count(*) from attrs"); err != nil {
		const sql = `create table if not exists attrs(
			mac varchar(64) not null,
			time timestamp default current_timestamp,
			attr int unsigned not null,
			value bigint unsigned not null,
			primary key(mac, time, attr),
			key(mac),
			key(time),
			key(attr)
		)`
		if _, err = gDb.Exec(sql); err != nil {
			log.Fatalf("Cannot create table 'attrs': %v", err)
		}
		log.Print("table 'attrs' created")
	}
}
