package art

// traverseAction is an action to be taken during tree traversal.
type traverseAction int

const (
	traverseStop     traverseAction = iota // traverseStop stops the tree traversal.
	traverseContinue                       // traverseContinue continues the tree traversal.
)

// nullIdx is a special index value to indicate that there is no child node.
const nullIdx = -1

// iteratorLevel is a level of the iterator.
type iteratorLevel struct {
	node     *nodeRef
	childIdx int
}

// iterator is an iterator struct for the tree.
type iterator struct {
	version    int // tree version to detect concurrent modifications
	tree       *tree
	nextNode   *nodeRef
	depthLevel int
	depth      []*iteratorLevel
}

// bufferedIterator is a buffered iterator struct for the tree.
// It is used to implement the HasNext and Next methods.
type bufferedIterator struct {
	options  int
	nextNode Node
	err      error
	it       *iterator
}

// newTreeIterator creates a new tree iterator.
func newTreeIterator(tr *tree) *iterator {
	return &iterator{
		version:    tr.version,
		tree:       tr,
		nextNode:   tr.root,
		depthLevel: 0,
		depth: []*iteratorLevel{
			{
				node:     tr.root,
				childIdx: nullIdx,
			},
		},
	}
}

func traverseOptions(options ...int) int {
	opts := 0
	for _, opt := range options {
		opts |= opt
	}

	opts &= TraverseAll
	if opts == 0 {
		opts = TraverseLeaf // By default filter only leafs
	}

	return opts
}

func traverseFilter(options int, callback Callback) Callback {
	if options == TraverseAll {
		return callback
	}

	return func(node Node) bool {
		if options&TraverseLeaf == TraverseLeaf && node.Kind() == Leaf {
			return callback(node)
		} else if options&TraverseNode == TraverseNode && node.Kind() != Leaf {
			return callback(node)
		}
		return true
	}
}

func (tr *tree) forEachRecursively(current *nodeRef, callback Callback) traverseAction {
	if current == nil {
		return traverseContinue
	}

	if !callback(current) {
		return traverseStop
	}

	switch current.kind {
	case Node4:
		return tr.forEachChildren(current.node4().children[node4Max], current.node4().children[:node4Max], callback)

	case Node16:
		return tr.forEachChildren(current.node16().children[node16Max], current.node16().children[:node16Max], callback)

	case Node48:
		n48 := current.node48()
		if child := n48.children[node48Max]; child != nil {
			if tr.forEachRecursively(child, callback) == traverseStop {
				return traverseStop
			}
		}

		for idx, ch := range n48.keys {
			if !n48.hasChild(idx) {
				continue
			}

			child := n48.children[ch]
			if child != nil {
				if tr.forEachRecursively(child, callback) == traverseStop {
					return traverseStop
				}
			}
		}

	case Node256:
		return tr.forEachChildren(current.node256().children[node256Max], current.node256().children[:node256Max], callback)
	}

	return traverseContinue
}

func (tr *tree) forEachChildren(nullChild *nodeRef, children []*nodeRef, callback Callback) traverseAction {
	if nullChild != nil {
		if tr.forEachRecursively(nullChild, callback) == traverseStop {
			return traverseStop
		}
	}

	for _, child := range children {
		if child != nil && child != nullChild {
			if tr.forEachRecursively(child, callback) == traverseStop {
				return traverseStop
			}
		}
	}

	return traverseContinue
}

func (tr *tree) forEachPrefix(current *nodeRef, key Key, callback Callback) traverseAction {
	if current == nil {
		return traverseContinue
	}

	keyOffset := 0
	for current != nil {
		if current.isLeaf() {
			leaf := current.leaf()
			if leaf.prefixMatch(key) {
				if !callback(current) {
					return traverseStop
				}
			}
			break
		}

		if keyOffset == len(key) {
			leaf := current.minimum()
			if leaf.prefixMatch(key) {
				if tr.forEachRecursively(current, callback) == traverseStop {
					return traverseStop
				}
			}
			break
		}

		node := current.node()
		if node.prefixLen > 0 {
			prefixLen := current.matchDeep(key, keyOffset)
			if prefixLen > int(node.prefixLen) {
				prefixLen = int(node.prefixLen)
			}

			if prefixLen == 0 {
				break
			} else if keyOffset+prefixLen == len(key) {
				return tr.forEachRecursively(current, callback)

			}
			keyOffset += int(node.prefixLen)
		}

		// Find a child to recursive to
		next := current.findChildByKey(key, int(keyOffset))
		if *next == nil {
			break
		}
		current = *next
		keyOffset++
	}

	return traverseContinue
}

