package art

import (
	"bytes"
	"sort"
	"unsafe"
)

type prefix [MaxPrefixLen]byte

// Base part of all the various nodes, except leaf
type node struct {
	numChildren int
	prefixLen   int
	prefix      prefix
}

// Node with 4 children
type node4 struct {
	node
	keys     [node4Max]byte
	children [node4Max]*artNode
}

// Node with 16 children
type node16 struct {
	node
	keys     [node16Max]byte
	children [node16Max]*artNode
}

// Node with 48 children
type node48 struct {
	node
	keys     [node256Max]byte
	children [node48Max]*artNode
}

// Node with 256 children
type node256 struct {
	node
	children [node256Max]*artNode
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

//
var nullNode *artNode

var node2string = []string{"Leaf", "Node4", "Node16", "Node48", "Node256"}

// String returns string representation of the Kind value
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
	return an.minChildren()
}

func (an *artNode) minChildren() int {
	switch an.kind {
	case Node4:
		return node4Min
	case Node16:
		return node16Min
	case Node48:
		return node48Min
	case Node256:
		return node256Min
	}

	return 0
}

func (an *artNode) maxChildren() int {
	switch an.kind {
	case Node4:
		return node4Max
	case Node16:
		return node16Max
	case Node48:
		return node48Max
	case Node256:
		return node256Max
	}

	return 0
}

func (an *artNode) isLeaf() bool {
	return an.kind == Leaf
}

func (an *artNode) setPrefix(key Key, prefixLen int) *artNode {
	node := an.node()
	node.prefixLen = prefixLen
	for i := 0; i < min(prefixLen, MaxPrefixLen); i++ {
		node.prefix[i] = key[i]
	}
	return an
}

func (l *leaf) match(key Key) bool {
	if key == nil || len(l.key) != len(key) {
		return false
	}
	return bytes.Compare(l.key[:len(key)], key) == 0
}

func (l *leaf) prefixMatch(key Key) bool {
	if key == nil || len(l.key) < len(key) {
		return false
	}
	return bytes.Compare(l.key[:len(key)], key) == 0
}

