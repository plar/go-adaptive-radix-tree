package art

import "unsafe"

func newTree() *tree {
	return &tree{}
}

func newNode4() *artNode {
	return &artNode{kind: NODE_4, ref: unsafe.Pointer(&node4{})}
}

func newNode16() *artNode {
	return &artNode{kind: NODE_16, ref: unsafe.Pointer(&node16{})}
}

func newNode48() *artNode {
	return &artNode{kind: NODE_48, ref: unsafe.Pointer(&node48{})}
}

func newNode256() *artNode {
	return &artNode{kind: NODE_256, ref: unsafe.Pointer(&node256{})}
}

func newNodeFrom(an *artNode) *artNode {
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
	n := newNode16()
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
	n := newNode48()
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
	n := newNode256()
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

func newLeaf(key Key, value interface{}) *artNode {
	clonedKey := make(Key, len(key))
	copy(clonedKey, key)
	return &artNode{kind: NODE_LEAF, ref: unsafe.Pointer(&leaf{key: clonedKey, value: value})}
}
