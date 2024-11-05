package art

// Node with 48 children.
const (
	n48bitShift = 6  // 2^n48bitShift == n48maskLen
	n48maskLen  = 64 // it should be sizeof(node48.present[0])
)

// present48 is a bitfield to store the presence of keys in the node48.
// It is a bitfield of 256 bits, so it is stored in 4 uint64.
type present48 [4]uint64

func (p present48) hasChild(ch int) bool {
	return p[ch>>n48bitShift]&(1<<(ch%n48maskLen)) != 0
}

func (p *present48) setAt(ch int) {
	(*p)[ch>>n48bitShift] |= (1 << (ch % n48maskLen))
}

func (p *present48) clearAt(ch int) {
	(*p)[ch>>n48bitShift] &= ^(1 << (ch % n48maskLen))
}

type node48 struct {
	node
	children [node48Max + 1]*nodeRef // +1 is for the zero byte child
	keys     [node256Max]byte
	present  present48 // need 256 bits for keys
}

// minimum returns the minimum leaf node.
func (n *node48) minimum() *leaf {
	if n.children[node48Max] != nil {
		return n.children[node48Max].minimum()
	}

	idx := 0
	for !n.hasChild(idx) {
		idx++
	}

	if n.children[n.keys[idx]] != nil {
		return n.children[n.keys[idx]].minimum()
	}

	return nil
}

// maximum returns the maximum leaf node.
func (n *node48) maximum() *leaf {
	idx := node256Max - 1
	for !n.hasChild(idx) {
		idx--
	}

	return n.children[n.keys[idx]].maximum()
}

// index returns the index of the child with the given key.
func (n *node48) index(kc keyChar) int {
	if kc.invalid {
		return node48Max
	}

	if n.hasChild(int(kc.ch)) {
		idx := int(n.keys[kc.ch])
		if idx < node48Max && n.children[idx] != nil {
			return idx
		}
	}

	return indexNotFound
}

// childAt returns the child at the given index.
func (n *node48) childAt(idx int) **nodeRef {
	if idx < 0 || idx >= len(n.children) {
		return &nodeNotFound
	}
	return &n.children[idx]
}

func (n *node48) childZero() **nodeRef {
	return &n.children[node48Max]
}

// hasCapacityForChild returns true if the node has room for more children.
func (n *node48) hasCapacityForChild() bool {
	return n.childrenLen < node48Max
}

// grow converts the node to a node256.
func (n *node48) grow() *nodeRef {
	an256 := factory.newNode256()
	n256 := an256.node256()

	copyNode(&n256.node, &n.node)
	n256.children[node256Max] = n.children[node48Max]

	for i := 0; i < node256Max; i++ {
		if n.hasChild(i) {
			n256.addChild(keyChar{ch: byte(i)}, n.children[n.keys[i]])
		}
	}

	return an256
}

// isReadyToShrink returns true if the node can be shrunk to a smaller node type.
func (n *node48) isReadyToShrink() bool {
	return n.childrenLen < node48Min
}

// shrink converts the node to a node16.
func (n *node48) shrink() *nodeRef {
	an16 := factory.newNode16()
	n16 := an16.node16()

	copyNode(&n16.node, &n.node)
	n16.children[node16Max] = n.children[node48Max]

	pos := 0
	for i, idx := range n.keys {
		if !n.hasChild(i) {
			continue // skip if the key is not present
		}

		child := n.children[idx]
		if child == nil {
			continue // skip if the child is nil
		}

		// copy elements from n48 to n16 to the last position
		n16.insertChildAt(pos, byte(i), child)
		pos++
	}

	return an16
}

func (n *node48) hasChild(idx int) bool {
	return n.present.hasChild(idx)
}

// addChild adds a new child to the node.
func (n *node48) addChild(kc keyChar, child *nodeRef) {
	pos := n.findInsertPos(kc)
	n.insertChildAt(pos, kc.ch, child)
}

// find the insert position for the new child
func (n *node48) findInsertPos(kc keyChar) int {
	if kc.invalid {
		return node48Max
	}

	var i int
	for i < node48Max && n.children[i] != nil {
		i++
	}
	return i
}

// insertChildAt inserts a child at the given position.
func (n *node48) insertChildAt(pos int, ch byte, child *nodeRef) {
	if pos == node48Max {
		n.children[node48Max] = child
		return
	}
	n.keys[ch] = byte(pos)
	n.present.setAt(int(ch))
	n.children[pos] = child
	n.childrenLen++
}

// deleteChild removes the child with the given key.
func (n *node48) deleteChild(kc keyChar) int {
	if kc.invalid {
		// clear the zero byte child reference
		n.children[node48Max] = nil
	} else if idx := n.index(kc); idx >= 0 && n.children[idx] != nil {
		// clear the child at the given index
		n.children[idx] = nil
		n.keys[kc.ch] = 0
		n.present.clearAt(int(kc.ch))
		n.childrenLen--
	}

	return int(n.childrenLen)
}