func (ti *iterator) checkConcurrentModification() error {
	if ti.version == ti.tree.version {
		return nil
	}

	return ErrConcurrentModification
}

func (ti *iterator) HasNext() bool {
	return ti != nil && ti.nextNode != nil
}

func (ti *iterator) Next() (Node, error) {
	if !ti.HasNext() {
		return nil, ErrNoMoreNodes
	}

	err := ti.checkConcurrentModification()
	if err != nil {
		return nil, err
	}

	cur := ti.nextNode
	ti.next()

	return cur, nil
}

func nextChild(childIdx int, nullChild *nodeRef, children []*nodeRef) ( /*nextChildIdx*/ int /*nextNode*/, *nodeRef) {
	if childIdx == nullIdx {
		if nullChild != nil {
			return 0, nullChild
		}

		childIdx = 0
	}

	for i := childIdx; i < len(children); i++ {
		child := children[i]
		if child != nil && child != nullChild {
			return i + 1, child
		}
	}

	return 0, nil
}

func nextChild48(childIdx int, node *node48) ( /*nextChildIdx*/ int /*nextNode*/, *nodeRef) {
	nullChild := node.children[node48Max]

	if childIdx == nullIdx {
		if nullChild != nil {
			return 0, nullChild
		}
		childIdx = 0
	}

	for i := childIdx; i < len(node.keys); i++ {
		if !node.hasChild(i) {
			continue
		}

		child := node.children[node.keys[i]]
		if child != nil && child != nullChild {
			return i + 1, child
		}
	}

	return 0, nil
}

func (ti *iterator) next() {
	for {
		var nextNode *nodeRef
		nextChildIdx := nullIdx

		curNode := ti.depth[ti.depthLevel].node
		curChildIdx := ti.depth[ti.depthLevel].childIdx

		switch curNode.kind {
		case Node4:
			nextChildIdx, nextNode = nextChild(curChildIdx, curNode.node4().children[node4Max], curNode.node4().children[:node4Max])

		case Node16:
			nextChildIdx, nextNode = nextChild(curChildIdx, curNode.node16().children[node16Max], curNode.node16().children[:node16Max])

		case Node48:
			n48 := curNode.node48()
			nextChildIdx, nextNode = nextChild48(curChildIdx, n48)

		case Node256:
			nextChildIdx, nextNode = nextChild(curChildIdx, curNode.node256().children[node256Max], curNode.node256().children[:node256Max])
		}

		if nextNode == nil {
			if ti.depthLevel == 0 { // iterator is done
				ti.nextNode = nil
				return
			}
			ti.depthLevel-- // return to previous level

		} else {
			// star from the next when we come back from the child node
			ti.depth[ti.depthLevel].childIdx = nextChildIdx
			ti.nextNode = nextNode

			// make sure that we have enough space for levels
			if ti.depthLevel+1 >= cap(ti.depth) {
				newDepthLevel := make([]*iteratorLevel, ti.depthLevel+2)
				copy(newDepthLevel, ti.depth)
				ti.depth = newDepthLevel
			}

			ti.depthLevel++
			ti.depth[ti.depthLevel] = &iteratorLevel{nextNode, nullIdx}
			return
		}
	}
}

func (bti *bufferedIterator) HasNext() bool {
	for bti.it.HasNext() {
		bti.nextNode, bti.err = bti.it.Next()
		if bti.err != nil {
			return true
		}
		if bti.options&TraverseLeaf == TraverseLeaf && bti.nextNode.Kind() == Leaf {
			return true
		} else if bti.options&TraverseNode == TraverseNode && bti.nextNode.Kind() != Leaf {
			return true
		}
	}
	bti.nextNode = nil
	bti.err = nil

	return false
}

func (bti *bufferedIterator) Next() (Node, error) {
	return bti.nextNode, bti.err
}
