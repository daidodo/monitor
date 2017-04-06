package inner

import (
	"os"
	"sync/atomic"
	"syscall"
	"unsafe"
)

const (
	filePath = "/tmp/attr.data"
	maxLen   = int64(10 << 10)
)

type Node struct {
	Attr  uint32
	pad   uint32
	Value uint64
}

type Nodes []Node

func (m Nodes) FindNode(attr uint32) *Node {
	if attr > 0 {
		for i := 0; i < len(m); i++ {
			n := &m[i]
			if n.Attr == 0 {
				if atomic.CompareAndSwapUint32(&n.Attr, 0, attr) {
					return n
				}
			} else if attr == n.Attr {
				return n
			}
		}
	}
	return nil
}

func Attach(create bool) (mem Nodes, err error) {
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
