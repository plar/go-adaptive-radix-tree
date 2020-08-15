package art

import (
	"bytes"
	"sort"
	"unsafe"
)

type prefix [MaxPrefixLen]byte

// a key with null suffix will be stored as a last child at the `children` array
// see +1 for each children definition in each nodeX struct

// Node with 4 children
type node4 struct {
	keys     [node4Max]byte
	present  [node4Max]byte
	children [node4Max + 1]*artNode
}

// Node with 16 children
type node16 struct {
	keys     [node16Max]byte
	present  [node16Max]byte
	children [node16Max + 1]*artNode
}

// Node with 48 children
type node48 struct {
	keys     [node256Max]byte
	present  [node256Max]byte
	children [node48Max + 1]*artNode
}

// Node with 256 children
type node256 struct {
	children [node256Max + 1]*artNode
}

// Leaf node with variable key length
type leaf struct {
	key   Key
	value interface{}
}

// ART node stores all available nodes, leaf and node type
type artNode struct {
	prefix      prefix
	ref         unsafe.Pointer
	prefixLen   uint32
	numChildren uint16
	kind        Kind
}

// String returns string representation of the Kind value
func (k Kind) String() string {
	return []string{"Leaf", "Node4", "Node16", "Node48", "Node256"}[k]
}

func (k Key) charAt(pos int) byte {
	if pos < 0 || pos >= len(k) {
		return 0
	}
	return k[pos]
}

func (k Key) valid(pos int) bool {
	return pos >= 0 && pos < len(k)
}

// Node interface implementation
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

func (an *artNode) shrinkThreshold() uint16 {
	return an.minChildren()
}

