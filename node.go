package art

import (
	"bytes"
	"sort"
	"unsafe"
)

type kind int

func (k kind) String() string {
	var node2string = []string{"", "NODE4", "NODE16", "NODE48", "NODE256", "LEAF"}
	return node2string[k]
}

type node struct {
	numChildren int
	prefixLen   int
	prefix      [MAX_PREFIX_LENGTH]byte
}

type node4 struct {
	node
	keys     [NODE_4_MAX]byte
	children [NODE_4_MAX]*artNode
}

type node16 struct {
	node
	keys     [NODE_16_MAX]byte
	children [NODE_16_MAX]*artNode
}

type node48 struct {
	node
	keys     [NODE_256_MAX]byte
	children [NODE_48_MAX]*artNode
}

type node256 struct {
	node
	children [NODE_256_MAX]*artNode
}

type leaf struct {
	key   Key
	value interface{}
}

type artNode struct {
	kind kind
	ref  unsafe.Pointer
}

var nullNode *artNode = nil

func (n *node) SetKey(key Key, keyLen int) {
	for i := 0; i < keyLen; i++ {
		n.prefix[i] = key[i]
	}
}

func (n *node) CheckPrefix(key Key, depth int) int {
	idx, limit := 0, min(min(n.prefixLen, MAX_PREFIX_LENGTH), len(key)-depth)
	for ; idx < limit; idx++ {
		if n.prefix[idx] != key[idx+depth] {
			return idx
		}
	}
	return idx
}

func (l *leaf) Match(key Key, depth int) bool {
	return l != nil && len(l.key) == len(key) && bytes.Compare(l.key, key) == 0
}

func (an *artNode) PrefixMismatch(key Key, depth int) (idx int) {
	idx = 0

	n := an.BaseNode()
	limit := min(min(MAX_PREFIX_LENGTH, n.prefixLen), len(key)-depth)
	for ; idx < limit; idx++ {
		if n.prefix[idx] != key[idx+depth] {
			return idx
		}
	}

	if n.prefixLen > MAX_PREFIX_LENGTH {
		leaf := an.Minimum()
		limit = min(len(leaf.key), len(key)) - depth
		for ; idx < limit; idx++ {
			if leaf.key[idx+depth] != key[idx+depth] {
				return idx
			}
		}
	}

	return idx
}

func (an *artNode) Grow(c byte, child *artNode) *artNode {
	//fmt.Printf("Grow %v %v\n", c, child)
	return newNodeFrom(an)
}

// Find the minimum leaf under a artNode
func (an *artNode) Minimum() *leaf {
	if an == nil {
		return nil
	}

	if an.kind == NODE_LEAF {
		return an.Leaf()
	}

	switch an.kind {
	case NODE_4:
		if an.Node4().children[0] != nil {
			return an.Node4().children[0].Minimum()
		} else {
			return nil
		}

	case NODE_16:
		if an.Node16().children[0] != nil {
			return an.Node16().children[0].Minimum()
		} else {
			return nil
		}

	case NODE_48:
		idx := 0
		n := an.Node48()
		for n.keys[idx] == 0 {
			idx++
		}
		if n.children[n.keys[idx]-1] != nil {
			return n.children[n.keys[idx]-1].Minimum()
		} else {
			return nil
		}

	case NODE_256:
		idx := 0
		n := an.Node256()
		for n.children[idx] == nil {
			idx++
		}

		if idx < len(n.children) {
			return n.children[idx].Minimum()
		} else {
			return nil
		}

	default:
		//fmt.Printf("Kind: %+v\n", an)
		panic("Could find minimum")
	}

}

// Find the maximum leaf under a artNode
func (an *artNode) Maximum() *leaf {
	if an == nil {
		return nil
	}

	if an.kind == NODE_LEAF {
		return an.Leaf()
	}

	switch an.kind {
	case NODE_4:
		n := an.Node4()
		return n.children[n.numChildren-1].Maximum()

	case NODE_16:
		n := an.Node16()
		return n.children[n.numChildren-1].Maximum()

	case NODE_48:
		idx := 255
		n := an.Node48()
		for n.keys[idx] == 0 {
			idx--
		}
		return n.children[n.keys[idx]-1].Maximum()

	case NODE_256:
		idx := 255
		n := an.Node256()
		for n.children[idx] == nil {
			idx--
		}
		return n.children[idx].Maximum()
	}

	panic("Could find maximum")
}

func (an *artNode) Index(c byte) int {
	switch an.kind {
	case NODE_4:
		n := an.Node4()
		for i := 0; i < n.numChildren; i++ {
			if n.keys[i] == c {
				return i
			}
		}

	case NODE_16:
		// From the specification: First, the searched key is replicated and then compared to the
		// 16 keys stored in the inner artNode. In the next step, a
		// mask is created, because the artNode may have less than
		// 16 valid entries. The result of the comparison is converted to
		// a bit ﬁeld and the mask is applied. Finally, the bit
		// ﬁeld is converted to an index using the count trailing zero
		// instruction. Alternatively, binary search can be used
		// if SIMD instructions are not available.
		//
		// TODO It is currently unclear if golang has intentions of supporting SIMD instructions
		//      So until then, go-art will opt for Binary Search
		n := an.Node16()
		i := sort.Search(int(n.numChildren), func(i int) bool { return n.keys[i] >= c })
		if i < len(n.keys) && n.keys[i] == c {
			return i
		}

	case NODE_48:
		// ArtNodes of type NODE48 store the indicies in which to access their children
		// in the keys array which are byte-accessible by the desired key.
		// However, when this key array initialized, it contains many 0 value indicies.
		// In order to distinguish if a child actually exists, we increment this value
		// during insertion and decrease it during retrieval.
		n := an.Node48()
		idx := int(n.keys[c])
		if idx > 0 {
			return int(idx) - 1
		}

	case NODE_256:
		// ArtNodes of type NODE256 possibly have the simplest lookup algorithm.
		// Since all of their keys are byte-addressable, we can simply idx to the specific child with the key.
		return int(c)

	}

	return -1
}

