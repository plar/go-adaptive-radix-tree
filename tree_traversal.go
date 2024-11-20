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

// traverseOpts defines the options for tree traversal.
type traverseOpts int

func (opts traverseOpts) hasLeaf() bool {
	return opts&TraverseLeaf == TraverseLeaf
}

func (opts traverseOpts) hasNode() bool {
	return opts&TraverseNode == TraverseNode
}

func (opts traverseOpts) hasAll() bool {
	return opts&TraverseAll == TraverseAll
}

func (opts traverseOpts) hasReverse() bool {
	return opts&TraverseReverse == TraverseReverse
}

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
func traversalContext4_16_256(childZeroIdx int, _ bool) *traverse4_16_256 {
	return &traverse4_16_256{
		numChildren:     childZeroIdx,
		childZeroActive: true,
		childCurIdx:     0,
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
func traversalContext48(n48 *node48, _ bool) *traverse48 {
	return &traverse48{
		childZeroIdx:    node48Max,
		childZeroActive: true,
		keyCurIdx:       0,
		n48:             n48,
	}
}

func newTraversalContext(n *nodeRef, reverse bool) traverseContext {
	if n == nil {
		return noopTraverseCtx
	}

	switch n.kind { //nolint:exhaustive
	case Node4:
		return traversalContext4_16_256(node4Max, reverse)
	case Node16:
		return traversalContext4_16_256(node16Max, reverse)
	case Node48:
		return traversalContext48(n.node48(), reverse)
	case Node256:
		return traversalContext4_16_256(node256Max, reverse)
	default:
		return noopTraverseCtx
	}
}

func mergeOptions(options ...int) int {
	opts := 0
	for _, opt := range options {
		opts |= opt
	}

	return opts
}

func traverseOptions(options ...int) traverseOpts {
	opts := mergeOptions(options...)

	typeOpts := opts & TraverseAll
	if typeOpts == 0 {
		typeOpts = TraverseLeaf // By default filter only leafs
	}

	orderOpts := opts & TraverseReverse

	return traverseOpts(typeOpts | orderOpts)
}

func traverseFilter(opts traverseOpts, callback Callback) Callback {
	if opts.hasAll() {
		return callback
	}

	return func(node Node) bool {
		if opts.hasLeaf() && node.Kind() == Leaf {
			return callback(node)
		}

		if opts.hasNode() && node.Kind() != Leaf {
			return callback(node)
		}

		return true
	}
}

func (tr *tree) forEachRecursively(current *nodeRef, callback Callback, reverse bool) traverseAction {
	if current == nil {
		return traverseContinue
	}

	if !callback(current) {
		return traverseStop
	}

	ctx := newTraversalContext(current, reverse)
	children := toNode(current).allChildren()

	return tr.traverseNode(ctx, children, callback, reverse)
}

func (tr *tree) traverseNode(ctx traverseContext, children []*nodeRef, callback Callback, reverse bool) traverseAction {
	for {
		idx, ok := ctx.nextChildIdx()
		if !ok {
			break
		}

		if child := children[idx]; child != nil {
			if tr.forEachRecursively(child, callback, reverse) == traverseStop {
				return traverseStop
			}
		}
	}

	return traverseContinue
}

func (tr *tree) forEachPrefix(key Key, callback Callback, opts int) traverseAction {
	opts &= (TraverseLeaf | TraverseReverse) // keep only leaf and reverse options

	tr.ForEach(func(n Node) bool {
		current, ok := n.(*nodeRef)
		if !ok {
			return false
		}

		if leaf := current.leaf(); leaf.prefixMatch(key) {
			return callback(current)
		}

		return true
	}, opts)

	return traverseContinue
}
