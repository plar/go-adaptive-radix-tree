package art

import (
	"testing"

	//"fmt"

	"github.com/stretchr/testify/assert"
)

func TestNodeKind(t *testing.T) {
	n4 := factory.newNode4()
	assert.NotNil(t, n4)
	assert.Equal(t, Node4, n4.kind)

	n16 := factory.newNode16()
	assert.NotNil(t, n16)
	assert.Equal(t, Node16, n16.kind)

	n48 := factory.newNode48()
	assert.NotNil(t, n48)
	assert.Equal(t, Node48, n48.kind)

	n256 := factory.newNode256()
	assert.NotNil(t, n256)
	assert.Equal(t, Node256, n256.kind)

	leaf := factory.newLeaf([]byte("key"), "value")
	assert.NotNil(t, leaf)
	assert.Equal(t, Leaf, leaf.kind)
	assert.Equal(t, leaf.Key(), Key([]byte("key")))
	assert.Equal(t, leaf.Value().(string), "value")

	assert.Equal(t, "Node4", n4.kind.String())
	assert.Equal(t, "Node16", n16.kind.String())
	assert.Equal(t, "Node48", n48.kind.String())
	assert.Equal(t, "Node256", n256.kind.String())
	assert.Equal(t, "Leaf", leaf.kind.String())

	unknowNode := &artNode{kind: Kind(0xFF)}
	assert.Nil(t, unknowNode.maximum())
	assert.Nil(t, unknowNode.minimum())
}

func TestLeaf(t *testing.T) {
	leaf := factory.newLeaf([]byte("key"), "value")
	assert.NotNil(t, leaf)
	assert.Equal(t, Leaf, leaf.kind)

	assert.False(t, leaf.leaf().match([]byte("unknown-key")))

	// we cannot shrink/grow leaf node
	assert.Nil(t, leaf.shrink())
	assert.Nil(t, leaf.grow())
}

func TestLeafMatch(t *testing.T) {
	leaf := factory.newLeaf([]byte("key"), "value")
	assert.False(t, leaf.leaf().match([]byte("unknown-key")))
	assert.False(t, leaf.leaf().match(nil))

	assert.True(t, leaf.leaf().match([]byte("key")))
}

func TestLeafPrefixMatch(t *testing.T) {
	leaf := factory.newLeaf([]byte("key"), "value")
	assert.False(t, leaf.leaf().prefixMatch([]byte("unknown-key")))
	assert.False(t, leaf.leaf().prefixMatch(nil))

	assert.True(t, leaf.leaf().prefixMatch([]byte("ke")))
}

func TestNodeSetPrefix(t *testing.T) {
	n4 := factory.newNode4()
	assert.NotNil(t, n4)
	nn := n4.node()

	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n4.setPrefix(key, 2)

	assert.Equal(t, uint32(2), nn.prefixLen)
	assert.Equal(t, byte(1), nn.prefix[0])
	assert.Equal(t, byte(2), nn.prefix[1])

	n4.setPrefix(key, MaxPrefixLen)
	assert.Equal(t, uint32(MaxPrefixLen), nn.prefixLen)
	assert.Equal(t, byte(1), nn.prefix[0])
	assert.Equal(t, byte(2), nn.prefix[1])
	assert.Equal(t, byte(3), nn.prefix[2])
	assert.Equal(t, byte(4), nn.prefix[3])
}

func TestNodeMatchWithKey(t *testing.T) {
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n16 := factory.newNode16()

	n16.setPrefix([]byte{1, 2, 3, 4, 5, 66, 77, 88, 99}, 5)

	idx := n16.match(key, 0)
	assert.Equal(t, uint32(5), idx)
	idx = n16.match(key, 1)
	assert.Equal(t, uint32(0), idx)
	idx = n16.match(key, 100)
	assert.Equal(t, uint32(0), idx)
}

func TestNodeCopyMeta(t *testing.T) {
	newNode := factory.newNode4()
	node4 := newNode.node4()
	node4.numChildren = 2
	node4.prefixLen = 2
	node4.prefix[0] = byte(10)
	node4.prefix[1] = byte(20)

	assert.Equal(t, newNode, newNode.copyMeta(nil))

	newNode2 := factory.newNode4()
	node4 = newNode2.node4()
	node4.numChildren = 4
	node4.prefixLen = 3
	node4.prefix[0] = byte(11)
	node4.prefix[1] = byte(22)
	node4.prefix[2] = byte(33)

	assert.Equal(t, newNode, newNode.copyMeta(newNode2))
	assert.Equal(t, uint32(3), newNode.node().prefixLen)
	assert.Equal(t, uint16(4), newNode.node().numChildren)
	assert.Equal(t, byte(11), newNode.node().prefix[0])
	assert.Equal(t, byte(22), newNode.node().prefix[1])
	assert.Equal(t, byte(33), newNode.node().prefix[2])
}
func TestLeafFindChild(t *testing.T) {
	leaf := factory.newLeaf(Key("key"), "value")
	res := leaf.findChild('k', true)
	assert.Equal(t, &nodeNotFound, res)
}

