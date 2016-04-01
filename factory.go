package art

import (
	// "sync"

	"unsafe"
)

type nodeFactory interface {
	newNode4() *artNode
	newNode16() *artNode
	newNode48() *artNode
	newNode256() *artNode
	newLeaf(key Key, value interface{}) *artNode

	releaseNode(n *artNode)
}

var factory = newObjFactory()

func newTree() *tree {
	return &tree{}
}

// type poolObjFactory struct {
// 	artNodePool sync.Pool
// 	node4Pool   sync.Pool
// 	node16Pool  sync.Pool
// 	node48Pool  sync.Pool
// 	node256Pool sync.Pool
// 	leafPool    sync.Pool
// }

type objFactory struct{}

func newObjFactory() nodeFactory {
	return &objFactory{}
}

// Simple obj factory implementation
func (f *objFactory) newNode4() *artNode {
	return &artNode{kind: Node4, ref: unsafe.Pointer(new(node4))}
}

func (f *objFactory) newNode16() *artNode {
	return &artNode{kind: Node16, ref: unsafe.Pointer(&node16{})}
}

func (f *objFactory) newNode48() *artNode {
	return &artNode{kind: Node48, ref: unsafe.Pointer(&node48{})}
}

func (f *objFactory) newNode256() *artNode {
	return &artNode{kind: Node256, ref: unsafe.Pointer(&node256{})}
}

func (f *objFactory) newLeaf(key Key, value interface{}) *artNode {
	clonedKey := make(Key, len(key))
	copy(clonedKey, key)
	return &artNode{kind: Leaf, ref: unsafe.Pointer(&leaf{key: clonedKey, value: value})}
}

func (f *objFactory) releaseNode(an *artNode) {
	// do nothing
}

// func newPoolObjFactory() nodeFactory {
// 	return &poolObjFactory{
// 		artNodePool: sync.Pool{New: func() interface{} { return new(artNode) }},
// 		node4Pool:   sync.Pool{New: func() interface{} { return new(node4) }},
// 		node16Pool:  sync.Pool{New: func() interface{} { return new(node16) }},
// 		node48Pool:  sync.Pool{New: func() interface{} { return new(node48) }},
// 		node256Pool: sync.Pool{New: func() interface{} { return new(node256) }},
// 		leafPool:    sync.Pool{New: func() interface{} { return new(leaf) }},
// 	}
// }

// func initArtNode(an *artNode, kind Kind, ref unsafe.Pointer) {
// 	an.kind = kind
// 	an.ref = ref

// 	switch kind {
// 	case Node4, Node16, Node48, Node256:
// 		n := an.node()
// 		n.numChildren = 0
// 		n.prefixLen = 0
// 		// for i := range n.prefix {
// 		// 	n.prefix[i] = 0
// 		// }
// 	}

// 	switch an.kind {
// 	case Node4:
// 		n := an.node4()
// 		for i := range n.keys {
// 			n.keys[i] = 0
// 			n.children[i] = nil
// 		}
// 	case Node16:
// 		n := an.node16()
// 		for i := range n.keys {
// 			n.keys[i] = 0
// 			n.children[i] = nil
// 		}
// 	case Node48:
// 		n := an.node48()
// 		for i := range n.keys {
// 			n.keys[i] = 0
// 		}
// 		for i := range n.children {
// 			n.children[i] = nil
// 		}
// 	case Node256:
// 		n := an.node256()
// 		for i := range n.children {
// 			n.children[i] = nil
// 		}
// 	case Leaf:
// 		n := an.leaf()
// 		n.key = nil
// 		n.value = nil
// 	}
// }

// // Pool based factory implementation

// func (f *poolObjFactory) newNode4() *artNode {
// 	an := f.artNodePool.Get().(*artNode)
// 	node := f.node4Pool.Get().(*node4)
// 	initArtNode(an, Node4, unsafe.Pointer(node))
// 	return an
// }

// func (f *poolObjFactory) newNode16() *artNode {
// 	an := f.artNodePool.Get().(*artNode)
// 	node := f.node16Pool.Get().(*node16)
// 	initArtNode(an, Node16, unsafe.Pointer(node))
// 	return an
// }

// func (f *poolObjFactory) newNode48() *artNode {
// 	an := f.artNodePool.Get().(*artNode)
// 	node := f.node48Pool.Get().(*node48)
// 	initArtNode(an, Node48, unsafe.Pointer(node))
// 	return an
// }

// func (f *poolObjFactory) newNode256() *artNode {
// 	an := f.artNodePool.Get().(*artNode)
// 	node := f.node256Pool.Get().(*node256)
// 	initArtNode(an, Node256, unsafe.Pointer(node))
// 	return an
// }

// func (f *poolObjFactory) newLeaf(key Key, value interface{}) *artNode {
// 	an := f.artNodePool.Get().(*artNode)
// 	node := f.leafPool.Get().(*leaf)

// 	initArtNode(an, Leaf, unsafe.Pointer(node))

// 	clonedKey := make(Key, len(key))
// 	copy(clonedKey, key)
// 	node.key = clonedKey
// 	node.value = value

// 	return an
// }

// func (f *poolObjFactory) releaseNode(an *artNode) {
// 	if an == nil {
// 		return
// 	}

// 	// fmt.Printf("releaseNode %p\n", an)
// 	// return

// 	switch an.kind {
// 	case Node4:
// 		f.node4Pool.Put(an.node4())

// 	case Node16:
// 		f.node16Pool.Put(an.node16())

// 	case Node48:
// 		f.node48Pool.Put(an.node48())

// 	case Node256:
// 		f.node256Pool.Put(an.node256())

// 	case Leaf:
// 		f.leafPool.Put(an.leaf())
// 	}

// 	f.artNodePool.Put(an)
// }
