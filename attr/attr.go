package attr

import (
	"fmt"
	"sync/atomic"

	"github.com/daidodo/overlord/inner"
)

var ns inner.Nodes

func init() {
	var err error
	if ns, err = inner.Attach(false); err != nil {
		panic(fmt.Errorf("Cannot init overlord/attr: %v", err))
	}
}

// Add increments value for 'attr' by 'delta'
func Add(attr uint32, delta uint64) {
	if delta > 0 {
		n := ns.FindNode(attr)
		if n != nil && attr == n.Attr {
			atomic.AddUint64(&n.Value, delta)
		}
	}
}

// Set sets value for 'attr' to 'value'
func Set(attr uint32, value uint64) {
	n := ns.FindNode(attr)
	if n != nil && attr == n.Attr {
		atomic.StoreUint64(&n.Value, value)
	}
}
