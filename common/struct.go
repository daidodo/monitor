package common

import "sync/atomic"

type Node struct {
	Attr  uint32
	pad   uint32
	Value uint64
}

type Mem []Node

func (m Mem) FindNode(attr uint32) *Node {
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
