package art

// Node with 256 children.
type node256 struct {
	node
	children [node256Max]*nodeRef
}

// minimum returns the minimum leaf node.
func (n *node256) minimum() *leaf {
	return nodeMinimum(n.zeroChild, n.children[:])
}

// maximum returns the maximum leaf node.
func (n *node256) maximum() *leaf {
	return nodeMaximum(n.children[:node256Max])
}

// index returns the index of the child with the given key.
func (n *node256) index(ch byte) int {
	return int(ch)
}

// childAt returns the child at the given index.
func (n *node256) childAt(idx int) **nodeRef {
	return &n.children[idx]
}

// addChild adds a new child to the node.
func (n *node256) addChild(ch byte, valid bool, child *nodeRef) {
	if !valid { // handle zero byte in the key
		n.zeroChild = child
		return
	}

	// insert new child
	n.children[ch] = child
	n.childrenLen++
}

// canAddChild for node256 always returns true.
func (n *node256) canAddChild() bool {
	return true
}

// grow is implemeneted to satisfy the noder interface
// but it does not do anything for node256.
func (n *node256) grow() *nodeRef {
	return nil
}

// canShrinkNode returns true if the node can be shrinked.
func (n *node256) canShrinkNode() bool {
	return n.childrenLen < node256Min
}

// shrink shrinks the node to a smaller type.
func (n *node256) shrink() *nodeRef {
	an48 := factory.newNode48()
	n48 := an48.node48()

	copyNode(&n48.node, &n.node)

	pos := 0
	for ch, child := range n.children {
		if child == nil {
			continue // skip if the child is nil
		}

		// copy elements from n256 to n48 to the last position
		n48.insertChildAt(pos, byte(ch), child)
		pos++
	}

	return an48
}

// deleteChild removes the child with the given key.
func (n *node256) deleteChild(ch byte, valid bool) int {
	if !valid {
		// clear the zero byte child reference
		n.zeroChild = nil
	} else if idx := n.index(ch); n.children[idx] != nil {
		// clear the child at the given index
		n.children[idx] = nil
		n.childrenLen--
	}

	return int(n.childrenLen)
}
