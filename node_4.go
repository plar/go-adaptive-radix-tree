package art

// node4 represents a node with 4 children.
type node4 struct {
	node
	children [node4Max]*nodeRef // pointers to the child nodes
	keys     [node4Max]byte     // keys for the children
	present  [node4Max]byte     // present bits for the keys
}

// minimum returns the minimum leaf node.
func (n *node4) minimum() *leaf {
	return nodeMinimum(n.zeroChild, n.children[:])
}

// maximum returns the maximum leaf node.
func (n *node4) maximum() *leaf {
	return nodeMaximum(n.children[:n.childrenLen])
}

// index returns the index of the given character.
func (n *node4) index(ch byte) int {
	return findIndex(n.keys[:n.childrenLen], ch)
}

// childAt returns the child at the given index.
func (n *node4) childAt(idx int) **nodeRef {
	return &n.children[idx]
}

// canAddChild returns true if the node has room for more children.
func (n *node4) canAddChild() bool {
	return n.childrenLen < node4Max
}

// grow converts the node4 into the node16.
func (n *node4) grow() *nodeRef {
	an16 := factory.newNode16()
	n16 := an16.node16()
	copyNode(&n16.node, &n.node)

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

// canShrinkNode returns true if the node can be shrunk.
func (n *node4) canShrinkNode() bool {
	// we have to return the number of children for the current node(node4) as
	// `node.numChildren` plus one if null node is not nil.
	// `Shrink` method can be invoked after this method,
	// `Shrink` can convert this node into a leaf node type.
	// For all higher nodes(16/48/256) we simply copy null node to a smaller node
	// see deleteChild() and shrink() methods for implementation details
	numChildren := n.childrenLen
	if n.zeroChild != nil {
		numChildren++
	}
	return numChildren < node4Min
}

// shrink converts the node4 into the leaf node or node with fewer children.
func (n *node4) shrink() *nodeRef {
	child := n.children[0]
	if child == nil {
		child = n.zeroChild
	}

	if child.isLeaf() {
		// node has only one child and it is a leaf node
		// we can convert this node into a leaf node
		return child
	}

	curPrefixLen := int(n.prefixLen)
	if curPrefixLen < MaxPrefixLen {
		n.prefix[curPrefixLen] = n.keys[0]
		curPrefixLen++
	}

	childNode := child.node()
	if curPrefixLen < MaxPrefixLen {
		childPrefixLen := min(int(childNode.prefixLen), MaxPrefixLen-curPrefixLen)
		copy(n.prefix[curPrefixLen:], childNode.prefix[:childPrefixLen])
		curPrefixLen += childPrefixLen
	}

	prefixLen := min(curPrefixLen, MaxPrefixLen)
	copy(childNode.prefix[:], n.prefix[:prefixLen])
	childNode.prefixLen += n.prefixLen + 1

	return child
}

// addChild adds a new child to the node.
func (n *node4) addChild(ch byte, valid bool, child *nodeRef) {
	if !valid { // handle zero byte in the key
		n.zeroChild = child
		return
	}

	pos := n.findInsertPos(ch)
	n.makeRoom(pos)
	n.insertChildAt(pos, ch, child)
}

// find the insert position for the new child
func (n *node4) findInsertPos(ch byte) int {
	var insertPos int
	for ; insertPos < int(n.childrenLen); insertPos++ {
		if n.keys[insertPos] > ch {
			break
		}
	}
	return insertPos
}

// makeRoom creates space for the new child by shifting the elements to the right.
func (n *node4) makeRoom(pos int) {
	for i := int(n.childrenLen); i > pos; i-- {
		n.keys[i] = n.keys[i-1]
		n.present[i] = n.present[i-1]
		n.children[i] = n.children[i-1]
	}
}

// insertChildAt inserts the child at the given position.
func (n *node4) insertChildAt(pos int, ch byte, child *nodeRef) {
	n.keys[pos] = ch
	n.present[pos] = 1
	n.children[pos] = child
	n.childrenLen++
}

// deleteChild deletes the child from the node.
func (n *node4) deleteChild(ch byte, valid bool) int {
	if !valid {
		// clear the zero byte child reference
		n.zeroChild = nil
	} else if idx := n.index(ch); idx >= 0 {
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
	if n.zeroChild != nil {
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
	n.present[lastIdx] = 1
	n.children[lastIdx] = nil
}
