package art

import (
	"sync"

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

type poolObjFactory struct {
	artNodePool sync.Pool
	node4Pool   sync.Pool
	node16Pool  sync.Pool
	node48Pool  sync.Pool
	node256Pool sync.Pool
	leafPool    sync.Pool
}

type objFactory struct {
}

var factory = newObjFactory()

func newTree() *tree {
	return &tree{}
}

func newPoolObjFactory() nodeFactory {
	return &poolObjFactory{
		artNodePool: sync.Pool{New: func() interface{} { return new(artNode) }},
		node4Pool:   sync.Pool{New: func() interface{} { return new(node4) }},
		node16Pool:  sync.Pool{New: func() interface{} { return new(node16) }},
		node48Pool:  sync.Pool{New: func() interface{} { return new(node48) }},
		node256Pool: sync.Pool{New: func() interface{} { return new(node256) }},
		leafPool:    sync.Pool{New: func() interface{} { return new(leaf) }},
	}
}

func newObjFactory() nodeFactory {
	return &objFactory{}
}

func initArtNode(an *artNode, kind kind, ref unsafe.Pointer) {
	an.kind = kind
	an.ref = ref

	switch kind {
	case NODE_4, NODE_16, NODE_48, NODE_256:
		n := an.BaseNode()
		n.numChildren = 0
		n.prefixLen = 0
		// for i := range n.prefix {
		// 	n.prefix[i] = 0
		// }
	}

	switch an.kind {
	case NODE_4:
		n := an.Node4()
		for i := range n.keys {
			n.keys[i] = 0
			n.children[i] = nil
		}
	case NODE_16:
		n := an.Node16()
		for i := range n.keys {
			n.keys[i] = 0
			n.children[i] = nil
		}
	case NODE_48:
		n := an.Node48()
		for i := range n.keys {
			n.keys[i] = 0
		}
		for i := range n.children {
			n.children[i] = nil
		}
	case NODE_256:
		n := an.Node256()
		for i := range n.children {
			n.children[i] = nil
		}
	case NODE_LEAF:
		n := an.Leaf()
		n.key = nil
		n.value = nil
	}
}

// Pool based factory implementation

func (f *poolObjFactory) newNode4() *artNode {
	an := f.artNodePool.Get().(*artNode)
	node := f.node4Pool.Get().(*node4)
	initArtNode(an, NODE_4, unsafe.Pointer(node))
	return an
}

func (f *poolObjFactory) newNode16() *artNode {
	an := f.artNodePool.Get().(*artNode)
	node := f.node16Pool.Get().(*node16)
	initArtNode(an, NODE_16, unsafe.Pointer(node))
	return an
}

func (f *poolObjFactory) newNode48() *artNode {
	an := f.artNodePool.Get().(*artNode)
	node := f.node48Pool.Get().(*node48)
	initArtNode(an, NODE_48, unsafe.Pointer(node))
	return an
}

func (f *poolObjFactory) newNode256() *artNode {
	an := f.artNodePool.Get().(*artNode)
	node := f.node256Pool.Get().(*node256)
	initArtNode(an, NODE_256, unsafe.Pointer(node))
	return an
}

func (f *poolObjFactory) newLeaf(key Key, value interface{}) *artNode {
	an := f.artNodePool.Get().(*artNode)
	node := f.leafPool.Get().(*leaf)

	initArtNode(an, NODE_LEAF, unsafe.Pointer(node))

	clonedKey := make(Key, len(key))
	copy(clonedKey, key)
	node.key = clonedKey
	node.value = value

	return an
}

func (f *poolObjFactory) releaseNode(an *artNode) {
	if an == nil {
		return
	}

	// fmt.Printf("releaseNode %p\n", an)
	// return

	switch an.kind {
	case NODE_4:
		f.node4Pool.Put(an.Node4())

	case NODE_16:
		f.node16Pool.Put(an.Node16())

	case NODE_48:
		f.node48Pool.Put(an.Node48())

	case NODE_256:
		f.node256Pool.Put(an.Node256())

	case NODE_LEAF:
		f.leafPool.Put(an.Leaf())
	}

	f.artNodePool.Put(an)
}

// Simple obj factory implementation

func (f *objFactory) newNode4() *artNode {
	return &artNode{kind: NODE_4, ref: unsafe.Pointer(new(node4))}
}

func (f *objFactory) newNode16() *artNode {
	return &artNode{kind: NODE_16, ref: unsafe.Pointer(&node16{})}
}

func (f *objFactory) newNode48() *artNode {
	return &artNode{kind: NODE_48, ref: unsafe.Pointer(&node48{})}
}

func (f *objFactory) newNode256() *artNode {
	return &artNode{kind: NODE_256, ref: unsafe.Pointer(&node256{})}
}

func (f *objFactory) newLeaf(key Key, value interface{}) *artNode {
	clonedKey := make(Key, len(key))
	copy(clonedKey, key)
	return &artNode{kind: NODE_LEAF, ref: unsafe.Pointer(&leaf{key: clonedKey, value: value})}
}

func (f *objFactory) releaseNode(an *artNode) {
	// do nothing
}

func growNode(an *artNode) *artNode {
	//fmt.Printf("newNodeFrom %v\n", an)
	switch an.kind {
	case NODE_4:
		return newNode16From4(an)
	case NODE_16:
		return newNode48From16(an)
	case NODE_48:
		return newNode256From48(an)
	default:
		panic("newNodeFrom")
	}
	return nil
}

func newNode16From4(o *artNode) *artNode {
	n := factory.newNode16()
	copyMeta(n, o)

	d := n.Node16()
	s := o.Node4()

	for i := 0; i < s.numChildren; i++ {
		d.keys[i] = s.keys[i]
		d.children[i] = s.children[i]
	}
	return n
}

func newNode48From16(o *artNode) *artNode {
	n := factory.newNode48()
	copyMeta(n, o)

	d := n.Node48()
	s := o.Node16()

	for i := 0; i < s.numChildren; i++ {
		d.keys[s.keys[i]] = byte(i + 1)
		d.children[i] = s.children[i]
	}
	return n
}

func newNode256From48(o *artNode) *artNode {
	n := factory.newNode256()
	copyMeta(n, o)

	d := n.Node256()
	s := o.Node48()

	for i := 0; i < 256; i++ {
		if s.keys[i] > 0 {
			d.children[i] = s.children[s.keys[i]-1]
		}
	}
	return n
}
