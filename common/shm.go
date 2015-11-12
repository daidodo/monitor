package common

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func AttachShm(create bool) Mem {
	const (
		filePath  = "/tmp/attr.data"
		maxLen    = int64(10 << 10)
		openFlags = os.O_RDWR
		mmapProts = syscall.PROT_READ | syscall.PROT_WRITE
	)
	// open file
	file, err := os.OpenFile(filePath, openFlags, 0)
	if err != nil {
		if !create {
			panic(fmt.Sprintf("open file '%v' failed, maybe create it first: %v", filePath, err))
		}
		// create file
		file, err = os.Create(filePath)
		if err != nil {
			panic(fmt.Sprintf("cannot create file '%v': %v", filePath, err))
		}
		// init file
		err = file.Truncate(maxLen * int64(unsafe.Sizeof(Node{})))
		if err != nil {
			file.Close()
			panic(fmt.Sprintf("cannot initialize file '%v': %v", filePath, err))
		}
	}
	// get file size
	ls, err := file.Stat()
	if err != nil {
		file.Close()
		panic(fmt.Sprintf("cannot lstat file '%v': %v", filePath, err))
	}
	size := ls.Size()
	// calc node count
	len := size / int64(unsafe.Sizeof(Node{}))
	if len > maxLen {
		len = maxLen
	}
	// memory map
	ptr, err := syscall.Mmap(int(file.Fd()), 0, int(size), mmapProts, syscall.MAP_SHARED)
	if err != nil {
		panic(fmt.Sprintf("cannot mmap file '%v'(size=%v): %v", filePath, size, err))
	}
	return (*[maxLen]Node)(unsafe.Pointer(&ptr[0]))[:len]
}
