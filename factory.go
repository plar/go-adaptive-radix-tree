package art

import (
	"unsafe"
)

// nodeFactory is an interface for creating various types of ART nodes,
// including nodes with different capacities and leaf nodes.
type nodeFactory interface {
	newNode4() *nodeRef
	newNode16() *nodeRef
	newNode48() *nodeRef
	newNode256() *nodeRef

	newLeaf(key Key, value interface{}) *nodeRef
}

// make sure that objFactory implements all methods of nodeFactory interface.
var _ nodeFactory = &objFactory{}

//nolint:gochecknoglobals
var (
	factory = newObjFactory()
)

// newTree creates a new tree.
func newTree() *tree {
	return &tree{
		version: 0,
		root:    nil,
		size:    0,
	}
}

// objFactory implements nodeFactory interface.
type objFactory struct{}

// newObjFactory creates a new objFactory.
func newObjFactory() nodeFactory {
	return &objFactory{}
}

// Simple obj factory implementation.
func (f *objFactory) newNode4() *nodeRef {
	return &nodeRef{
		kind: Node4,
		ref:  unsafe.Pointer(new(node4)), //#nosec:G103
	}
}

// newNode16 creates a new node16 as a nodeRef.
func (f *objFactory) newNode16() *nodeRef {
	return &nodeRef{
		kind: Node16,
		ref:  unsafe.Pointer(new(node16)), //#nosec:G103
	}
}

// newNode48 creates a new node48 as a nodeRef.
func (f *objFactory) newNode48() *nodeRef {
	return &nodeRef{
		kind: Node48,
		ref:  unsafe.Pointer(new(node48)), //#nosec:G103
	}
}

// newNode256 creates a new node256 as a nodeRef.
func (f *objFactory) newNode256() *nodeRef {
	return &nodeRef{
		kind: Node256,
		ref:  unsafe.Pointer(new(node256)), //#nosec:G103
	}
}

// newLeaf creates a new leaf node as a nodeRef.
// It clones the key to avoid any source key mutation.
func (f *objFactory) newLeaf(key Key, value interface{}) *nodeRef {
	keyClone := make(Key, len(key))
	copy(keyClone, key)

	return &nodeRef{
		kind: Leaf,
		ref: unsafe.Pointer(&leaf{ //#nosec:G103
			key:   keyClone,
			value: value,
		}),
	}
}
