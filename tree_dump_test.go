package art

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTreeDump4(t *testing.T) {
	n4 := factory.newNode4()
	n4leaf := factory.newLeaf([]byte("key4"), "value4")
	n4._addChild4('k', n4leaf)
	tree := &tree{root: n4}
	o := tree.String()
	assert.Contains(t, o, "Node4")
	assert.Contains(t, o, "prefix(0): [0 0 0 0 0 0 0 0 0 0]")
	assert.Contains(t, o, "keys: [107 0 0 0] [k···]")
	assert.Contains(t, o, "children(1)")
	assert.Contains(t, o, "Leaf")
	assert.Contains(t, o, "key: [107 101 121 52] [key4]")
	assert.Contains(t, o, "val: value4")
}

func TestTreeDump4BinaryValue(t *testing.T) {
	n4 := factory.newNode4()
	n4leaf := factory.newLeaf([]byte("key4"), []byte("value4"))
	n4._addChild4('k', n4leaf)
	tree := &tree{root: n4}
	o := tree.String()
	assert.Contains(t, o, "Node4")
	assert.Contains(t, o, "prefix(0): [0 0 0 0 0 0 0 0 0 0]")
	assert.Contains(t, o, "keys: [107 0 0 0] [k···]")
	assert.Contains(t, o, "children(1)")
	assert.Contains(t, o, "Leaf")
	assert.Contains(t, o, "key: [107 101 121 52] [key4]")
	assert.Contains(t, o, "val: value4", "Binary value should be print as a string")
}

func TestTreeDump4Int(t *testing.T) {
	n4 := factory.newNode4()
	n4leaf := factory.newLeaf([]byte("key4"), 4)
	n4._addChild4('k', n4leaf)
	tree := &tree{root: n4}
	o := tree.String()
	assert.Contains(t, o, "Node4")
	assert.Contains(t, o, "prefix(0): [0 0 0 0 0 0 0 0 0 0]")
	assert.Contains(t, o, "keys: [107 0 0 0] [k···]")
	assert.Contains(t, o, "children(1)")
	assert.Contains(t, o, "Leaf")
	assert.Contains(t, o, "key: [107 101 121 52] [key4]")
	assert.Contains(t, o, "val: 4", "Int value should be print as a number")
}

func TestTreeDump2NodeWithIntValue(t *testing.T) {
	n16 := factory.newNode16()
	n16_2 := factory.newNode16()
	n16_2leaf := factory.newLeaf([]byte("zima"), 4)
	n16_2.addChild('z', n16_2leaf)

	n16leaf := factory.newLeaf([]byte("key4"), 4)
	c4leaf := factory.newLeaf([]byte("cey4"), 44)
	n16._addChild4('k', n16leaf)
	n16._addChild4('c', c4leaf)
	n16._addChild4('z', n16_2)
	tree := &tree{root: n16}
	o := tree.String()
	assert.Contains(t, o, "Node16")
	assert.Contains(t, o, "prefix(0): [0 0 0 0 0 0 0 0 0 0]")
	assert.Contains(t, o, "keys: [99 107 122 ")
	assert.Contains(t, o, "keys: [122 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]")
	assert.Contains(t, o, "children(3)")
	assert.Contains(t, o, "children(1)")
	assert.Contains(t, o, "Leaf")
	assert.Contains(t, o, "key: [107 101 121 52] [key4]")
	assert.Contains(t, o, "key: [122 105 109 97] [zima]")
	assert.Contains(t, o, "val: 4", "Int value should be print as a number")
}

func TestTreeDump16(t *testing.T) {
	n16 := factory.newNode16()
	n16leaf := factory.newLeaf([]byte("key16"), "value16")
	n16._addChild16('k', n16leaf)
	tree := &tree{root: n16}
	o := tree.String()
	// fmt.Println(o)
	assert.Contains(t, o, "Node16")
	assert.Contains(t, o, "prefix(0): [0 0 0 0 0 0 0 0 0 0]")
	assert.Contains(t, o, "keys: [107 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0] [k···············]")
	assert.Contains(t, o, "children(1)")
	assert.Contains(t, o, "Leaf")
	assert.Contains(t, o, "key: [107 101 121 49 54] [key16]")
	assert.Contains(t, o, "val: value16")
}

func TestTreeDump48(t *testing.T) {
	n48 := factory.newNode48()
	n48leaf := factory.newLeaf([]byte("key48"), "value48")
	n48._addChild48('k', n48leaf)
	tree := &tree{root: n48}
	o := tree.String()
	assert.Contains(t, o, "Node48")
	assert.Contains(t, o, "prefix(0): [0 0 0 0 0 0 0 0 0 0]")
	assert.Contains(t, o, "keys: [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]")
	assert.Contains(t, o, "children(1)")
	assert.Contains(t, o, "Leaf")
	assert.Contains(t, o, "key: [107 101 121 52 56] [key48]")
	assert.Contains(t, o, "val: value48")
}

func TestTreeDump256(t *testing.T) {
	n256 := factory.newNode256()
	n256leaf := factory.newLeaf([]byte("key256"), "value256")
	n256._addChild256('k', n256leaf)
	tree := &tree{root: n256}
	o := tree.String()
	// fmt.Println(o)
	assert.Contains(t, o, "Node256")
	assert.Contains(t, o, "prefix(0): [0 0 0 0 0 0 0 0 0 0]")
	assert.Contains(t, o, "children(1)")
	assert.Contains(t, o, "Leaf")
	assert.Contains(t, o, "key: [107 101 121 50 53 54] [key256]")
	assert.Contains(t, o, "val: value256")
}
