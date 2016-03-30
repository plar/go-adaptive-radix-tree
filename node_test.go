package art

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeMinMaxSize(t *testing.T) {
	n4 := newNode4()
	assert.NotNil(t, n4)
	assert.Equal(t, NODE_4, int(n4.kind))

	n16 := newNode16()
	assert.NotNil(t, n16)
	assert.Equal(t, NODE_16, int(n16.kind))

	n48 := newNode48()
	assert.NotNil(t, n48)
	assert.Equal(t, NODE_48, int(n48.kind))

	n256 := newNode256()
	assert.NotNil(t, n256)
	assert.Equal(t, NODE_256, int(n256.kind))
}

func TestLeaf(t *testing.T) {
	leaf := newLeaf([]byte("key"), "value")
	assert.NotNil(t, leaf)
	assert.Equal(t, NODE_LEAF, int(leaf.kind))

	assert.False(t, leaf.Leaf().Match([]byte("unknown-key"), 0))
}

func TestLeafMatch(t *testing.T) {
	leaf := newLeaf([]byte("key"), "value")
	assert.False(t, leaf.Leaf().Match([]byte("unknown-key"), 0))
	assert.False(t, leaf.Leaf().Match(nil, 0))

	assert.True(t, leaf.Leaf().Match([]byte("key"), 0))
}
