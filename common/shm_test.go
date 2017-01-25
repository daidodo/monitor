package common

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttachShm(t *testing.T) {
	cases := []struct {
		create bool
		mem    Mem
		memEq  bool
		err    error
		errEq  bool
	}{
		{false, nil, true, nil, false},
		{true, nil, false, nil, true},
		{false, nil, false, nil, true},
	}
	// setup
	if e := os.Remove(filePath); e != nil {
		if _, ok := e.(*os.PathError); !ok {

			assert.FailNow(t, fmt.Sprintf("cannot remove '%v': %v(%T)", filePath, e, e))
		}
	}
	// test
	for _, c := range cases {
		m, e := AttachShm(c.create)
		if c.memEq {
			assert.Equal(t, c.mem, m)
		} else {
			assert.NotEqual(t, c.mem, m)
		}
		if c.errEq {
			assert.Equal(t, c.err, e)
		} else {
			assert.NotEqual(t, c.err, e)
		}
	}
}
