package art

// present16 is a bitfield to store the presence of keys in the node16.
// node16 needs 16 bits to store the presence of keys.
type present16 uint16

func (p present16) hasChild(idx int) bool {
	return p&(1<<idx) != 0
}

func (p *present16) setAt(idx int) {
	*p |= 1 << idx
}

func (p *present16) clearAt(idx int) {
	*p &= ^(1 << idx)
}

func (p *present16) shiftRight(idx int) {
	p.clearAt(idx)
	*p |= ((*p & (1 << (idx - 1))) << 1)
}

func (p *present16) shiftLeft(idx int) {
	p.clearAt(idx)
	*p |= ((*p & (1 << (idx + 1))) >> 1)
}

// node16 represents a node with 16 children.
type node16 struct {
	node
	children [node16Max + 1]*nodeRef // +1 is for the zero byte child
	keys     [node16Max]byte
	present  present16
}

// minimum returns the minimum leaf node.
func (n *node16) minimum() *leaf {
	return nodeMinimum(n.children[:])
}

// maximum returns the maximum leaf node.
func (n *node16) maximum() *leaf {
	return nodeMaximum(n.children[:n.childrenLen])
}

// index returns the child index for the given key.
func (n *node16) index(kc keyChar) int {
	if kc.invalid {
		return node16Max
	}

	return findIndex(n.keys[:n.childrenLen], kc.ch)
}

// childAt returns the child at the given index.
func (n *node16) childAt(idx int) **nodeRef {
	if idx < 0 || idx >= len(n.children) {
		return &nodeNotFound
	}

	return &n.children[idx]
}

func (n *node16) childZero() **nodeRef {
	return &n.children[node16Max]
}

// hasCapacityForChild returns true if the node has room for more children.
func (n *node16) hasCapacityForChild() bool {
	return n.childrenLen < node16Max
}

// grow converts the node to a node48.
func (n *node16) grow() *nodeRef {
	an48 := factory.newNode48()
	n48 := an48.node48()

	copyNode(&n48.node, &n.node)
	n48.children[node48Max] = n.children[node16Max]

	for numChildren, i := 0, 0; i < int(n.childrenLen); i++ {
		if !n.hasChild(i) {
			continue // skip if the key is not present
		}

		n48.insertChildAt(numChildren, n.keys[i], n.children[i])

		numChildren++
	}

	return an48
}

// caShrinkNode returns true if the node can be shriken.
func (n *node16) isReadyToShrink() bool {
	return n.childrenLen < node16Min
}

// shrink converts the node16 into the node4.
func (n *node16) shrink() *nodeRef {
	an4 := factory.newNode4()
	n4 := an4.node4()

	copyNode(&n4.node, &n.node)
	n4.children[node4Max] = n.children[node16Max]

	for i := 0; i < node4Max; i++ {
		n4.keys[i] = n.keys[i]

		if n.hasChild(i) {
			n4.present[i] = 1
		}

		n4.children[i] = n.children[i]
		if n4.children[i] != nil {
			n4.childrenLen++
		}
	}

	return an4
}

func (n *node16) hasChild(idx int) bool {
	return n.present.hasChild(idx)
}

// addChild adds a new child to the node.
func (n *node16) addChild(kc keyChar, child *nodeRef) {
	pos := n.findInsertPos(kc)
	n.makeRoom(pos)
	n.insertChildAt(pos, kc.ch, child)
}

// find the insert position for the new child.
func (n *node16) findInsertPos(kc keyChar) int {
	if kc.invalid {
		return node16Max
	}

	for i := 0; i < int(n.childrenLen); i++ {
		if n.keys[i] > kc.ch {
			return i
		}
	}

	return int(n.childrenLen)
}

// makeRoom makes room for a new child at the given position.
func (n *node16) makeRoom(pos int) {
	if pos < 0 || pos >= int(n.childrenLen) {
		return
	}

	// Shift keys and children to the right starting from the position
	copy(n.keys[pos+1:], n.keys[pos:int(n.childrenLen)])
	copy(n.children[pos+1:], n.children[pos:int(n.childrenLen)])

	for i := int(n.childrenLen); i > pos; i-- {
		n.present.shiftRight(i)
	}
}

// insertChildAt inserts a new child at the given position.
func (n *node16) insertChildAt(pos int, ch byte, child *nodeRef) {
	if pos < 0 || pos > node16Max {
		return
	}

	if pos == node16Max {
		n.children[pos] = child
	} else {
		n.keys[pos] = ch
		n.present.setAt(pos)
		n.children[pos] = child
		n.childrenLen++
	}
}

// deleChild removes a child from the node.
func (n *node16) deleteChild(kc keyChar) int {
	if kc.invalid {
		// clear the zero byte child reference
		n.children[node16Max] = nil
	} else if idx := n.index(kc); idx >= 0 {
		n.deleteChildAt(idx)
		n.clearLastElement()
	}

	return int(n.childrenLen)
}

// deleteChildAt removes a child at the given position.
func (n *node16) deleteChildAt(idx int) {
	childrenLen := int(n.childrenLen)
	if idx >= childrenLen {
		return
	}

	// Shift keys and children to the left, overwriting the deleted index
	copy(n.keys[idx:], n.keys[idx+1:childrenLen])
	copy(n.children[idx:], n.children[idx+1:childrenLen])

	// shift elements to the left to fill the gap
	for i := idx; i < childrenLen && i+1 < node16Max; i++ {
		n.present.shiftLeft(i)
	}

	n.childrenLen--
}

// clearLastElement clears the last element in the node.
func (n *node16) clearLastElement() {
	lastIdx := int(n.childrenLen)
	n.keys[lastIdx] = 0
	n.present.clearAt(lastIdx)
	n.children[lastIdx] = nil
}
