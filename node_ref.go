package art

import (
	"unsafe"
)

// indexNotFound is a special index value
// that indicates that the index is not found.
const indexNotFound = -1

// nodeNotFound is a special node pointer
// that indicates that the node is not found
// for different internal tree operations.
var nodeNotFound *nodeRef

// nodeRef stores all available nodes, leaf and node type.
type nodeRef struct {
	// ref is a pointer to one of the node types:
	// leaf, node4, node16, node48, node256.
	ref  unsafe.Pointer
	kind Kind
}

// noder is an interface that defines methods that
// must be implemented by nodeRef and all node types.
type noder interface {
	minimum() *leaf
	maximum() *leaf

	index(byte) int
	childAt(int) **nodeRef
	zeroChild() **nodeRef

	canAddChild() bool
	grow() *nodeRef

	canShrinkNode() bool
	shrink() *nodeRef

	addChild(byte, bool, *nodeRef)
	deleteChild(byte, bool) int
}

// toNode converts the nodeRef to specific node type.
func toNode(an *nodeRef) noder {
	switch an.kind {
	case Node4:
		return an.node4()
	case Node16:
		return an.node16()
	case Node48:
		return an.node48()
	case Node256:
		return an.node256()
	}
	return noopNoder
}

// noop is a no-op noder implementation.
type noop struct{}

func (*noop) minimum() *leaf                { return nil }
func (*noop) maximum() *leaf                { return nil }
func (*noop) index(byte) int                { return indexNotFound }
func (*noop) childAt(int) **nodeRef         { return &nodeNotFound }
func (*noop) zeroChild() **nodeRef          { return &nodeNotFound }
func (*noop) canAddChild() bool             { return true }
func (*noop) grow() *nodeRef                { return nil }
func (*noop) canShrinkNode() bool           { return false }
func (*noop) shrink() *nodeRef              { return nil }
func (*noop) addChild(byte, bool, *nodeRef) {}
func (*noop) deleteChild(byte, bool) int    { return 0 }

// noopNoder is the default Noder implementation.
var noopNoder noder = &noop{}

// assert that all node types implement artNoder interface.
var _ noder = (*node4)(nil)
var _ noder = (*node16)(nil)
var _ noder = (*node48)(nil)
var _ noder = (*node256)(nil)

// assert that nodeRef implements public Node interface.
var _ Node = (*nodeRef)(nil)

// Kind returns the node kind.
func (nr *nodeRef) Kind() Kind {
	return nr.kind
}

// Key returns the node key.
func (nr *nodeRef) Key() Key {
	if nr.isLeaf() {
		return nr.leaf().key
	}
	return nil
}

// Value returns the node value.
func (nr *nodeRef) Value() Value {
	if nr.isLeaf() {
		return nr.leaf().value
	}
	return nil
}

// isLeaf returns true if the node is a leaf node.
func (nr *nodeRef) isLeaf() bool {
	return nr.kind == Leaf
}

// setPrefix sets the node prefix with the new prefix and prefix length.
func (nr *nodeRef) setPrefix(newPrefix []byte, prefixLen int) {
	n := nr.node()

	n.prefixLen = uint16(prefixLen)
	for i := 0; i < min(prefixLen, MaxPrefixLen); i++ {
		n.prefix[i] = newPrefix[i]
	}
}

// minimum returns the minimum leaf node under the current node.
func (nr *nodeRef) minimum() *leaf {
	if nr.kind == Leaf {
		return nr.leaf()
	}
	return toNode(nr).minimum()
}

// maximum returns the maximum leaf node under the current node.
func (nr *nodeRef) maximum() *leaf {
	if nr.kind == Leaf {
		return nr.leaf()
	}
	return toNode(nr).maximum()
}

// findChildByKey returns the child node pointer by the given key and key index.
func (nr *nodeRef) findChildByKey(key Key, keyIdx int) **nodeRef {
	n := toNode(nr)

	ch, valid := key.charAt(keyIdx)
	if !valid {
		return n.zeroChild()
	}

	idx := n.index(ch)
	if idx == indexNotFound {
		return &nodeNotFound
	}

	return n.childAt(idx)
}

func (nr *nodeRef) node() *node       { return (*node)(nr.ref) }    // node casts nodeRef to node.
func (nr *nodeRef) node4() *node4     { return (*node4)(nr.ref) }   // node4 casts nodeRef to node4.
func (nr *nodeRef) node16() *node16   { return (*node16)(nr.ref) }  // node16 casts nodeRef to node16.
func (nr *nodeRef) node48() *node48   { return (*node48)(nr.ref) }  // node48 casts nodeRef to node48.
func (nr *nodeRef) node256() *node256 { return (*node256)(nr.ref) } // node256 casts nodeRef to node256.
func (nr *nodeRef) leaf() *leaf       { return (*leaf)(nr.ref) }    // leaf casts nodeRef to leaf.

// addChild adds a new child node to the current node.
// If the node is full, it grows to the next node type.
func (nr *nodeRef) addChild(ch byte, valid bool, child *nodeRef) {
	n := toNode(nr)

	if n.canAddChild() {
		n.addChild(ch, valid, child)
	} else {
		newNode := n.grow()                // grow to the next node type
		newNode.addChild(ch, valid, child) // recursively add the child to the new node
		replaceNode(nr, newNode)           // replace the current node with the new node
	}
}

// deleteChild deletes the child node from the current node.
// If the node can shrink, it shrinks to the previous node type.
func (nr *nodeRef) deleteChild(ch byte, valid bool) bool {
	n := toNode(nr)

	n.deleteChild(ch, valid)
	if n.canShrinkNode() {
		shrankNode := n.shrink()
		replaceNode(nr, shrankNode)
		return true
	}

	return false
}

// match finds the first mismatched index between
// the node's prefix and the specified key prefix.
// This approach efficiently identifies the mismatch by
// leveraging the node's existing prefix data.
func (nr *nodeRef) match(key Key, keyOffset int) int /* 1st mismatch index*/ {
	// calc the remaining key length from offset
	keyRemaining := len(key) - keyOffset
	if keyRemaining < 0 {
		return 0
	}

	node := nr.node()
	// the maximum length we can check against the node's prefix
	prefLen := min(int(node.prefixLen), MaxPrefixLen)
	checkLen := min(prefLen, keyRemaining)

	// compare the key against the node's prefix
	for i := 0; i < checkLen; i++ {
		if node.prefix[i] != key[keyOffset+i] {
			return i
		}
	}

	return checkLen
}

// matchDeep returns the first index where the key mismatches,
// starting with the node's prefix(see match) and continuing with the minimum leaf's key.
// It returns the mismatch index or matches up to the key's end.
func (nr *nodeRef) matchDeep(key Key, keyOffset int) int /* mismatch index*/ {
	mismatchIdx := nr.match(key, keyOffset)
	if mismatchIdx < MaxPrefixLen {
		return mismatchIdx
	}

	leafKey := nr.minimum().key
	limit := min(len(leafKey), len(key)) - keyOffset

	for ; mismatchIdx < limit; mismatchIdx++ {
		if leafKey[keyOffset+mismatchIdx] != key[keyOffset+mismatchIdx] {
			break
		}
	}

	return mismatchIdx
}
