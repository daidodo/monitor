package common

import (
	"os"
	"testing"

	"github.com/daidodo/testa/assert"
)

func TestAttach(t *testing.T) {
	cases := []struct {
		create bool
		mem    Nodes
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
		assert.EqualType(t, e, (*os.PathError)(nil), "cannot remove '%v': %v(%T)", filePath, e, e)
	}
	// test
	for _, c := range cases {
		m, e := Attach(c.create)
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
