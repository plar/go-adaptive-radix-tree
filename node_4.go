package art

// node4 represents a node with 4 children.
type node4 struct {
	node
	children [node4Max + 1]*nodeRef // pointers to the child nodes, +1 is for the zero byte child
	keys     [node4Max]byte         // keys for the children
	present  [node4Max]byte         // present bits for the keys
}

// minimum returns the minimum leaf node.
func (n *node4) minimum() *leaf {
	return nodeMinimum(n.children[:])
}

// maximum returns the maximum leaf node.
func (n *node4) maximum() *leaf {
	return nodeMaximum(n.children[:n.childrenLen])
}

// index returns the index of the given character.
func (n *node4) index(kc keyChar) int {
	if kc.invalid {
		return node4Max
	}

	return findIndex(n.keys[:n.childrenLen], kc.ch)
}

// childAt returns the child at the given index.
func (n *node4) childAt(idx int) **nodeRef {
	if idx < 0 || idx >= len(n.children) {
		return &nodeNotFound
	}

	return &n.children[idx]
}

func (n *node4) allChildren() []*nodeRef {
	return n.children[:]
}

// hasCapacityForChild returns true if the node has room for more children.
func (n *node4) hasCapacityForChild() bool {
	return n.childrenLen < node4Max
}

// grow converts the node4 into the node16.
func (n *node4) grow() *nodeRef {
	an16 := factory.newNode16()
	n16 := an16.node16()

	copyNode(&n16.node, &n.node)
	n16.children[node16Max] = n.children[node4Max] // copy zero byte child

	for i := 0; i < int(n.childrenLen); i++ {
		// skip if the key is not present
		if n.present[i] == 0 {
			continue
		}

		// copy elements from n4 to n16 to the last position
		n16.insertChildAt(i, n.keys[i], n.children[i])
	}

	return an16
}

// isReadyToShrink returns true if the node is under-utilized and ready to shrink.
func (n *node4) isReadyToShrink() bool {
	// we have to return the number of children for the current node(node4) as
	// `node.numChildren` plus one if zero node is not nil.
	// For all higher nodes(16/48/256) we simply copy zero node to a smaller node
	// see deleteChild() and shrink() methods for implementation details
	numChildren := n.childrenLen
	if n.children[node4Max] != nil {
		numChildren++
	}

	return numChildren < node4Min
}

// shrink converts the node4 into the leaf node or a node with fewer children.
func (n *node4) shrink() *nodeRef {
	// Select the non-nil child node
	var nonNilChild *nodeRef
	if n.children[0] != nil {
		nonNilChild = n.children[0]
	} else {
		nonNilChild = n.children[node4Max]
	}

	// if the only child is a leaf node, return it
	if nonNilChild.isLeaf() {
		return nonNilChild
	}

	// update the prefix of the child node
	n.adjustPrefix(nonNilChild.node())

	return nonNilChild
}

// adjustPrefix handles prefix adjustments for a non-leaf child.
func (n *node4) adjustPrefix(childNode *node) {
	nodePrefLen := int(n.prefixLen)

	// at this point, the node has only one child
	// copy the key part of the current node as prefix
	if nodePrefLen < maxPrefixLen {
		n.prefix[nodePrefLen] = n.keys[0]
		nodePrefLen++
	}

	// copy the part of child prefix that fits into the current node
	if nodePrefLen < maxPrefixLen {
		childPrefLen := minInt(int(childNode.prefixLen), maxPrefixLen-nodePrefLen)
		copy(n.prefix[nodePrefLen:], childNode.prefix[:childPrefLen])
		nodePrefLen += childPrefLen
	}

	// copy the part of the current node prefix that fits into the child node
	prefixLen := minInt(nodePrefLen, maxPrefixLen)
	copy(childNode.prefix[:], n.prefix[:prefixLen])
	childNode.prefixLen += n.prefixLen + 1
}

// addChild adds a new child to the node.
func (n *node4) addChild(kc keyChar, child *nodeRef) {
	pos := n.findInsertPos(kc)
	n.makeRoom(pos)
	n.insertChildAt(pos, kc.ch, child)
}

// find the insert position for the new child.
func (n *node4) findInsertPos(kc keyChar) int {
	if kc.invalid {
		return node4Max
	}

	numChildren := int(n.childrenLen)
	for i := 0; i < numChildren; i++ {
		if n.keys[i] > kc.ch {
			return i
		}
	}

	return numChildren
}

// makeRoom creates space for the new child by shifting the elements to the right.
func (n *node4) makeRoom(pos int) {
	if pos < 0 || pos >= int(n.childrenLen) {
		return
	}

	for i := int(n.childrenLen); i > pos; i-- {
		n.keys[i] = n.keys[i-1]
		n.present[i] = n.present[i-1]
		n.children[i] = n.children[i-1]
	}
}

// insertChildAt inserts the child at the given position.
func (n *node4) insertChildAt(pos int, ch byte, child *nodeRef) {
	if pos == node4Max {
		n.children[pos] = child
	} else {
		n.keys[pos] = ch
		n.present[pos] = 1
		n.children[pos] = child
		n.childrenLen++
	}
}

// deleteChild deletes the child from the node.
func (n *node4) deleteChild(kc keyChar) int {
	if kc.invalid {
		// clear the zero byte child reference
		n.children[node4Max] = nil
	} else if idx := n.index(kc); idx >= 0 {
		n.deleteChildAt(idx)
		n.clearLastElement()
	}

	// we have to return the number of children for the current node(node4) as
	// `n.numChildren` plus one if null node is not nil.
	// `Shrink` method can be invoked after this method,
	// `Shrink` can convert this node into a leaf node type.
	// For all higher nodes(16/48/256) we simply copy null node to a smaller node
	// see deleteChild() and shrink() methods for implementation details
	numChildren := int(n.childrenLen)
	if n.children[node4Max] != nil {
		numChildren++
	}

	return numChildren
}

// deleteChildAt deletes the child at the given index
// by shifting the elements to the left to overwrite deleted child.
func (n *node4) deleteChildAt(idx int) {
	for i := idx; i < int(n.childrenLen) && i+1 < node4Max; i++ {
		n.keys[i] = n.keys[i+1]
		n.present[i] = n.present[i+1]
		n.children[i] = n.children[i+1]
	}

	n.childrenLen--
}

// clearLastElement clears the last element in the node.
func (n *node4) clearLastElement() {
	lastIdx := int(n.childrenLen)
	n.keys[lastIdx] = 0
	n.present[lastIdx] = 0
	n.children[lastIdx] = nil
}
