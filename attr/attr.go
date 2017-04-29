package attr

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/daidodo/overlord/inner"
)

// Add increments value for 'attr' by 'delta'
func Add(attr uint32, delta uint64) {
	if delta > 0 {
		n := node(attr)
		if n != nil && attr == n.Attr {
			atomic.AddUint64(&n.Value, delta)
		}
	}
}

// Set sets value for 'attr' to 'value'
func Set(attr uint32, value uint64) {
	n := node(attr)
	if n != nil && attr == n.Attr {
		atomic.StoreUint64(&n.Value, value)
	}
}

var ns inner.Nodes

func node(attr uint32) *inner.Node {
	if ns != nil {
		return ns.FindNode(attr)
	}
	return nil
}

func init() {
	if ns, _ = inner.Attach(false); ns == nil {
		go func() {
			for ns == nil {
				time.Sleep(time.Second)
				var err error
				if ns, err = inner.Attach(false); err != nil {
					log.Printf("Cannot init overlord/attr: %v", err)
				}
			}
		}()
	}
}
