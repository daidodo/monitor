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
	setup()
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

func setup() {
	db, err := sql.Open("mysql", "root:mysql@(db:3306)/")
	if err != nil {
		log.Fatalf("Cannot connect to DB: %v", err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Cannot begin a transaction: %v", err)
	}
	// database 'overlord'
	if _, err = tx.Exec("create database if not exists overlord"); err != nil {
		log.Fatalf("Cannot create database 'overlord': %v", err)
	}
	if _, err = tx.Exec("use overlord"); err != nil {
		log.Fatalf("Cannot create database 'overlord': %v", err)
	}
	// table 'machines'
	const sql = `create table if not exists machines(
		ip varchar(64) not null,
		mac varchar(64) not null,
		primary key(ip),
		key(mac)
	)`
	if _, err = tx.Exec(sql); err != nil {
		log.Fatalf("Cannot create table 'machines': %v", err)
	}

	if err = tx.Commit(); err != nil {
		log.Fatalf("Cannot commit the transaction: %v", err)
	}
}
