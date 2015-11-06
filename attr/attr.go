package attr

import (
	"fmt"
	"os"
	"sync/atomic"
	"syscall"
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

var gNodes []node

func init() {
	gNodes = initCreateWrite(true)
}

// Add increments value for 'attr' by 'delta'
func Add(attr, delta uint64) {
	if delta > 0 {
		n := findNode(attr)
		if n != nil && attr == n.attr {
			atomic.AddUint64(&n.value, delta)
		}
	}
}

// Set sets value for 'attr' to 'value'
func Set(attr, value uint64) {
	n := findNode(attr)
	if n != nil && attr == n.attr {
		atomic.StoreUint64(&n.value, value)
	}
}

func initCreateWrite(write bool) []node {
	// prepare
	openFlags := os.O_RDONLY
	mmapProts := syscall.PROT_READ
	if write {
		openFlags = os.O_RDWR
		mmapProts |= syscall.PROT_WRITE
	}
	// open file
	file, err := os.OpenFile(filePath, openFlags, 0)
	if err != nil {
		// create file
		file, err = os.Create(filePath)
		if err != nil {
			panic(fmt.Sprintf("cannot create file '%v' for package 'attr': %v", filePath, err))
		}
		// init file
		err = file.Truncate(maxLen * int64(unsafe.Sizeof(node{})))
		if err != nil {
			file.Close()
			panic(fmt.Sprintf("cannot initialize file '%v' for package 'attr': %v", filePath, err))
		}
	}
	// get file size
	ls, err := file.Stat()
	if err != nil {
		file.Close()
		panic(fmt.Sprintf("cannot lstat file '%v' for package 'attr': %v", filePath, err))
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
		panic(fmt.Sprintf("cannot mmap file '%v'(size=%v) for package 'attr': %v", filePath, size, err))
	}
	return (*[maxLen]node)(unsafe.Pointer(&ptr[0]))[:len]
}

func findNode(attr uint64) *node {
	if attr > 0 {
		for i := 0; i < len(gNodes); i++ {
			n := &gNodes[i]
			if n.attr == 0 {
				if atomic.CompareAndSwapUint64(&n.attr, 0, attr) {
					return n
				}
			} else if attr == n.attr {
				return n
			}
		}
	}
	return nil
}
