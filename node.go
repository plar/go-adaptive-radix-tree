package art

import (
	"bytes"
	"sort"
	"unsafe"
)

type prefix [MAX_PREFIX_LENGTH]byte

// Base part of all the various nodes, except leaf
type node struct {
	numChildren int
	prefixLen   int
	prefix      prefix
}

// Node with 4 children
type node4 struct {
	node
	keys     [NODE_4_MAX]byte
	children [NODE_4_MAX]*artNode
}

// Node with 16 children
type node16 struct {
	node
	keys     [NODE_16_MAX]byte
	children [NODE_16_MAX]*artNode
}

// Node with 48 children
type node48 struct {
	node
	keys     [NODE_256_MAX]byte
	children [NODE_48_MAX]*artNode
}

// Node with 256 children
type node256 struct {
	node
	children [NODE_256_MAX]*artNode
}

// Leaf node with variable key length
type leaf struct {
	key   Key
	value interface{}
}

// ART node stores all available nodes, leaf and node type
type artNode struct {
	kind Kind
	ref  unsafe.Pointer
}

var nullNode *artNode = nil

var node2string = []string{"LEAF", "NODE4", "NODE16", "NODE48", "NODE256"}

func (k Kind) String() string {
	return node2string[k]
}

func (k Key) charAt(pos int) byte {
	if pos < 0 || pos >= len(k) {
		return 0
	}
	return k[pos]
}

// Node interface implemenation

func (an *artNode) Kind() Kind {
	return an.kind
}

func (an *artNode) Key() Key {
	if an.isLeaf() {
		return an.leaf().key
	}
	return nil
}

func (an *artNode) Value() Value {
	if an.isLeaf() {
		return an.leaf().value
	}
	return nil
}

func (an *artNode) shrinkThreshold() int {
	switch an.kind {
	case NODE_4:
		return NODE_4_SHRINK
	case NODE_16:
		return NODE_16_SHRINK
	case NODE_48:
		return NODE_48_SHRINK
	case NODE_256:
		return NODE_256_SHRINK
	}

	return 0
}

func (an *artNode) minChildren() int {
	switch an.kind {
	case NODE_4:
		return NODE_4_MIN
	case NODE_16:
		return NODE_16_MIN
	case NODE_48:
		return NODE_48_MIN
	case NODE_256:
		return NODE_256_MIN
	}

	return 0
}

func (an *artNode) maxChildren() int {
	switch an.kind {
	case NODE_4:
		return NODE_4_MAX
	case NODE_16:
		return NODE_16_MAX
	case NODE_48:
		return NODE_48_MAX
	case NODE_256:
		return NODE_256_MAX
	}

	return 0
}

func (an *artNode) isLeaf() bool {
	return an.kind == NODE_LEAF
}

func (an *artNode) setPrefix(key Key, prefixLen int) *artNode {
	node := an.node()
	node.prefixLen = prefixLen
	for i := 0; i < min(prefixLen, MAX_PREFIX_LENGTH); i++ {
		node.prefix[i] = key[i]
	}
	return an
}

func (l *leaf) match(key Key) bool {
	if key == nil || len(l.key) < len(key) {
		return false
	}
	return bytes.Compare(l.key[:len(key)], key) == 0
}

func (n *node) match(key Key, depth int) int /* mismatch index*/ {
	idx := 0
	limit := min(min(n.prefixLen, MAX_PREFIX_LENGTH), len(key)-depth)
	for ; idx < limit; idx++ {
		if n.prefix[idx] != key[idx+depth] {
			return idx
		}
	}
	return idx
}