func (n *node) match(key Key, depth int) int /* mismatch index*/ {
	idx := 0
	limit := min(min(n.prefixLen, MaxPrefixLen), len(key)-depth)
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
	if mismatchIdx < MaxPrefixLen {
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
	*oldNode = newNode
}

func replaceNode(oldNode *artNode, newNode *artNode) {
	*oldNode = *newNode
}

// Find the minimum leaf under a artNode
func (an *artNode) minimum() *leaf {
	switch an.kind {
	case Leaf:
		return an.leaf()

	case Node4:
		node := an.node4()
		if node.children[0] != nil {
			return node.children[0].minimum()
		}

	case Node16:
		node := an.node16()
		if node.children[0] != nil {
			return node.children[0].minimum()
		}

	case Node48:
		idx := 0
		node := an.node48()
		for node.keys[idx] == 0 {
			idx++
		}
		if node.children[node.keys[idx]-1] != nil {
			return node.children[node.keys[idx]-1].minimum()
		}

	case Node256:
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
	case Leaf:
		return an.leaf()

	case Node4:
		node := an.node4()
		return node.children[node.numChildren-1].maximum()

	case Node16:
		node := an.node16()
		return node.children[node.numChildren-1].maximum()

	case Node48:
		idx := 255
		node := an.node48()
		for node.keys[idx] == 0 {
			idx--
		}
		return node.children[node.keys[idx]-1].maximum()

	case Node256:
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
	case Node4:
		node := an.node4()
		for idx := 0; idx < node.numChildren; idx++ {
			if node.keys[idx] == c {
				return idx
			}
		}

	case Node16:
		node := an.node16()
		idx := sort.Search(int(node.numChildren), func(i int) bool { return node.keys[i] >= c })
		if idx < len(node.keys) && node.keys[idx] == c {
			return idx
		}

	case Node48:
		node := an.node48()
		if idx := int(node.keys[c]); idx > 0 {
			return idx - 1
		}

	case Node256:
		return int(c)
	}

	return -1
}

func (an *artNode) findChild(c byte) **artNode {
	switch an.kind {
	case Node4:
		if idx := an.index(c); idx >= 0 {
			return &an.node4().children[idx]
		}

	case Node16:
		if idx := an.index(c); idx >= 0 {
			return &an.node16().children[idx]
		}

	case Node48:
		if idx := an.index(c); idx >= 0 {
			return &an.node48().children[idx]
		}

	case Node256:
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

func (an *artNode) _addChild4(c byte, child *artNode) bool {
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

	return false
}

func (an *artNode) _addChild16(c byte, child *artNode) bool {
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
	return false
}

func (an *artNode) _addChild48(c byte, child *artNode) bool {
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

	return false
}

func (an *artNode) _addChild256(c byte, child *artNode) bool {
	node := an.node256()
	node.numChildren++
	node.children[c] = child
	return false
}

func (an *artNode) addChild(c byte, child *artNode) bool {
	switch an.kind {
	case Node4:
		return an._addChild4(c, child)

	case Node16:
		return an._addChild16(c, child)

	case Node48:
		return an._addChild48(c, child)

	case Node256:
		return an._addChild256(c, child)

	default:
		return false
	}
}

func (an *artNode) _deleteChild4(c byte) int {
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
	return node.numChildren
}

func (an *artNode) _deleteChild16(c byte) int {
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
	return node.numChildren
}

func (an *artNode) _deleteChild48(c byte) int {
	node := an.node48()
	if idx := an.index(c); idx >= 0 && node.children[idx] != nil {
		node.children[idx] = nil
		node.keys[idx] = 0
		node.numChildren--
	}

	return node.numChildren
}

func (an *artNode) _deleteChild256(c byte) int {
	node := an.node256()
	if idx := an.index(c); node.children[idx] != nil {
		node.children[idx] = nil
		node.numChildren--
	}
	return node.numChildren
}

func (an *artNode) deleteChild(c byte) bool {
	numChildren := -1
	switch an.kind {
	case Node4:
		numChildren = an._deleteChild4(c)

	case Node16:
		numChildren = an._deleteChild16(c)

	case Node48:
		numChildren = an._deleteChild48(c)

	case Node256:
		numChildren = an._deleteChild256(c)
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

	for i, limit := 0, min(s.prefixLen, MaxPrefixLen); i < limit; i++ {
		d.prefix[i] = s.prefix[i]
	}

	return an
}

func (an *artNode) grow() *artNode {
	switch an.kind {
	case Node4:
		node := factory.newNode16().copyMeta(an)

		d := node.node16()
		s := an.node4()

		for i := 0; i < s.numChildren; i++ {
			d.keys[i] = s.keys[i]
			d.children[i] = s.children[i]
		}
		return node

	case Node16:
		node := factory.newNode48().copyMeta(an)

		d := node.node48()
		s := an.node16()

		for i := 0; i < s.numChildren; i++ {
			d.keys[s.keys[i]] = byte(i + 1)
			d.children[i] = s.children[i]
		}
		return node

	case Node48:
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

func (an *artNode) _shrink4() *artNode {
	node4 := an.node4()
	child := node4.children[0]
	if child.isLeaf() {
		return child
	}

	curPrefixLen := node4.prefixLen
	if curPrefixLen < MaxPrefixLen {
		node4.prefix[curPrefixLen] = node4.keys[0]
		curPrefixLen++
	}

	childNode := child.node()
	if curPrefixLen < MaxPrefixLen {
		childPrefixLen := min(childNode.prefixLen, MaxPrefixLen-curPrefixLen)
		for i := 0; i < childPrefixLen; i++ {
			node4.prefix[curPrefixLen+i] = childNode.prefix[i]
		}
		curPrefixLen += childPrefixLen
	}

	for i := 0; i < min(curPrefixLen, MaxPrefixLen); i++ {
		childNode.prefix[i] = node4.prefix[i]
	}
	childNode.prefixLen += node4.prefixLen + 1
	return child
}

func (an *artNode) _shrink16() *artNode {
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
}

func (an *artNode) _shrink48() *artNode {
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
}

func (an *artNode) _shrink256() *artNode {
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

func (an *artNode) shrink() *artNode {
	switch an.kind {
	case Node4:
		return an._shrink4()

	case Node16:
		return an._shrink16()

	case Node48:
		return an._shrink48()

	case Node256:
		return an._shrink256()

	default:
		return nil
	}
}
