package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/daidodo/overlord/inner"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/proto"
)

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
	log.Printf("process report=%v from %v", *report, addr)
}

func setupDB() {
	db, err := sql.Open("mysql", "root:mysql@(db:3306)/")
	if err != nil {
		log.Fatalf("Cannot connect to DB: %v", err)
	}
	// database 'overlord'
	if _, err = db.Exec("use overlord"); err != nil {
		if _, err := db.Exec("create database if not exists overlord"); err != nil {
			log.Fatalf("Cannot create database 'overlord': %v", err)
		}
		if _, err = db.Exec("use overlord"); err != nil {
			log.Fatalf("Cannot create database 'overlord': %v", err)
		}
		log.Print("database 'overlord' created")
	}
	// table 'machines'
	if _, err = db.Exec("select count(*) from machines"); err != nil {
		const sql = `create table if not exists machines(
			ip varchar(64) not null,
			mac varchar(64) not null,
			primary key(ip),
			key(mac)
		)`
		if _, err = db.Exec(sql); err != nil {
			log.Fatalf("Cannot create table 'machines': %v", err)
		}
		log.Print("table 'machines' created")
	}
	// table 'attrs'
	if _, err = db.Exec("select count(*) from attrs"); err != nil {
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
		if _, err = db.Exec(sql); err != nil {
			log.Fatalf("Cannot create table 'attrs': %v", err)
		}
		log.Print("table 'attrs' created")
	}

}
