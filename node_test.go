package art

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test basic properties and behavior of each node kind.
func TestNodeKindProperties(t *testing.T) {
	t.Parallel()

	// Define a Table of Node Types to Test
	nodeTests := []struct {
		name string
		node *nodeRef
		kind Kind
	}{
		{"Node4 Test", factory.newNode4(), Node4},
		{"Node16 Test", factory.newNode16(), Node16},
		{"Node48 Test", factory.newNode48(), Node48},
		{"Node256 Test", factory.newNode256(), Node256},
	}

	// Run Node Kind Tests
	for _, tt := range nodeTests {
		tt := tt // Pin
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.NotNil(t, tt.node)
			assert.Equal(t, tt.kind, tt.node.kind)
			assert.Equal(t, tt.kind.String(), tt.kind.String())
		})
	}

	// Test Leaf Node
	t.Run("Leaf Node Test", func(t *testing.T) {
		leaf := factory.newLeaf(Key("key"), "value")
		assert.NotNil(t, leaf)
		assert.Equal(t, Leaf, leaf.kind)
		assert.Equal(t, Key("key"), leaf.Key())
		assert.Equal(t, "value", leaf.Value().(string))
		assert.Equal(t, "Leaf", leaf.kind.String())
	})
}

func TestUnknownNode(t *testing.T) {
	t.Parallel()

	unknownNode := &nodeRef{kind: Kind(0xFF)}
	assert.Nil(t, unknownNode.maximum())
	assert.Nil(t, unknownNode.minimum())
}

func TestLeafFunctionality(t *testing.T) {
	t.Parallel()

	leaf := factory.newLeaf([]byte("key"), "value")
	assert.NotNil(t, leaf)
	assert.Equal(t, Leaf, leaf.kind)

	assert.False(t, leaf.leaf().match(Key("unknown-key")))

	// Ensure we cannot shrink/grow leaf node
	assert.Nil(t, toNode(leaf).shrink())
	assert.Nil(t, toNode(leaf).grow())
}

// Test matching behavior of leaf nodes.
func TestLeafMatchBehavior(t *testing.T) {
	t.Parallel()

	leaf := factory.newLeaf(Key("key"), "value")

	assert.False(t, leaf.leaf().match(Key("unknown-key")))
	assert.False(t, leaf.leaf().match(nil))
	assert.True(t, leaf.leaf().match(Key("key")))

	assert.False(t, leaf.leaf().prefixMatch(Key("unknown-key")))
	assert.False(t, leaf.leaf().prefixMatch(nil))
	assert.True(t, leaf.leaf().prefixMatch(Key("ke")))
}

// Check the setting of prefixes.
func TestNodePrefixSetting(t *testing.T) {
	t.Parallel()

	n4 := factory.newNode4()
	nn := n4.node()

	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	n4.setPrefix(key, 2)
	assert.Equal(t, 2, int(nn.prefixLen))
	assert.Equal(t, byte(1), nn.prefix[0])
	assert.Equal(t, byte(2), nn.prefix[1])

	n4.setPrefix(key, MaxPrefixLen)
	assert.Equal(t, MaxPrefixLen, int(nn.prefixLen))
	assert.Equal(t, []byte{1, 2, 3, 4}, nn.prefix[:4])
}

// Test the matching of nodes with keys.
func TestNodeMatchKeyBehavior(t *testing.T) {
	t.Parallel()

	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n16 := factory.newNode16()
	n16.setPrefix([]byte{1, 2, 3, 4, 5, 66, 77, 88, 99}, 5)

	assert.Equal(t, 5, n16.match(key, 0))
	assert.Equal(t, 0, n16.match(key, 1))
	assert.Equal(t, 0, n16.match(key, 100))
}

func TestCopyNode(t *testing.T) {
	t.Parallel()

	// Define test data
	src := &node{
		childrenLen: 3,
		prefixLen:   5,
		prefix:      [MaxPrefixLen]byte{'a', 'b', 'c', 'd', 'e'},
	}

	dst := &node{
		childrenLen: 0,
		prefixLen:   0,
		prefix:      [MaxPrefixLen]byte{},
	}

	// Call the function being tested
	copyNode(dst, src)

	// Use assertions to verify the outcomes
	assert.Equal(t, dst.childrenLen, uint16(0), "childrenLen should not be copied")
	assert.Equal(t, src.prefixLen, dst.prefixLen, "prefixLen should be copied correctly")

	maxCopyLen := min(int(src.prefixLen), MaxPrefixLen)
	for i := 0; i < maxCopyLen; i++ {
		assert.Equal(t, src.prefix[i], dst.prefix[i], "prefix[%d] should be copied correctly", i)
	}
}

// Test adding children to nodes and retrieving them.
func TestNodeAddChildAndFindChild(t *testing.T) {
	nodeKinds := []struct {
		name        string
		node        *nodeRef
		maxChildren int
	}{
		{"Node4", factory.newNode4(), node4Max},
		{"Node16", factory.newNode16(), node16Max},
		{"Node48", factory.newNode48(), node48Max},
		{"Node256", factory.newNode256(), node256Max},
	}

	for _, n := range nodeKinds {
		n := n // Pin for parallel execution
		t.Run(n.name, func(t *testing.T) {
			t.Parallel()

			for i := 0; i < n.maxChildren; i++ {
				leaf := factory.newLeaf(Key{byte(i)}, i)
				n.node.addChild(keyChar{ch: byte(i)}, leaf)
			}

			for i := 0; i < n.maxChildren; i++ {
				leaf := n.node.findChildByKey(Key{byte(i)}, 0)
				assert.NotNil(t, *leaf, "child should not be nil for key %d", i)
				val := (*leaf).leaf().value.(int)
				assert.Equal(t, i, val, "value should be %d", i)
			}
		})
	}
}