func (an *artNode) minChildren() uint16 {
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

func (an *artNode) maxChildren() uint16 {
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

func (an *artNode) setPrefix(key Key, prefixLen uint32) *artNode {
	an.prefixLen = prefixLen
	for i := uint32(0); i < min(prefixLen, MaxPrefixLen); i++ {
		an.prefix[i] = key[i]
	}

	return an
}

func (an *artNode) matchDeep(key Key, depth uint32) uint32 /* mismatch index*/ {
	mismatchIdx := an.match(key, depth)
	if mismatchIdx < MaxPrefixLen {
		return mismatchIdx
	}

	leaf := an.minimum()
	limit := min(uint32(len(leaf.key)), uint32(len(key))) - depth
	for ; mismatchIdx < limit; mismatchIdx++ {
		if leaf.key[mismatchIdx+depth] != key[mismatchIdx+depth] {
			break
		}
	}

	return mismatchIdx
}

// Find the minimum leaf under a artNode
func (an *artNode) minimum() *leaf {
	switch an.kind {
	case Leaf:
		return an.leaf()

	case Node4:
		node := an.node4()
		if node.children[an.maxChildren()] != nil {
			return node.children[an.maxChildren()].minimum()
		} else if node.children[0] != nil {
			return node.children[0].minimum()
		}

	case Node16:
		node := an.node16()
		if node.children[an.maxChildren()] != nil {
			return node.children[an.maxChildren()].minimum()
		} else if node.children[0] != nil {
			return node.children[0].minimum()
		}

	case Node48:
		node := an.node48()
		if node.children[an.maxChildren()] != nil {
			return node.children[an.maxChildren()].minimum()
		} else {
			idx := 0
			for node.present[idx] == 0 {
				idx++
			}
			if node.children[node.keys[idx]] != nil {
				return node.children[node.keys[idx]].minimum()
			}
		}

	case Node256:
		node := an.node256()
		if node.children[an.maxChildren()] != nil {
			return node.children[an.maxChildren()].minimum()
		} else if len(node.children) > 0 {
			idx := 0
			for ; node.children[idx] == nil; idx++ {
				// find 1st non empty
			}
			return node.children[idx].minimum()
		}
	}

	return nil // that should never happen in normal case
}

func (an *artNode) maximum() *leaf {
	switch an.kind {
	case Leaf:
		return an.leaf()

	case Node4:
		node := an.node4()
		return node.children[an.numChildren-1].maximum()

	case Node16:
		node := an.node16()
		return node.children[an.numChildren-1].maximum()

	case Node48:
		idx := node256Max - 1
		node := an.node48()
		for node.present[idx] == 0 {
			idx--
		}
		return node.children[node.keys[idx]].maximum()

	case Node256:
		idx := node256Max - 1
		node := an.node256()
		for node.children[idx] == nil {
			idx--
		}
		return node.children[idx].maximum()
	}

	return nil // that should never happen in normal case
}

func (an *artNode) index(c byte) int {
	switch an.kind {
	case Node4:
		node := an.node4()
		for idx := 0; idx < int(an.numChildren); idx++ {
			if node.keys[idx] == c {
				return idx
			}
		}

	case Node16:
		node := an.node16()
		idx := sort.Search(int(an.numChildren), func(i int) bool {
			return node.keys[i] >= c
		})

		if idx < len(node.keys) && node.keys[idx] == c {
			return idx
		}

	case Node48:
		node := an.node48()
		if s := node.present[c]; s > 0 {
			if idx := int(node.keys[c]); idx >= 0 {
				return idx
			}
		}

	case Node256:
		return int(c)
	}

	return -1 // not found
}

func (an *artNode) findChild(c byte, valid bool) **artNode {
	idx := 0
	if valid {
		idx = an.index(c)
	} else {
		idx = int(an.maxChildren())
	}

	if idx >= 0 {
		switch an.kind {
		case Node4:
			return &an.node4().children[idx]

		case Node16:
			return &an.node16().children[idx]

		case Node48:
			return &an.node48().children[idx]

		case Node256:
			return &an.node256().children[idx]
		}
	}

	var nullNode *artNode
	return &nullNode
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

func (an *artNode) _addChild4(c byte, valid bool, child *artNode) bool {
	node := an.node4()
	if an.numChildren < an.maxChildren() {
		if !valid {
			node.children[an.maxChildren()] = child
		} else {
			i := uint16(0)
			for ; i < an.numChildren; i++ {
				if c < node.keys[i] {
					break
				}
			}

			limit := an.numChildren - i
			for j := limit; limit > 0 && j > 0; j-- {
				node.keys[i+j] = node.keys[i+j-1]
				node.present[i+j] = node.present[i+j-1]
				node.children[i+j] = node.children[i+j-1]
			}
			node.keys[i] = c
			node.present[i] = 1
			node.children[i] = child
			an.numChildren++
		}
		return false
	} else {
		newNode := an.grow()
		newNode.addChild(c, valid, child)
		replaceNode(an, newNode)
		return true
	}
}

func (an *artNode) _addChild16(c byte, valid bool, child *artNode) bool {
	node := an.node16()
	if an.numChildren < an.maxChildren() {
		if !valid {
			node.children[an.maxChildren()] = child
		} else {
			index := sort.Search(int(an.numChildren), func(i int) bool {
				return c <= node.keys[byte(i)]
			})

			for i := an.numChildren; i > uint16(index); i-- {
				node.keys[i] = node.keys[i-1]
				node.present[i] = node.present[i-1]
				node.children[i] = node.children[i-1]
			}

			node.keys[index] = c
			node.present[index] = 1
			node.children[index] = child
			an.numChildren++
		}

		return false
	} else {
		newNode := an.grow()
		newNode.addChild(c, valid, child)
		replaceNode(an, newNode)

		return true
	}
}

func (an *artNode) _addChild48(c byte, valid bool, child *artNode) bool {
	node := an.node48()
	if an.numChildren < an.maxChildren() {
		if !valid {
			node.children[an.maxChildren()] = child
		} else {
			index := byte(0)
			for node.children[index] != nil {
				index++
			}

			node.keys[c] = index
			node.present[c] = 1
			node.children[index] = child
			an.numChildren++
		}

		return false
	} else {
		newNode := an.grow()
		newNode.addChild(c, valid, child)
		replaceNode(an, newNode)

		return true
	}
}

func (an *artNode) _addChild256(c byte, valid bool, child *artNode) bool {
	node := an.node256()
	if !valid {
		node.children[an.maxChildren()] = child
	} else {
		an.numChildren++
		node.children[c] = child
	}

	return false
}

func (an *artNode) addChild(c byte, valid bool, child *artNode) bool {
	switch an.kind {
	case Node4:
		return an._addChild4(c, valid, child)

	case Node16:
		return an._addChild16(c, valid, child)

	case Node48:
		return an._addChild48(c, valid, child)

	case Node256:
		return an._addChild256(c, valid, child)
	}

	return false
}

func (an *artNode) _deleteChild4(c byte, valid bool) uint16 {
	node := an.node4()
	if !valid {
		node.children[an.maxChildren()] = nil
	} else if idx := an.index(c); idx >= 0 {
		an.numChildren--
		node.keys[idx] = 0
		node.present[idx] = 0
		node.children[idx] = nil

		for i := uint16(idx); i <= an.numChildren && i+1 < node4Max; i++ {
			node.keys[i] = node.keys[i+1]
			node.present[i] = node.present[i+1]
			node.children[i] = node.children[i+1]
		}

		node.keys[an.numChildren] = 0
		node.present[an.numChildren] = 0
		node.children[an.numChildren] = nil
	}

	// we have to return the number of children for the current node(node4) as
	// `node.numChildren` plus one if null node is not nil.
	// `Shrink` method can be invoked after this method,
	// `Shrink` can convert this node into a leaf node type.
	// For all higher nodes(16/48/256) we simply copy null node to a smaller node
	// see deleteChild() and shrink() methods for implementation details
	numChildren := an.numChildren
	if node.children[an.maxChildren()] != nil {
		numChildren++
	}

	return numChildren
}

func (an *artNode) _deleteChild16(c byte, valid bool) uint16 {
	node := an.node16()
	if !valid {
		node.children[an.maxChildren()] = nil
	} else if idx := an.index(c); idx >= 0 {
		an.numChildren--
		node.keys[idx] = 0
		node.present[idx] = 0
		node.children[idx] = nil

		for i := uint16(idx); i <= an.numChildren && i+1 < node16Max; i++ {
			node.keys[i] = node.keys[i+1]
			node.present[i] = node.present[i+1]
			node.children[i] = node.children[i+1]
		}

		node.keys[an.numChildren] = 0
		node.present[an.numChildren] = 0
		node.children[an.numChildren] = nil
	}

	return an.numChildren
}

func (an *artNode) _deleteChild48(c byte, valid bool) uint16 {
	node := an.node48()
	if !valid {
		node.children[an.maxChildren()] = nil
	} else if idx := an.index(c); idx >= 0 && node.children[idx] != nil {
		node.children[idx] = nil
		node.keys[c] = 0
		node.present[c] = 0
		an.numChildren--
	}

	return an.numChildren
}

func (an *artNode) _deleteChild256(c byte, valid bool) uint16 {
	node := an.node256()
	if !valid {
		node.children[an.maxChildren()] = nil
		return an.numChildren
	} else if idx := an.index(c); node.children[idx] != nil {
		node.children[idx] = nil
		an.numChildren--
	}

	return an.numChildren
}

func (an *artNode) deleteChild(c byte, valid bool) bool {
	deleted := false
	var numChildren uint16
	switch an.kind {
	case Node4:
		numChildren = an._deleteChild4(c, valid)
		deleted = true

	case Node16:
		numChildren = an._deleteChild16(c, valid)
		deleted = true

	case Node48:
		numChildren = an._deleteChild48(c, valid)
		deleted = true

	case Node256:
		numChildren = an._deleteChild256(c, valid)
		deleted = true
	}

	if deleted && numChildren < an.shrinkThreshold() {
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

	d := an
	s := src

	d.numChildren = s.numChildren
	d.prefixLen = s.prefixLen

	for i, limit := uint32(0), min(s.prefixLen, MaxPrefixLen); i < limit; i++ {
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
		d.children[node.maxChildren()] = s.children[an.maxChildren()]

		for i := uint16(0); i < an.numChildren; i++ {
			if s.present[i] != 0 {
				d.keys[i] = s.keys[i]
				d.present[i] = s.present[i]
				d.children[i] = s.children[i]
			}
		}

		return node

	case Node16:
		node := factory.newNode48().copyMeta(an)

		d := node.node48()
		s := an.node16()
		d.children[node.maxChildren()] = s.children[an.maxChildren()]

		var numChildren byte
		for i := uint16(0); i < an.numChildren; i++ {
			if s.present[i] != 0 {
				ch := s.keys[i]
				d.keys[ch] = numChildren
				d.present[ch] = 1
				d.children[numChildren] = s.children[i]
				numChildren++
			}
		}

		return node

	case Node48:
		node := factory.newNode256().copyMeta(an)

		d := node.node256()
		s := an.node48()
		d.children[node.maxChildren()] = s.children[an.maxChildren()]

		for i := 0; i < node256Max; i++ {
			if s.present[i] != 0 {
				d.children[i] = s.children[s.keys[i]]
			}
		}

		return node
	}

	return nil
}

func (an *artNode) shrink() *artNode {
	switch an.kind {
	case Node4:
		node4 := an.node4()
		child := node4.children[0]
		if child == nil {
			child = node4.children[an.maxChildren()]
		}

		if child.isLeaf() {
			return child
		}

		curPrefixLen := an.prefixLen
		if curPrefixLen < MaxPrefixLen {
			an.prefix[curPrefixLen] = node4.keys[0]
			curPrefixLen++
		}

		childNode := child
		if curPrefixLen < MaxPrefixLen {
			childPrefixLen := min(childNode.prefixLen, MaxPrefixLen-curPrefixLen)
			for i := uint32(0); i < childPrefixLen; i++ {
				an.prefix[curPrefixLen+i] = childNode.prefix[i]
			}
			curPrefixLen += childPrefixLen
		}

		for i := uint32(0); i < min(curPrefixLen, MaxPrefixLen); i++ {
			childNode.prefix[i] = an.prefix[i]
		}
		childNode.prefixLen += an.prefixLen + 1

		return child

	case Node16:
		node16 := an.node16()

		newNode := factory.newNode4().copyMeta(an)
		node4 := newNode.node4()
		newNode.numChildren = 0
		for i := 0; i < len(node4.keys); i++ {
			node4.keys[i] = node16.keys[i]
			node4.present[i] = node16.present[i]
			node4.children[i] = node16.children[i]
			newNode.numChildren++
		}

		node4.children[newNode.maxChildren()] = node16.children[an.maxChildren()]

		return newNode

	case Node48:
		node48 := an.node48()

		newNode := factory.newNode16().copyMeta(an)
		node16 := newNode.node16()
		newNode.numChildren = 0
		for i, idx := range node48.keys {
			if node48.present[i] == 0 {
				continue
			}

			if child := node48.children[idx]; child != nil {
				node16.children[newNode.numChildren] = child
				node16.keys[newNode.numChildren] = byte(i)
				node16.present[newNode.numChildren] = 1
				newNode.numChildren++
			}
		}

		node16.children[newNode.maxChildren()] = node48.children[an.maxChildren()]

		return newNode

	case Node256:
		node256 := an.node256()

		newNode := factory.newNode48().copyMeta(an)
		node48 := newNode.node48()
		newNode.numChildren = 0
		for i, child := range node256.children {
			if child != nil {
				node48.children[newNode.numChildren] = child
				node48.keys[byte(i)] = byte(newNode.numChildren)
				node48.present[byte(i)] = 1
				newNode.numChildren++
			}
		}

		node48.children[newNode.maxChildren()] = node256.children[an.maxChildren()]

		return newNode
	}

	return nil
}

// Leaf methods
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

// Base node methods
func (n *artNode) match(key Key, depth uint32) uint32 /* 1st mismatch index*/ {
	idx := uint32(0)
	if len(key)-int(depth) < 0 {
		return idx
	}

	limit := min(min(n.prefixLen, MaxPrefixLen), uint32(len(key))-depth)
	for ; idx < limit; idx++ {
		if n.prefix[idx] != key[idx+depth] {
			return idx
		}
	}

	return idx
}

// Node helpers
func replaceRef(oldNode **artNode, newNode *artNode) {
	*oldNode = newNode
}

func replaceNode(oldNode *artNode, newNode *artNode) {
	*oldNode = *newNode
}
