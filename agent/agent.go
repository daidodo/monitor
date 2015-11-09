package main

import (
	"log"
	"os"
	"syscall"
	"time"
	"unsafe"
)

type node struct {
	attr  uint64
	value uint64
}

const (
	filePath = "/tmp/attr.data"
	maxLen   = int64(10 << 10)
)

var out *log.Logger
var gm []node

func main() {
	// init log
	log.SetPrefix("[monitor_agent]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	// init shm
	gm = initShm()
	for i := 0; i < len(gm); i++ {
		n := gm[i]
		if n.attr == 0 {
			break
		}
		n.value = 0
	}
	// loop
	for {
		time.Sleep(60 * time.Second)
		for i := 0; i < len(gm); i++ {
			n := gm[i]
			if n.attr == 0 {
				break
			}
			// TODO: gather all [attr, value] pairs and send to server
		}
	}
	// exit
	log.Println("program exit!")
}

func initShm() []node {
	// prepare
	const (
		openFlags = os.O_RDWR
		mmapProts = syscall.PROT_READ | syscall.PROT_WRITE
	)
	// open file
	file, err := os.OpenFile(filePath, openFlags, 0)
	if err != nil {
		// create file
		file, err = os.Create(filePath)
		if err != nil {
			log.Panicf("cannot create file '%v': %v", filePath, err)
		}
		// init file
		err = file.Truncate(maxLen * int64(unsafe.Sizeof(node{})))
		if err != nil {
			file.Close()
			log.Panicf("cannot initialize file '%v': %v", filePath, err)
		}
	}
	// get file size
	ls, err := file.Stat()
	if err != nil {
		file.Close()
		log.Panicf("cannot lstat file '%v': %v", filePath, err)
	}
	size := ls.Size()
	// calc node count
	len := size / int64(unsafe.Sizeof(node{}))
	if len > maxLen {
		len = maxLen
	}
	// memory map
	ptr, err := syscall.Mmap(int(file.Fd()), 0, int(size), mmapProts, syscall.MAP_SHARED)
	if err != nil {
		log.Panicf("cannot mmap file '%v'(size=%v): %v", filePath, size, err)
	}
	return (*[maxLen]node)(unsafe.Pointer(&ptr[0]))[:len]
}
