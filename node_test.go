package art

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeKind(t *testing.T) {
	n4 := factory.newNode4()
	assert.NotNil(t, n4)
	assert.Equal(t, NODE_4, n4.kind)

	n16 := factory.newNode16()
	assert.NotNil(t, n16)
	assert.Equal(t, NODE_16, n16.kind)

	n48 := factory.newNode48()
	assert.NotNil(t, n48)
	assert.Equal(t, NODE_48, n48.kind)

	n256 := factory.newNode256()
	assert.NotNil(t, n256)
	assert.Equal(t, NODE_256, n256.kind)
}

func TestLeaf(t *testing.T) {
	leaf := factory.newLeaf([]byte("key"), "value")
	assert.NotNil(t, leaf)
	assert.Equal(t, NODE_LEAF, leaf.kind)

	assert.False(t, leaf.Leaf().Match([]byte("unknown-key")))
}

func TestLeafMatch(t *testing.T) {
	leaf := factory.newLeaf([]byte("key"), "value")
	assert.False(t, leaf.Leaf().Match([]byte("unknown-key")))
	assert.False(t, leaf.Leaf().Match(nil))

	assert.True(t, leaf.Leaf().Match([]byte("key")))
}

func TestNodeSetPrefix(t *testing.T) {
	n4 := factory.newNode4()
	assert.NotNil(t, n4)
	node := n4.Node4()

	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n4.SetPrefix(key, 2)

	assert.Equal(t, 2, node.prefixLen)
	assert.Equal(t, byte(1), node.prefix[0])
	assert.Equal(t, byte(2), node.prefix[1])

	n4.SetPrefix(key, MAX_PREFIX_LENGTH)
	assert.Equal(t, MAX_PREFIX_LENGTH, node.prefixLen)
	assert.Equal(t, byte(1), node.prefix[0])
	assert.Equal(t, byte(2), node.prefix[1])
	assert.Equal(t, byte(3), node.prefix[2])
	assert.Equal(t, byte(4), node.prefix[3])

}

func TestNodeCheckPrefix(t *testing.T) {
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n16 := factory.newNode16()

	node := n16.Node16()
	n16.SetPrefix([]byte{1, 2, 3, 4, 5, 66, 77, 88, 99}, 5)

	assert.Equal(t, 5, node.CheckPrefix(key, 0))
	assert.Equal(t, 0, node.CheckPrefix(key, 1))
	assert.Equal(t, 0, node.CheckPrefix(key, 100))
}

func TestNodeAddChild(t *testing.T) {
	nodes := []*artNode{factory.newNode4(), factory.newNode16(), factory.newNode48(), factory.newNode256()}

	for _, n := range nodes {
		l := n.maxChildren()
		for i := 0; i < l; i++ {
			leaf := factory.newLeaf([]byte{byte(i)}, i)
			n.AddChild(byte(i), leaf)
		}

		for i := 0; i < l; i++ {
			leaf := n.FindChild(byte(i))
			assert.NotNil(t, *leaf)
			assert.Equal(t, i, (*leaf).Leaf().value.(int))
		}
	}
}

func TestNodeIndex(t *testing.T) {
	nodes := []*artNode{factory.newNode4(), factory.newNode16(), factory.newNode48(), factory.newNode256()}

	for _, n := range nodes {
		l := n.maxChildren()
		for i := 0; i < l; i++ {
			leaf := factory.newLeaf([]byte{byte(i)}, i)
			n.AddChild(byte(i), leaf)
		}

		for i := 0; i < l; i++ {
			assert.Equal(t, i, n.Index(byte(i)))
		}
	}
}

func TestNode4AddChildAndFindChild(t *testing.T) {
	parent := factory.newNode4()
	child := factory.newNode4()
	parent.AddChild(1, child)

	assert.Equal(t, 1, parent.Node4().numChildren)
	assert.Equal(t, child, *parent.FindChild(1))
}

func TestNode4AddChildTwicePreserveSorted(t *testing.T) {
	parent := factory.newNode4()
	child1 := factory.newNode4()
	child2 := factory.newNode4()
	parent.AddChild(byte(2), child1)
	parent.AddChild(byte(1), child2)

	assert.Equal(t, 2, parent.Node4().numChildren)
	assert.Equal(t, byte(1), parent.Node4().keys[0])
	assert.Equal(t, byte(2), parent.Node4().keys[1])
}

func TestNode4AddChild4PreserveSorted(t *testing.T) {
	parent := factory.newNode4()
	for i := 4; i > 0; i-- {
		parent.AddChild(byte(i), factory.newNode4())
	}

	assert.Equal(t, 4, parent.Node4().numChildren)
	assert.Equal(t, []byte{1, 2, 3, 4}, parent.Node4().keys[:])
}

func TestNode16AddChild16PreserveSorted(t *testing.T) {
	parent := factory.newNode16()
	for i := 16; i > 0; i-- {
		parent.AddChild(byte(i), factory.newNode16())
	}

	assert.Equal(t, 16, parent.Node4().numChildren)
	for i := 0; i < 16; i++ {
		assert.Equal(t, byte(i+1), parent.Node16().keys[i])
	}
}

// Art Nodes of all types should grow to the next biggest size in sequence
func TestGrow(t *testing.T) {
	nodes := []*artNode{factory.newNode4(), factory.newNode16(), factory.newNode48()}
	expected := []Kind{NODE_16, NODE_48, NODE_256}

	for i, node := range nodes {
		newNode := node.grow()
		assert.Equal(t, expected[i], newNode.kind)
	}
}

// Art Nodes of all types should next smallest size in sequence
func TestShrink(t *testing.T) {
	nodes := []*artNode{factory.newNode256(), factory.newNode48(), factory.newNode16(), factory.newNode4()}
	expected := []Kind{NODE_48, NODE_16, NODE_4, NODE_LEAF}

	for i, node := range nodes {

		for j := 0; j < node.minChildren(); j++ {
			if node.kind != NODE_4 {
				node.AddChild(byte(i), factory.newNode4())
			} else {
				node.AddChild(byte(i), factory.newLeaf([]byte{byte(i)}, "value"))
			}
		}

		newNode := node.shrink()
		assert.Equal(t, expected[i], newNode.kind)
	}
}
