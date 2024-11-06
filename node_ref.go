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
var nodeNotFound *nodeRef //nolint:gochecknoglobals

// nodeRef stores all available tree nodes leaf and nodeX types
// as a ref to *unsafe* pointer.
// The kind field is used to determine the type of the node.
type nodeRef struct {
	ref  unsafe.Pointer
	kind Kind
}

type nodeLeafer interface {
	minimum() *leaf
	maximum() *leaf
}

type nodeSizeManager interface {
	hasCapacityForChild() bool
	grow() *nodeRef

	isReadyToShrink() bool
	shrink() *nodeRef
}

type nodeOperations interface {
	addChild(kc keyChar, child *nodeRef)
	deleteChild(kc keyChar) int
}

type nodeChildren interface {
	childAt(idx int) **nodeRef
	childZero() **nodeRef
	allChildren() []*nodeRef
}

type nodeKeyIndexer interface {
	index(kc keyChar) int
}

// noder is an interface that defines methods that
// must be implemented by nodeRef and all node types.
// extra interfaces are used to group methods by their purpose
// and help with code readability.
type noder interface {
	nodeLeafer
	nodeOperations
	nodeChildren
	nodeKeyIndexer
	nodeSizeManager
}

// toNode converts the nodeRef to specific node type.
// the idea is to avoid type assertion in the code in multiple places.
func toNode(nr *nodeRef) noder {
	if nr == nil {
		return noopNoder
	}

	switch nr.kind { //nolint:exhaustive
	case Node4:
		return nr.node4()
	case Node16:
		return nr.node16()
	case Node48:
		return nr.node48()
	case Node256:
		return nr.node256()
	default:
		return noopNoder
	}
}

// noop is a no-op noder implementation.
type noop struct{}

func (*noop) minimum() *leaf             { return nil }
func (*noop) maximum() *leaf             { return nil }
func (*noop) index(keyChar) int          { return indexNotFound }
func (*noop) childAt(int) **nodeRef      { return &nodeNotFound }
func (*noop) childZero() **nodeRef       { return &nodeNotFound }
func (*noop) allChildren() []*nodeRef    { return nil }
func (*noop) hasCapacityForChild() bool  { return true }
func (*noop) grow() *nodeRef             { return nil }
func (*noop) isReadyToShrink() bool      { return false }
func (*noop) shrink() *nodeRef           { return nil }
func (*noop) addChild(keyChar, *nodeRef) {}
func (*noop) deleteChild(keyChar) int    { return 0 }

// noopNoder is the default Noder implementation.
var noopNoder noder = &noop{} //nolint:gochecknoglobals

// assert that all node types implement noder interface.
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

// Key returns the node key for leaf nodes.
// for nodeX types, it returns nil.
func (nr *nodeRef) Key() Key {
	if nr.isLeaf() {
		return nr.leaf().key
	}

	return nil
}

// Value returns the node value for leaf nodes.
// for nodeX types, it returns nil.
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

	n.prefixLen = uint16(prefixLen) //#nosec:G115
	for i := 0; i < minInt(prefixLen, maxPrefixLen); i++ {
		n.prefix[i] = newPrefix[i]
	}
}

// minimum returns itself if the node is a leaf node.
// otherwise it returns the minimum leaf node under the current node.
func (nr *nodeRef) minimum() *leaf {
	if nr.kind == Leaf {
		return nr.leaf()
	}

	return toNode(nr).minimum()
}

// maximum returns itself if the node is a leaf node.
// otherwise it returns the maximum leaf node under the current node.
func (nr *nodeRef) maximum() *leaf {
	if nr.kind == Leaf {
		return nr.leaf()
	}

	return toNode(nr).maximum()
}

// findChildByKey returns the child node reference for the given key.
func (nr *nodeRef) findChildByKey(key Key, keyOffset int) **nodeRef {
	n := toNode(nr)
	idx := n.index(key.charAt(keyOffset))

	return n.childAt(idx)
}

// nodeX/leaf casts the nodeRef to the specific nodeX/leaf type.
func (nr *nodeRef) node() *node       { return (*node)(nr.ref) }    // node casts nodeRef to node.
func (nr *nodeRef) node4() *node4     { return (*node4)(nr.ref) }   // node4 casts nodeRef to node4.
func (nr *nodeRef) node16() *node16   { return (*node16)(nr.ref) }  // node16 casts nodeRef to node16.
func (nr *nodeRef) node48() *node48   { return (*node48)(nr.ref) }  // node48 casts nodeRef to node48.
func (nr *nodeRef) node256() *node256 { return (*node256)(nr.ref) } // node256 casts nodeRef to node256.
func (nr *nodeRef) leaf() *leaf       { return (*leaf)(nr.ref) }    // leaf casts nodeRef to leaf.

// addChild adds a new child node to the current node.
// If the node is full, it grows to the next node type.
func (nr *nodeRef) addChild(kc keyChar, child *nodeRef) {
	n := toNode(nr)

	if n.hasCapacityForChild() {
		n.addChild(kc, child)
	} else {
		bigNode := n.grow()         // grow to the next node type
		bigNode.addChild(kc, child) // recursively add the child to the new node
		replaceNode(nr, bigNode)    // replace the current node with the new node
	}
}

// deleteChild deletes the child node from the current node.
// If the node can shrink after, it shrinks to the previous node type.
func (nr *nodeRef) deleteChild(kc keyChar) bool {
	shrank := false
	n := toNode(nr)
	n.deleteChild(kc)

	if n.isReadyToShrink() {
		shrank = true
		smallNode := n.shrink()    // shrink to the previous node type
		replaceNode(nr, smallNode) // replace the current node with the shrank node
	}

	return shrank
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

	n := nr.node()

	// the maximum length we can check against the node's prefix
	maxPrefixLen := minInt(int(n.prefixLen), maxPrefixLen)
	limit := minInt(maxPrefixLen, keyRemaining)

	// compare the key against the node's prefix
	for i := 0; i < limit; i++ {
		if n.prefix[i] != key[keyOffset+i] {
			return i
		}
	}

	return limit
}

// matchDeep returns the first index where the key mismatches,
// starting with the node's prefix(see match) and continuing with the minimum leaf's key.
// It returns the mismatch index or matches up to the key's end.
func (nr *nodeRef) matchDeep(key Key, keyOffset int) int /* mismatch index*/ {
	mismatchIdx := nr.match(key, keyOffset)
	if mismatchIdx < maxPrefixLen {
		return mismatchIdx
	}

	leafKey := nr.minimum().key
	limit := minInt(len(leafKey), len(key)) - keyOffset

	for ; mismatchIdx < limit; mismatchIdx++ {
		if leafKey[keyOffset+mismatchIdx] != key[keyOffset+mismatchIdx] {
			break
		}
	}

	return mismatchIdx
}
