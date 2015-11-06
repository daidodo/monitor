package attr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getValue(attr uint64) uint64 {
	if n := findNode(attr); n != nil {
		return n.value
	}
	return 0
}

func TestAdd(t *testing.T) {
	assert.Zero(t, getValue(0))
	Add(0, 1)
	assert.Zero(t, getValue(0))

	const attr = 123456
	val := getValue(attr)
	for i := 0; i < 10; i++ {
		Add(attr, uint64(i))
		val += uint64(i)
		assert.Equal(t, val, getValue(attr))
	}
}

func TestSet(t *testing.T) {
	assert.Zero(t, getValue(0))
	Set(0, 1)
	assert.Zero(t, getValue(0))

	const attr = 23456
	for i := 0; i < 10; i++ {
		Set(attr, uint64(i))
		assert.Equal(t, uint64(i), getValue(attr))
	}
}
