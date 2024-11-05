package art

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test basic properties and behavior of each node kind.
func TestNodeKind(t *testing.T) {
	nodes := []struct {
		node *nodeRef
		kind Kind
	}{
		{factory.newNode4(), Node4},
		{factory.newNode16(), Node16},
		{factory.newNode48(), Node48},
		{factory.newNode256(), Node256},
	}

	for _, n := range nodes {
		assert.NotNil(t, n.node)
		assert.Equal(t, n.kind, n.node.kind)
		assert.Equal(t, n.kind.String(), n.kind.String())
	}

	leaf := factory.newLeaf(Key("key"), "value")
	assert.NotNil(t, leaf)
	assert.Equal(t, Leaf, leaf.kind)
	assert.Equal(t, Key(Key("key")), leaf.Key())
	assert.Equal(t, "value", leaf.Value().(string))
	assert.Equal(t, "Leaf", leaf.kind.String())

	unknowNode := &nodeRef{kind: Kind(0xFF)}
	assert.Nil(t, unknowNode.maximum())
	assert.Nil(t, unknowNode.minimum())
}

func TestLeaf(t *testing.T) {
	leaf := factory.newLeaf([]byte("key"), "value")
	assert.NotNil(t, leaf)
	assert.Equal(t, Leaf, leaf.kind)

	assert.False(t, leaf.leaf().match(Key("unknown-key")))

	// we cannot shrink/grow leaf node
	assert.Nil(t, toNode(leaf).shrink())
	assert.Nil(t, toNode(leaf).grow())
}

// Test matching behavior of leaf nodes.
func TestLeafMatch(t *testing.T) {
	leaf := factory.newLeaf(Key("key"), "value")

	// Matching against keys
	assert.False(t, leaf.leaf().match(Key("unknown-key")))
	assert.False(t, leaf.leaf().match(nil))
	assert.True(t, leaf.leaf().match(Key("key")))

	// Prefix matching
	assert.False(t, leaf.leaf().prefixMatch(Key("unknown-key")))
	assert.False(t, leaf.leaf().prefixMatch(nil))
	assert.True(t, leaf.leaf().prefixMatch(Key("ke")))
}

// Check the setting of prefixes.
func TestNodeSetPrefix(t *testing.T) {
	n4 := factory.newNode4()
	nn := n4.node()

	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n4.setPrefix(key, 2)

	assert.Equal(t, 2, int(nn.prefixLen))
	assert.Equal(t, byte(1), nn.prefix[0])
	assert.Equal(t, byte(2), nn.prefix[1])

	n4.setPrefix(key, MaxPrefixLen)
	assert.Equal(t, MaxPrefixLen, int(nn.prefixLen))
	assert.Equal(t, byte(1), nn.prefix[0])
	assert.Equal(t, byte(2), nn.prefix[1])
	assert.Equal(t, byte(3), nn.prefix[2])
	assert.Equal(t, byte(4), nn.prefix[3])
}

// Test the matching of nodes with keys.
func TestNodeMatchWithKey(t *testing.T) {
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n16 := factory.newNode16()
	n16.setPrefix([]byte{1, 2, 3, 4, 5, 66, 77, 88, 99}, 5)

	assert.Equal(t, 5, n16.match(key, 0))
	assert.Equal(t, 0, n16.match(key, 1))
	assert.Equal(t, 0, n16.match(key, 100))
}

func TestCopyNode(t *testing.T) {

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
	// assert.Equal(t, src.zeroChild, dst.zeroChild, "zeroChild should be copied correctly")

	maxCopyLen := min(int(src.prefixLen), MaxPrefixLen)
	for i := 0; i < maxCopyLen; i++ {
		assert.Equal(t, src.prefix[i], dst.prefix[i], "prefix[%d] should be copied correctly", i)
	}
}

// Test adding children to nodes and retrieving them.
func TestNodeAddChild(t *testing.T) {
	nodes := []*nodeRef{
		factory.newNode4(),
		// factory.newNode16(),
		// factory.newNode48(),
		// factory.newNode256(),
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

		fmt.Println(DumpNode(n))

		for i := 0; i < maxChildren; i++ {
			k := Key{byte(i)}
			leaf := n.findChildByKey(k, 0)
			assert.NotNil(t, *leaf, "child should not be nil for key %d", i)
			val := (*leaf).leaf().value.(int)
			assert.Equal(t, i, val, "value should be %d", i)
		}
	}
}

// // Ensure that adding children to a leaf node doesn't alter the structure.
// func TestNodeAddChildForLeaf(t *testing.T) {
// 	leaf := factory.newLeaf(Key("key"), "value")
// 	assert.False(t, leaf.addChild('c', true, nil))
// }

// Test indexing functionality across different nodes.
func TestNodeIndex(t *testing.T) {
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
	// TODO: Merge nodes and inserts
	nodes := []*nodeRef{
		factory.newNode4(),
		factory.newNode16(),
		factory.newNode48(),
		factory.newNode256(),
	}

	inserts := []int{3, 15, 47, 255}

	for i, node := range nodes {
		for j := 1; j <= inserts[i]; j++ {
			node.addChild(keyChar{ch: byte(j)}, factory.newLeaf([]byte{byte(j)}, byte(j)))
		}

		minLeaf := node.minimum()
		assert.Equal(t, minLeaf.key, Key{1})
		assert.Equal(t, minLeaf.value.(byte), minLeaf.value.(byte))

		maxLeaf := node.maximum()
		assert.Equal(t, maxLeaf.key, Key{byte(inserts[i])})
		assert.Equal(t, maxLeaf.value.(byte), maxLeaf.value.(byte))
	}
}

// Test adding and finding children in a Node4.
func TestNode4AddChildAndFindChild(t *testing.T) {
	parent := factory.newNode4()
	child := factory.newNode4()
	k := Key{1}
	parent.addChild(keyChar{ch: k[0]}, child)

	assert.Equal(t, 1, int(parent.node().childrenLen))
	assert.Equal(t, child, *parent.findChildByKey(k, 0))
}

// Test that Node4 maintains sorted order when adding children.
func TestNode4AddChildTwicePreserveSorted(t *testing.T) {
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
func TestGrow(t *testing.T) {
	nodes := []*nodeRef{
		factory.newNode4(),
		factory.newNode16(),
		factory.newNode48(),
	}
	expected := []Kind{
		Node16,
		Node48,
		Node256,
	}

	for i, node := range nodes {
		newNode := toNode(node).grow()
		assert.Equal(t, expected[i], newNode.kind)
	}
}

// Test shrinking a node.
func TestShrink(t *testing.T) {
	nodes := []*nodeRef{
		factory.newNode256(),
		factory.newNode48(),
		factory.newNode16(),
		factory.newNode4(),
	}

	expected := []Kind{
		Node48,
		Node16,
		Node4,
		Leaf,
	}

	for i, node := range nodes {
		var minChildren int
		switch node.kind {
		case Node4:
			minChildren = node4Min
		case Node16:
			minChildren = node16Min
		case Node48:
			minChildren = node48Min
		case Node256:
			minChildren = node256Min
		}

		for j := 0; j < minChildren; j++ {
			if node.kind != Node4 {
				node.addChild(keyChar{ch: byte(i)}, factory.newNode4())
			} else {
				node.addChild(keyChar{ch: byte(i)}, factory.newLeaf(Key{byte(i)}, "value"))
			}
		}

		newNode := toNode(node).shrink()
		assert.Equal(t, expected[i], newNode.kind)
	}
}
