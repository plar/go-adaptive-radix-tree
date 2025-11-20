package art

// traverseAction is an action to be taken during tree traversal.
type traverseAction int

const (
	traverseStop     traverseAction = iota // traverseStop stops the tree traversal.
	traverseContinue                       // traverseContinue continues the tree traversal.
)

// traverseFunc defines the function for tree traversal.
// It returns the index of the next child node to traverse.
// The second return value indicates whether there are more child nodes to traverse.
type traverseFunc func() (int, bool)

// noopTraverseFunc is a no-op function for tree traversal.
func noopTraverseFunc() (int, bool) {
	return 0, false
}

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

// traverseContext is a context for traversing nodes with 4, 16, or 256 children.
type traverseContext struct {
	numChildren   int
	zeroChildDone bool
	curChildIdx   int
}

// ascTraversal traverses the children in ascending order.
func (ctx *traverseContext) ascTraversal() (int, bool) {
	if !ctx.zeroChildDone {
		ctx.zeroChildDone = true

		return ctx.numChildren, true
	}

	idx := ctx.curChildIdx
	ctx.curChildIdx++

	return idx, idx < ctx.numChildren
}

// descTraversal traverses the children in descending order.
func (ctx *traverseContext) descTraversal() (int, bool) {
	if ctx.curChildIdx >= 0 {
		idx := ctx.curChildIdx
		ctx.curChildIdx--

		return idx, true
	}

	if !ctx.zeroChildDone {
		ctx.zeroChildDone = true

		return ctx.numChildren, true
	}

	return 0, false
}

// newTraverseGenericFunc creates a new traverseFunc for nodes with 4, 16, or 256 children.
// The reverse parameter indicates whether to traverse the children in reverse order.
func newTraverseGenericFunc(numChildren int, reverse bool) traverseFunc {
	ctx := &traverseContext{
		numChildren:   numChildren,
		zeroChildDone: false,
		curChildIdx:   ternary(reverse, numChildren-1, 0),
	}

	return ternary(reverse, ctx.descTraversal, ctx.ascTraversal)
}

// traverse48Context is a context for traversing nodes with 48 children.
type traverse48Context struct {
	curKeyIdx     int
	curKeyCh      byte
	zeroChildDone bool
	n48           *node48
}

// ascTraversal traverses the children in ascending order.
func (ctx *traverse48Context) ascTraversal() (int, bool) {
	if !ctx.zeroChildDone {
		ctx.zeroChildDone = true

		return node48Max, true
	}

	for ; ctx.curKeyIdx < node256Max; ctx.curKeyIdx++ {
		if ctx.n48.hasChild(ctx.curKeyIdx) {
			ctx.curKeyCh = ctx.n48.keys[ctx.curKeyIdx]
			ctx.curKeyIdx++

			return int(ctx.curKeyCh), true
		}
	}

	return 0, false
}

// descTraversal traverses the children in descending order.
func (ctx *traverse48Context) descTraversal() (int, bool) {
	for ; ctx.curKeyIdx > 0; ctx.curKeyIdx-- {
		if ctx.n48.hasChild(ctx.curKeyIdx) {
			ctx.curKeyCh = ctx.n48.keys[ctx.curKeyIdx]
			ctx.curKeyIdx--

			return int(ctx.curKeyCh), true
		}
	}

	if !ctx.zeroChildDone {
		ctx.zeroChildDone = true

		return node48Max, true
	}

	return 0, false
}

// newTraverse48Func creates a new traverseFunc for nodes with 48 children.
// The reverse parameter indicates whether to traverse the children in reverse order.
func newTraverse48Func(n48 *node48, reverse bool) traverseFunc {
	ctx := &traverse48Context{
		curKeyIdx: ternary(reverse, node256Max-1, 0),
		n48:       n48,
	}

	return ternary(reverse, ctx.descTraversal, ctx.ascTraversal)
}

func newTraverseFunc(n *nodeRef, reverse bool) traverseFunc {
	if n == nil {
		return noopTraverseFunc
	}

	switch n.kind { //nolint:exhaustive
	case Node4:
		return newTraverseGenericFunc(node4Max, reverse)
	case Node16:
		return newTraverseGenericFunc(node16Max, reverse)
	case Node48:
		return newTraverse48Func(n.node48(), reverse)
	case Node256:
		return newTraverseGenericFunc(node256Max, reverse)
	default:
		return noopTraverseFunc
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

	nextFn := newTraverseFunc(current, reverse)
	children := toNode(current).allChildren()

	return tr.traverseChildren(nextFn, children, callback, reverse)
}

func (tr *tree) traverseChildren(nextFn traverseFunc, children []*nodeRef, cb Callback, reverse bool) traverseAction {
	for {
		idx, hasMore := nextFn()
		if !hasMore {
			break
		}

		if child := children[idx]; child != nil {
			if tr.forEachRecursively(child, cb, reverse) == traverseStop {
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