func (an *artNode) matchDeep(key Key, depth int) int /* mismatch index*/ {
	node := an.node()
	mismatchIdx := node.match(key, depth)
	if mismatchIdx < MAX_PREFIX_LENGTH {
		return mismatchIdx
	}

	leaf := an.minimum()
	limit := min(len(leaf.key), len(key)) - depth
	for ; mismatchIdx < limit; mismatchIdx++ {
		if leaf.key[mismatchIdx+depth] != key[mismatchIdx+depth] {
			break
		}
	}

	return mismatchIdx
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
func (an *artNode) minimum() *leaf {
	switch an.kind {
	case NODE_LEAF:
		return an.leaf()

	case NODE_4:
		node := an.node4()
		if node.children[0] != nil {
			return node.children[0].minimum()
		}

	case NODE_16:
		node := an.node16()
		if node.children[0] != nil {
			return node.children[0].minimum()
		}

	case NODE_48:
		idx := 0
		node := an.node48()
		for node.keys[idx] == 0 {
			idx++
		}
		if node.children[node.keys[idx]-1] != nil {
			return node.children[node.keys[idx]-1].minimum()
		}

	case NODE_256:
		idx := 0
		node := an.node256()
		for node.children[idx] == nil {
			idx++
		}

		if idx < len(node.children) {
			return node.children[idx].minimum()
		}
	}
	// that should never happen in normal case
	return nil
}

func (an *artNode) maximum() *leaf {
	switch an.kind {
	case NODE_LEAF:
		return an.leaf()

	case NODE_4:
		node := an.node4()
		return node.children[node.numChildren-1].maximum()

	case NODE_16:
		node := an.node16()
		return node.children[node.numChildren-1].maximum()

	case NODE_48:
		idx := 255
		node := an.node48()
		for node.keys[idx] == 0 {
			idx--
		}
		return node.children[node.keys[idx]-1].maximum()

	case NODE_256:
		idx := 255
		node := an.node256()
		for node.children[idx] == nil {
			idx--
		}
		return node.children[idx].maximum()
	}

	// that should never happen in normal case
	return nil
}

func (an *artNode) index(c byte) int {
	switch an.kind {
	case NODE_4:
		node := an.node4()
		for idx := 0; idx < node.numChildren; idx++ {
			if node.keys[idx] == c {
				return idx
			}
		}

	case NODE_16:
		node := an.node16()
		idx := sort.Search(int(node.numChildren), func(i int) bool { return node.keys[i] >= c })
		if idx < len(node.keys) && node.keys[idx] == c {
			return idx
		}

	case NODE_48:
		node := an.node48()
		if idx := int(node.keys[c]); idx > 0 {
			return idx - 1
		}

	case NODE_256:
		return int(c)
	}

	return -1
}

func (an *artNode) findChild(c byte) **artNode {
	switch an.kind {
	case NODE_4:
		if idx := an.index(c); idx >= 0 {
			return &an.node4().children[idx]
		}

	case NODE_16:
		if idx := an.index(c); idx >= 0 {
			return &an.node16().children[idx]
		}

	case NODE_48:
		if idx := an.index(c); idx >= 0 {
			return &an.node48().children[idx]
		}

	case NODE_256:
		node := an.node256()
		if child := node.children[c]; child != nil {
			return &node.children[c]
		}
	}

	return &nullNode

}

func (an *artNode) node() *node {
	return (*node)(an.ref)
}

func (an *artNode) node4() *node4 {
	return (*node4)(an.ref)
}

func (an *artNode) node16() *node16 {
	return (*node16)(an.ref)
}

func (an *artNode) node48() *node48 {
	return (*node48)(an.ref)
}

func (an *artNode) node256() *node256 {
	return (*node256)(an.ref)
}

func (an *artNode) leaf() *leaf {
	return (*leaf)(an.ref)
}

func (an *artNode) addChild(c byte, child *artNode) bool {
	switch an.kind {
	case NODE_4:
		node := an.node4()
		if node.numChildren < an.maxChildren() {
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
		} else {
			newNode := an.grow()
			newNode.addChild(c, child)
			replaceNode(an, newNode)
			return true
		}

	case NODE_16:
		node := an.node16()
		if node.numChildren < an.maxChildren() {
			index := sort.Search(node.numChildren, func(i int) bool { return c <= node.keys[byte(i)] })
			for i := node.numChildren; i > index; i-- {
				node.keys[i] = node.keys[i-1]
				node.children[i] = node.children[i-1]
			}

			node.keys[index] = c
			node.children[index] = child
			node.numChildren++
		} else {
			newNode := an.grow()
			newNode.addChild(c, child)
			replaceNode(an, newNode)
			return true
		}

	case NODE_48:
		node := an.node48()
		if node.numChildren < an.maxChildren() {
			index := byte(0)
			for node.children[index] != nil {
				index++
			}

			node.keys[c] = index + 1
			node.children[index] = child
			node.numChildren++
		} else {
			newNode := an.grow()
			newNode.addChild(c, child)
			replaceNode(an, newNode)
			return true
		}

	case NODE_256:
		node := an.node256()
		node.numChildren++
		node.children[c] = child

	}

	return false
}

func (an *artNode) deleteChild(c byte) bool {
	numChildren := -1
	switch an.kind {
	case NODE_4:
		node := an.node4()
		if idx := an.index(c); idx >= 0 {
			node.numChildren--
			node.keys[idx] = 0
			node.children[idx] = nil
			for i := idx; i <= node.numChildren && i+1 < len(node.keys); i++ {
				node.keys[i] = node.keys[i+1]
				node.children[i] = node.children[i+1]
			}

			node.keys[node.numChildren] = 0
			node.children[node.numChildren] = nil

		}
		numChildren = node.numChildren

	case NODE_16:
		node := an.node16()
		if idx := an.index(c); idx >= 0 {
			node.numChildren--
			node.keys[idx] = 0
			node.children[idx] = nil
			for i := idx; i <= node.numChildren && i+1 < len(node.keys); i++ {
				node.keys[i] = node.keys[i+1]
				node.children[i] = node.children[i+1]
			}

			node.keys[node.numChildren] = 0
			node.children[node.numChildren] = nil
		}
		numChildren = node.numChildren

	case NODE_48:
		node := an.node48()
		if idx := an.index(c); idx >= 0 && node.children[idx] != nil {
			node.children[idx] = nil
			node.keys[idx] = 0
			node.numChildren--
		}
		numChildren = node.numChildren

	case NODE_256:
		node := an.node256()
		if idx := an.index(c); node.children[idx] != nil {
			node.children[idx] = nil
			node.numChildren--
		}
		numChildren = node.numChildren
	}

	if numChildren != -1 && numChildren < an.shrinkThreshold() {
		newNode := an.shrink()
		replaceNode(an, newNode)
		return true
	}
	return false

}

func (an *artNode) copyMeta(src *artNode) *artNode {
	if src == nil {
		return an
	}

	d := an.node()
	s := src.node()

	d.numChildren = s.numChildren
	d.prefixLen = s.prefixLen

	for i, limit := 0, min(s.prefixLen, MAX_PREFIX_LENGTH); i < limit; i++ {
		d.prefix[i] = s.prefix[i]
	}

	return an
}

func (an *artNode) grow() *artNode {
	switch an.kind {
	case NODE_4:
		node := factory.newNode16().copyMeta(an)

		d := node.node16()
		s := an.node4()

		for i := 0; i < s.numChildren; i++ {
			d.keys[i] = s.keys[i]
			d.children[i] = s.children[i]
		}
		return node

	case NODE_16:
		node := factory.newNode48().copyMeta(an)

		d := node.node48()
		s := an.node16()

		for i := 0; i < s.numChildren; i++ {
			d.keys[s.keys[i]] = byte(i + 1)
			d.children[i] = s.children[i]
		}
		return node

	case NODE_48:
		node := factory.newNode256().copyMeta(an)

		d := node.node256()
		s := an.node48()

		for i := 0; i < 256; i++ {
			if s.keys[i] > 0 {
				d.children[i] = s.children[s.keys[i]-1]
			}
		}
		return node
	default:
		return nil
	}

}

func (an *artNode) shrink() *artNode {
	switch an.kind {
	case NODE_4:
		node4 := an.node4()
		child := node4.children[0]
		if child.isLeaf() {
			return child
		}

		curPrefixLen := node4.prefixLen
		if curPrefixLen < MAX_PREFIX_LENGTH {
			node4.prefix[curPrefixLen] = node4.keys[0]
			curPrefixLen++
		}

		childNode := child.node()
		if curPrefixLen < MAX_PREFIX_LENGTH {
			childPrefixLen := min(childNode.prefixLen, MAX_PREFIX_LENGTH-curPrefixLen)
			for i := 0; i < childPrefixLen; i++ {
				node4.prefix[curPrefixLen+i] = childNode.prefix[i]
			}
			curPrefixLen += childPrefixLen
		}

		for i := 0; i < min(curPrefixLen, MAX_PREFIX_LENGTH); i++ {
			childNode.prefix[i] = node4.prefix[i]
		}
		childNode.prefixLen += node4.prefixLen + 1
		return child

	case NODE_16:
		node16 := an.node16()

		newNode := factory.newNode4().copyMeta(an)
		node4 := newNode.node4()
		node4.numChildren = 0
		for i := 0; i < len(node4.keys); i++ {
			node4.keys[i] = node16.keys[i]
			node4.children[i] = node16.children[i]
			node4.numChildren++
		}

		return newNode

		return nil
	case NODE_48:
		node48 := an.node48()

		newNode := factory.newNode16().copyMeta(an)
		node16 := newNode.node16()
		node16.numChildren = 0
		for i := 0; i < len(node48.keys); i++ {
			idx := node48.keys[byte(i)]
			if idx <= 0 {
				continue
			}

			if child := node48.children[idx-1]; child != nil {
				node16.children[node16.numChildren] = child
				node16.keys[node16.numChildren] = byte(i)
				node16.numChildren++
			}
		}

		return newNode

	case NODE_256:
		node256 := an.node256()

		newNode := factory.newNode48().copyMeta(an)
		node48 := newNode.node48()
		node48.numChildren = 0
		for i := 0; i < len(node256.children); i++ {
			child := node256.children[byte(i)]
			if child != nil {
				node48.children[node48.numChildren] = child
				node48.keys[byte(i)] = byte(node48.numChildren + 1)
				node48.numChildren++
			}
		}

		return newNode
	}
	return nil
}
