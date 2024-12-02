package art

// Node with 256 children.
type node256 struct {
	node
	children [node256Max + 1]*nodeRef // +1 is for the zero byte child
}

// minimum returns the minimum leaf node.
func (n *node256) minimum() *leaf {
	return nodeMinimum(n.children[:])
}

// maximum returns the maximum leaf node.
func (n *node256) maximum() *leaf {
	return nodeMaximum(n.children[:node256Max])
}

// index returns the index of the child with the given key.
func (n *node256) index(kc keyChar) int {
	if kc.invalid { // handle zero byte in the key
		return node256Max
	}

	return int(kc.ch)
}

// childAt returns the child at the given index.
func (n *node256) childAt(idx int) **nodeRef {
	if idx < 0 || idx >= len(n.children) {
		return &nodeNotFound
	}

	return &n.children[idx]
}

func (n *node256) allChildren() []*nodeRef {
	return n.children[:]
}

// addChild adds a new child to the node.
func (n *node256) addChild(kc keyChar, child *nodeRef) {
	if kc.invalid {
		// handle zero byte in the key
		n.children[node256Max] = child
	} else {
		// insert new child
		n.children[kc.ch] = child
		n.childrenLen++
	}
}

// hasCapacityForChild for node256 always returns true.
func (n *node256) hasCapacityForChild() bool {
	return true
}

// grow for node256 always returns nil,
// because node256 has the maximum capacity.
func (n *node256) grow() *nodeRef {
	return nil
}

// isReadyToShrink returns true if the node can be shrunk.
func (n *node256) isReadyToShrink() bool {
	return n.childrenLen < node256Min
}

// shrink shrinks the node to a smaller type.
func (n *node256) shrink() *nodeRef {
	an48 := factory.newNode48()
	n48 := an48.node48()

	copyNode(&n48.node, &n.node)
	n48.children[node48Min] = n.children[node256Max] // copy zero byte child

	for numChildren, i := 0, 0; i < node256Max; i++ {
		if n.children[i] == nil {
			continue // skip if the child is nil
		}
		// copy elements from n256 to n48 to the last position
		n48.insertChildAt(numChildren, byte(i), n.children[i])

		numChildren++
	}

	return an48
}

// deleteChild removes the child with the given key.
func (n *node256) deleteChild(kc keyChar) int {
	if kc.invalid {
		// clear the zero byte child reference
		n.children[node256Max] = nil
	} else if idx := n.index(kc); n.children[idx] != nil {
		// clear the child at the given index
		n.children[idx] = nil
		n.childrenLen--
	}

	return int(n.childrenLen)
}