// Test indexing functionality across different nodes.
func TestNodeIndex(t *testing.T) {
	t.Parallel()

	nodes := []*nodeRef{
		factory.newNode4(),
		factory.newNode16(),
		factory.newNode48(),
		factory.newNode256(),
	}

	for _, n := range nodes {
		var maxChildren int
		switch n.kind {
		case Node4:
			maxChildren = node4Max
		case Node16:
			maxChildren = node16Max
		case Node48:
			maxChildren = node48Max
		case Node256:
			maxChildren = node256Max
		}

		for i := 0; i < maxChildren; i++ {
			leaf := factory.newLeaf(Key{byte(i)}, i)
			n.addChild(keyChar{ch: byte(i)}, leaf)
		}

		for i := 0; i < maxChildren; i++ {
			assert.Equal(t, i, toNode(n).index(keyChar{ch: byte(i)}))
		}
	}
}

// Test minimum and maximum functionality to ensure they return correct leaf nodes.
func TestNodesMinimumMaximum(t *testing.T) {
	t.Parallel()

	nodes := []struct {
		node  *nodeRef
		count int
	}{
		{factory.newNode4(), 3},
		{factory.newNode16(), 15},
		{factory.newNode48(), 47},
		{factory.newNode256(), 255},
	}

	for _, n := range nodes {
		n := n // Pin for parallel execution
		t.Run(n.node.kind.String(), func(t *testing.T) {
			t.Parallel()

			for j := 1; j <= n.count; j++ {
				n.node.addChild(keyChar{ch: byte(j)}, factory.newLeaf([]byte{byte(j)}, byte(j)))
			}

			minLeaf := n.node.minimum()
			assert.Equal(t, minLeaf.key, Key{1})
			assert.Equal(t, minLeaf.value.(byte), minLeaf.value.(byte))

			maxLeaf := n.node.maximum()
			assert.Equal(t, maxLeaf.key, Key{byte(n.count)})
			assert.Equal(t, maxLeaf.value.(byte), maxLeaf.value.(byte))
		})
	}
}

// Test adding and finding children in a Node4.
func TestNode4AddChildAndFindChild(t *testing.T) {
	t.Parallel()

	parent := factory.newNode4()
	child := factory.newNode4()
	k := Key{1}
	parent.addChild(keyChar{ch: k[0]}, child)

	assert.Equal(t, 1, int(parent.node().childrenLen))
	assert.Equal(t, child, *parent.findChildByKey(k, 0))
}

// Test that Node4 maintains sorted order when adding children.
func TestNode4AddChildTwicePreserveSorted(t *testing.T) {
	t.Parallel()

	parent := factory.newNode4()
	child1 := factory.newNode4()
	child2 := factory.newNode4()
	parent.addChild(keyChar{ch: 2}, child1)
	parent.addChild(keyChar{ch: 1}, child2)

	assert.Equal(t, 2, int(parent.node().childrenLen))
	assert.Equal(t, byte(1), parent.node4().keys[0])
	assert.Equal(t, byte(2), parent.node4().keys[1])
}

// Test Node4 maintains sorted order with multiple children.
func TestNode4AddChild4PreserveSorted(t *testing.T) {
	t.Parallel()

	parent := factory.newNode4()
	for i := 4; i > 0; i-- {
		parent.addChild(keyChar{ch: byte(i)}, factory.newNode4())
	}

	assert.Equal(t, 4, int(parent.node().childrenLen))
	assert.Equal(t, []byte{1, 2, 3, 4}, parent.node4().keys[:])
}

// Test Node16 maintains sorted order with multiple children.
func TestNode16AddChild16PreserveSorted(t *testing.T) {
	parent := factory.newNode16()
	for i := 16; i > 0; i-- {
		parent.addChild(keyChar{ch: byte(i)}, factory.newNode16())
	}

	assert.Equal(t, 16, int(parent.node().childrenLen))
	for i := 0; i < 16; i++ {
		assert.Equal(t, byte(i+1), parent.node16().keys[i])
	}
}

// Test growing a node.
func TestNodeGrow(t *testing.T) {
	nodeKinds := []struct {
		name     string
		node     *nodeRef
		expected Kind
	}{
		{"Node4", factory.newNode4(), Node16},
		{"Node16", factory.newNode16(), Node48},
		{"Node48", factory.newNode48(), Node256},
	}

	for _, tt := range nodeKinds {
		tt := tt // Pin for parallel execution
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			newNode := toNode(tt.node).grow()
			assert.Equal(t, tt.expected, newNode.kind)
		})
	}
}

// Test shrinking a node.
func TestNodeShrink(t *testing.T) {
	nodeKinds := []struct {
		name        string
		node        *nodeRef
		expected    Kind
		minChildren int
	}{
		{"Node256", factory.newNode256(), Node48, node256Min},
		{"Node48", factory.newNode48(), Node16, node48Min},
		{"Node16", factory.newNode16(), Node4, node16Min},
		{"Node4", factory.newNode4(), Leaf, node4Min},
	}

	for _, tt := range nodeKinds {
		tt := tt // Pin for parallel execution
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for j := 0; j < tt.minChildren; j++ {
				if tt.node.kind != Node4 {
					tt.node.addChild(keyChar{ch: byte(j)}, factory.newNode4())
				} else {
					tt.node.addChild(keyChar{ch: byte(j)}, factory.newLeaf(Key{byte(j)}, "value"))
				}
			}

			newNode := toNode(tt.node).shrink()
			assert.Equal(t, tt.expected, newNode.kind)
		})
	}
}
