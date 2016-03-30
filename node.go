package art

import (
	"bytes"
	"sort"
	"unsafe"
)

type kind int
type Prefix [MAX_PREFIX_LENGTH]byte

func (k kind) String() string {
	var node2string = []string{"", "NODE4", "NODE16", "NODE48", "NODE256", "LEAF"}
	return node2string[k]
}

func (k Key) charAt(pos int) byte {
	if pos < 0 || pos >= len(k) {
		return 0
	}
	return k[pos]
}

type node struct {
	numChildren int
	prefixLen   int
	prefix      Prefix
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

func (an *artNode) SetPrefix(key Key, prefixLen int) {
	n := an.BaseNode()
	n.prefixLen = prefixLen
	for i := 0; i < min(MAX_PREFIX_LENGTH, prefixLen); i++ {
		n.prefix[i] = key[i]
	}
}

// newNode.SetKey(key, depth, longestPrefix)
// newNode4 := newNode.Node4()

// newNode4.prefixLen = longestPrefix
// if longestPrefix > 0 {
// 	newNode4.SetKey(key[depth:], min(MAX_PREFIX_LENGTH, longestPrefix))
// }

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

func (an *artNode) Grow() *artNode {
	return growNode(an)
}

func replaceRef(oldNode **artNode, newNode *artNode) {
	// factory.releaseNode(*oldNode)
	*oldNode = newNode
}

func replaceNode(oldNode *artNode, newNode *artNode) {
	// factory.releaseNode(oldNode)
	*oldNode = *newNode
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

	default:
		panic("Could find maximum")
	}

}

func (an *artNode) Index(c byte) int {
	switch an.kind {
	case NODE_4:
		n := an.Node4()
		for idx := 0; idx < n.numChildren; idx++ {
			if n.keys[idx] == c {
				return idx
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
		idx := sort.Search(int(n.numChildren), func(i int) bool { return c <= n.keys[i] })
		if idx < len(n.keys) && n.keys[idx] == c {
			return idx
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
			return idx - 1
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

func (an *artNode) AddChild(c byte, child *artNode) bool {
	switch an.kind {
	case NODE_4:
		return an.AddChild4(an.Node4(), c, child)

	case NODE_16:
		return an.AddChild16(an.Node16(), c, child)

	case NODE_48:
		return an.AddChild48(an.Node48(), c, child)

	case NODE_256:
		return an.AddChild256(an.Node256(), c, child)
	}
	return false
}

func (an *artNode) AddChild4(node *node4, c byte, child *artNode) bool {
	if node.numChildren < NODE_4_MAX {
		i := 0
		for ; i < node.numChildren; i++ {
			if c < node.keys[i] {
				break
			}
		}

		limit := node.numChildren - i
		for j := limit; limit > 0 && j > 0; j-- {
			node.keys[i+j] = node.keys[i+j-1]
			node.children[i+j] = node.children[i+j-1]
		}

		node.keys[i] = c
		node.children[i] = child
		node.numChildren++
		return false
	} else {
		newNode := an.Grow()
		newNode.AddChild(c, child)
		replaceNode(an, newNode)
		return true
	}
}

func (an *artNode) AddChild16(node *node16, c byte, child *artNode) bool {
	if node.numChildren < NODE_16_MAX {
		index := sort.Search(node.numChildren, func(i int) bool { return c <= node.keys[byte(i)] })
		for i := node.numChildren; i > index; i-- {
			node.keys[i] = node.keys[i-1]
			node.children[i] = node.children[i-1]
		}

		node.keys[index] = c
		node.children[index] = child
		node.numChildren++
		return false
	} else {
		newNode := an.Grow()
		newNode.AddChild(c, child)
		replaceNode(an, newNode)
		return true
	}
}

func (an *artNode) AddChild48(node *node48, c byte, child *artNode) bool {
	if node.numChildren < NODE_48_MAX {
		index := byte(0)
		for node.children[index] != nil {
			index++
		}

		node.keys[c] = index + 1
		node.children[index] = child
		node.numChildren++
		return false
	} else {
		newNode := an.Grow()
		newNode.AddChild(c, child)
		replaceNode(an, newNode)
		return true
	}
}

func (an *artNode) AddChild256(node *node256, c byte, child *artNode) bool {
	node.numChildren++
	node.children[c] = child
	return false
}
