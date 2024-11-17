package art

// traverseAction is an action to be taken during tree traversal.
type traverseAction int

const (
	traverseStop     traverseAction = iota // traverseStop stops the tree traversal.
	traverseContinue                       // traverseContinue continues the tree traversal.
)

// traverseContext defines the interface for tree traversal context.
type traverseContext interface {
	// nextChildIdx returns the index of the next child node to traverse.
	// The second return value indicates whether there are more child nodes to traverse.
	nextChildIdx() (int, bool)
}

type noopTraverseContext struct{}

func (noopTraverseContext) nextChildIdx() (int, bool) {
	return 0, false
}

var noopTraverseCtx = &noopTraverseContext{} //nolint:gochecknoglobals

// traverse4_16_256 is a context for traversing nodes with 4, 16, or 256 children.
type traverse4_16_256 struct {
	numChildren     int
	childZeroActive bool
	childCurIdx     int
}

// nextChildIdx returns the index of the next child node to traverse.
func (tc *traverse4_16_256) nextChildIdx() (int, bool) {
	if tc.childZeroActive {
		tc.childZeroActive = false

		return tc.numChildren, true
	}

	idx := tc.childCurIdx
	tc.childCurIdx++

	return idx, idx < tc.numChildren
}

// traversalContext4_16_256 creates a new context for traversing node4/16/256 children.
func traversalContext4_16_256(childZeroIdx int) *traverse4_16_256 {
	return &traverse4_16_256{
		numChildren:     childZeroIdx,
		childZeroActive: true,
		childCurIdx:     0,
	}
}

// traverseDesc4_16_256 is a context for traversing nodes with 4, 16, or 256 children in descending order.
type traverseDesc4_16_256 struct {
	numChildren     int
	childZeroActive bool
	childCurIdx     int
}

// nextChildIdx returns the index of the next child node to traverse.
func (tc *traverseDesc4_16_256) nextChildIdx() (int, bool) {
	if tc.childZeroActive {
		tc.childZeroActive = false

		return tc.numChildren, true
	}

	idx := tc.childCurIdx
	tc.childCurIdx--

	return idx, idx >= 0
}

// traversalContext4_16_256 creates a new context for traversing node4/16/256 children.
func traversalDescContext4_16_256(childZeroIdx int) *traverseDesc4_16_256 {
	return &traverseDesc4_16_256{
		numChildren:     childZeroIdx,
		childZeroActive: true,
		childCurIdx:     childZeroIdx,
	}
}

// traverse48 is a context for traversing nodes with 48 children.
type traverse48 struct {
	childZeroIdx    int
	keyCurIdx       int
	keyCurCh        byte
	childZeroActive bool
	n48             *node48
}

// nextChildIdx returns the index of the next child node to traverse.
func (tc *traverse48) nextChildIdx() (int, bool) {
	if tc.childZeroActive {
		tc.childZeroActive = false
		idx := tc.childZeroIdx

		return idx, true
	}

	for {
		if tc.keyCurIdx >= node256Max {
			break
		}

		if tc.n48.hasChild(tc.keyCurIdx) {
			tc.keyCurCh = tc.n48.keys[tc.keyCurIdx]
			tc.keyCurIdx++

			return int(tc.keyCurCh), true
		}

		tc.keyCurIdx++
	}

	return 0, false
}

// traversalContext48 creates a new context for traversing node48 children.
func traversalContext48(n48 *node48) *traverse48 {
	return &traverse48{
		childZeroIdx:    node48Max,
		childZeroActive: true,
		keyCurIdx:       0,
		n48:             n48,
	}
}

// traverseDesc48 is a context for traversing nodes with 48 children in descending order.
type traverseDesc48 struct {
	childZeroIdx    int
	keyCurIdx       int
	keyCurCh        byte
	childZeroActive bool
	n48             *node48
}

// nextChildIdx returns the index of the next child node to traverse.
func (tc *traverseDesc48) nextChildIdx() (int, bool) {
	if tc.childZeroActive {
		tc.childZeroActive = false
		idx := tc.childZeroIdx

		return idx, true
	}

	for {
		if tc.keyCurIdx < 0 {
			break
		}

		if tc.n48.hasChild(tc.keyCurIdx) {
			tc.keyCurCh = tc.n48.keys[tc.keyCurIdx]
			tc.keyCurIdx--

			return int(tc.keyCurCh), true
		}

		tc.keyCurIdx--
	}

	return 0, false
}

// traversalDescContext48 creates a new context for traversing node48 children in descending order.
func traversalDescContext48(n48 *node48) *traverseDesc48 {
	return &traverseDesc48{
		childZeroIdx:    node48Max,
		childZeroActive: true,
		keyCurIdx:       node256Max - 1,
		n48:             n48,
	}
}

func newTraversalContext(n *nodeRef) traverseContext {
	if n == nil {
		return noopTraverseCtx
	}

	switch n.kind { //nolint:exhaustive
	case Node4:
		return traversalContext4_16_256(node4Max)
	case Node16:
		return traversalContext4_16_256(node16Max)
	case Node48:
		return traversalContext48(n.node48())
	case Node256:
		return traversalContext4_16_256(node256Max)
	default:
		return noopTraverseCtx
	}
}

func newTraversalDescContext(n *nodeRef) traverseContext {
	if n == nil {
		return noopTraverseCtx
	}

	switch n.kind { //nolint:exhaustive
	case Node4:
		return traversalDescContext4_16_256(node4Max)
	case Node16:
		return traversalDescContext4_16_256(node16Max)
	case Node48:
		return traversalDescContext48(n.node48())
	case Node256:
		return traversalDescContext4_16_256(node256Max)
	default:
		return noopTraverseCtx
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
		}

		if options&TraverseNode == TraverseNode && node.Kind() != Leaf {
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

	ctx := newTraversalContext(current)
	children := toNode(current).allChildren()

	return tr.traverseNode(ctx, children, callback)
}

func (tr *tree) forEachDescRecursively(current *nodeRef, callback Callback) traverseAction {
	if current == nil {
		return traverseContinue
	}

	ctx := newTraversalDescContext(current)
	children := toNode(current).allChildren()

	if tr.traverseDescNode(ctx, children, callback) == traverseStop {
		return traverseStop
	}

	if !callback(current) {
		return traverseStop
	}
	return traverseContinue

}

func (tr *tree) traverseNode(ctx traverseContext, children []*nodeRef, callback Callback) traverseAction {
	for {
		idx, ok := ctx.nextChildIdx()
		if !ok {
			break
		}

		if child := children[idx]; child != nil {
			if tr.forEachRecursively(child, callback) == traverseStop {
				return traverseStop
			}
		}
	}

	return traverseContinue
}

func (tr *tree) traverseDescNode(ctx traverseContext, children []*nodeRef, callback Callback) traverseAction {
	for {
		idx, ok := ctx.nextChildIdx()
		if !ok {
			break
		}

		if child := children[idx]; child != nil {
			if tr.forEachDescRecursively(child, callback) == traverseStop {
				return traverseStop
			}
		}
	}

	return traverseContinue
}

func (tr *tree) forEachPrefix(key Key, callback Callback) traverseAction {
	tr.ForEach(func(n Node) bool {
		current, ok := n.(*nodeRef)
		if !ok {
			return false
		}

		if leaf := current.leaf(); leaf.prefixMatch(key) {
			return callback(current)
		}

		return true
	}, TraverseLeaf)

	return traverseContinue
}
