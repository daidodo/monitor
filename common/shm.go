package common

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	filePath = "/tmp/attr.data"
	maxLen   = int64(10 << 10)
)

func AttachShm(create bool) (mem Mem, err error) {
	const (
		openFlags = os.O_RDWR
		mmapProts = syscall.PROT_READ | syscall.PROT_WRITE
	)
	// open file
	file, err := os.OpenFile(filePath, openFlags, 0)
	if err != nil {
		if !create {
			return
		}
		// create file
		file, err = os.Create(filePath)
		if err != nil {
			return
		}
		defer file.Close()
		// init file
		err = file.Truncate(maxLen * int64(unsafe.Sizeof(Node{})))
		if err != nil {
			return
		}
	} else {
		defer file.Close()
	}
	// calc node count
	ls, err := file.Stat()
	if err != nil {
		return
	}
	size := ls.Size()
	len := size / int64(unsafe.Sizeof(Node{}))
	if len > maxLen {
		len = maxLen
	}
	// memory map
	ptr, err := syscall.Mmap(int(file.Fd()), 0, int(size), mmapProts, syscall.MAP_SHARED)
	if err != nil {
		return
	}
	return (*[maxLen]Node)(unsafe.Pointer(&ptr[0]))[:len], nil
}