func (an *artNode) FindChild(c byte) **artNode {
	if an == nil {
		return &nullNode
	}

	switch an.kind {
	case NODE_4:
		idx := an.Index(c)
		if idx >= 0 {
			return &an.Node4().children[idx]
		}

	case NODE_16:
		idx := an.Index(c)
		if idx >= 0 {
			return &an.Node16().children[idx]
		}

	case NODE_48:
		idx := an.Index(c)
		if idx >= 0 {
			return &an.Node48().children[idx]
		}

	case NODE_256:
		// NODE256 Types directly address their children with bytes
		n := an.Node256()
		child := n.children[c]
		if child != nil {
			return &n.children[c]
		}
	}

	return &nullNode

}

func (an *artNode) BaseNode() *node {
	return (*node)(an.ref)
}

func (an *artNode) Node4() *node4 {
	return (*node4)(an.ref)
}

func (an *artNode) Node16() *node16 {
	return (*node16)(an.ref)
}

func (an *artNode) Node48() *node48 {
	return (*node48)(an.ref)
}

func (an *artNode) Node256() *node256 {
	return (*node256)(an.ref)
}

func (an *artNode) Leaf() *leaf {
	return (*leaf)(an.ref)
}

func (an *artNode) AddChild(ref **artNode, c byte, child *artNode) {
	// fmt.Printf("AddChild, an: %+v c: %v child: %+v\n", an, c, child)
	switch an.kind {
	case NODE_4:
		n4 := an.Node4()
		if n4.numChildren < NODE_4_MAX {
			i := 0
			for ; i < n4.numChildren; i++ {
				if c < n4.keys[i] {
					break
				}
			}

			// // Shift to make room
			// memmove(n->keys+idx+1, n->keys+idx, n->n.num_children - idx);
			// memmove(n->children+idx+1, n->children+idx,
			// (n->n.num_children - idx)*sizeof(void*));

			limit := n4.numChildren - i
			// fmt.Printf("format2, %v, %v, %v\n", i, n4.keys, limit)
			for j := limit; limit > 0 && j > 0; j-- {
				n4.keys[i+j] = n4.keys[i+j-1]
				n4.children[i+j] = n4.children[i+j-1]
			}

			n4.keys[i] = c
			n4.children[i] = child
			n4.numChildren++
			// fmt.Printf("format2, %v, %v, %v\n", i, n4.keys, limit)

			//fmt.Printf("Add N4 %v %+v\n", i, n4)

		} else {
			// fmt.Printf("Growing... %v\n", DumpNode(an))
			nn := an.Grow(c, child)
			// fmt.Printf("Grow... %+v\n", DumpNode(nn))
			// fmt.Printf("Ref %p\n", *ref)
			nn.AddChild(ref, c, child)
			*an = *nn
			// fmt.Printf("Growed %+v\n", DumpNode(nn))
			// fmt.Printf("Ref %p\n", *ref)
		}

	case NODE_16:
		n16 := an.Node16()
		if n16.numChildren < NODE_16_MAX {
			index := sort.Search(n16.numChildren, func(i int) bool { return c <= n16.keys[byte(i)] })
			// fmt.Printf("Add N16 %v\n", index)

			for i := n16.numChildren; i > index; i-- {
				// if n16.keys[i-1] > c {
				n16.keys[i] = n16.keys[i-1]
				n16.children[i] = n16.children[i-1]
				// }
			}

			// fmt.Printf("Add N16 %v\n", index)
			n16.keys[index] = c
			n16.children[index] = child
			n16.numChildren++
			// fmt.Printf("Add N16 %v %+v\n", index, n16)
		} else {
			// fmt.Printf("Growing... %v\n", an)
			nn := an.Grow(c, child)
			// fmt.Printf("Grow... %+v\n", nn)
			nn.AddChild(ref, c, child)
			*an = *nn
			// fmt.Printf("Growed %+v\n", nn)
		}

	case NODE_48:
		n := an.Node48()
		if n.numChildren < NODE_48_MAX {
			index := byte(0)
			for n.children[index] != nil {
				index++
			}

			// fmt.Printf("N48: %+v, %v\n", n, index)

			n.keys[c] = index + 1
			n.children[index] = child
			n.numChildren++
		} else {
			// fmt.Printf("Growing... %v\n", an)
			nn := an.Grow(c, child)
			// fmt.Printf("Grow... %+v\n", nn)
			nn.AddChild(ref, c, child)
			*an = *nn
			// fmt.Printf("Growed %+v\n", nn)
		}

	case NODE_256:
		n256 := an.Node256()
		n256.numChildren++
		n256.children[c] = child
	}

}