func TestNodeAddChild(t *testing.T) {
	nodes := []*artNode{
		factory.newNode4(),
		factory.newNode16(),
		factory.newNode48(),
		factory.newNode256(),
	}

	for _, n := range nodes {
		var maxChildren int = -1
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
			leaf := factory.newLeaf([]byte{byte(i)}, i)
			n.addChild(byte(i), true, leaf)
		}

		for i := 0; i < maxChildren; i++ {
			leaf := n.findChild(byte(i), true)
			assert.NotNil(t, *leaf)
			assert.Equal(t, i, (*leaf).leaf().value.(int))
		}
	}
}

func TestNodeAddChildForLeaf(t *testing.T) {
	leaf := factory.newLeaf([]byte("key"), "value")
	assert.False(t, leaf.addChild('c', true, nil))
}

func TestNodeIndex(t *testing.T) {
	nodes := []*artNode{
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
			leaf := factory.newLeaf([]byte{byte(i)}, i)
			n.addChild(byte(i), true, leaf)
		}

		for i := 0; i < maxChildren; i++ {
			assert.Equal(t, i, n.index(byte(i)))
		}
	}
}

func TestNodesMinimumMaximum(t *testing.T) {
	// TODO: Merge nodes and inserts
	nodes := []*artNode{
		factory.newNode4(),
		factory.newNode16(),
		factory.newNode48(),
		factory.newNode256(),
	}

	inserts := []int{3, 15, 47, 255}

	for i, node := range nodes {
		for j := 1; j <= inserts[i]; j++ {
			node.addChild(byte(j), true, factory.newLeaf([]byte{byte(j)}, byte(j)))
		}

		minLeaf := node.minimum()
		assert.Equal(t, minLeaf.key, Key{1})
		assert.Equal(t, minLeaf.value.(byte), minLeaf.value.(byte))

		maxLeaf := node.maximum()
		assert.Equal(t, maxLeaf.key, Key{byte(inserts[i])})
		assert.Equal(t, maxLeaf.value.(byte), maxLeaf.value.(byte))
	}
}

func TestNode4AddChildAndFindChild(t *testing.T) {
	parent := factory.newNode4()
	child := factory.newNode4()
	parent.addChild(1, true, child)

	assert.Equal(t, uint16(1), parent.node().numChildren)
	assert.Equal(t, child, *parent.findChild(1, true))
}

func TestNode4AddChildTwicePreserveSorted(t *testing.T) {
	parent := factory.newNode4()
	child1 := factory.newNode4()
	child2 := factory.newNode4()
	parent.addChild(2, true, child1)
	parent.addChild(1, true, child2)

	assert.Equal(t, uint16(2), parent.node().numChildren)
	assert.Equal(t, byte(1), parent.node4().keys[0])
	assert.Equal(t, byte(2), parent.node4().keys[1])
}

func TestNode4AddChild4PreserveSorted(t *testing.T) {
	parent := factory.newNode4()
	for i := 4; i > 0; i-- {
		parent.addChild(byte(i), true, factory.newNode4())
	}

	assert.Equal(t, uint16(4), parent.node().numChildren)
	assert.Equal(t, []byte{
		byte(1),
		byte(2),
		byte(3),
		byte(4),
	}, parent.node4().keys[:])
}

func TestNode16AddChild16PreserveSorted(t *testing.T) {
	parent := factory.newNode16()
	for i := 16; i > 0; i-- {
		parent.addChild(byte(i), true, factory.newNode16())
	}

	assert.Equal(t, uint16(16), parent.node().numChildren)
	for i := 0; i < 16; i++ {
		assert.Equal(t, byte(i+1), parent.node16().keys[i])
	}
}

func TestGrow(t *testing.T) {
	nodes := []*artNode{factory.newNode4(), factory.newNode16(), factory.newNode48()}
	expected := []Kind{Node16, Node48, Node256}

	for i, node := range nodes {
		newNode := node.grow()
		assert.Equal(t, expected[i], newNode.kind)
	}
}

func TestShrink(t *testing.T) {
	nodes := []*artNode{
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
				node.addChild(byte(i), true, factory.newNode4())
			} else {
				node.addChild(byte(i), true, factory.newLeaf(Key{byte(i)}, "value"))
			}
		}

		newNode := node.shrink()
		assert.Equal(t, expected[i], newNode.kind)
	}
}
