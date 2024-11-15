package art

// state represents the iteration state during tree traversal.
type state struct {
	items []*iteratorContext
}

func (s *state) push(ctx *iteratorContext) {
	s.items = append(s.items, ctx)
}

func (s *state) current() (*iteratorContext, bool) {
	if len(s.items) == 0 {
		return nil, false
	}

	return s.items[len(s.items)-1], true
}

func (s *state) discard() {
	if len(s.items) == 0 {
		return
	}

	s.items = s.items[:len(s.items)-1]
}

// iteratorContext represents the context of the tree iterator for one node.
type iteratorContext struct {
	tctx     traverseContext
	children []*nodeRef
}

// newIteratorContext creates a new iterator context for the given node.
func newIteratorContext(nr *nodeRef) *iteratorContext {
	return &iteratorContext{
		tctx:     newTraversalContext(nr),
		children: toNode(nr).allChildren(),
	}
}

// next returns the next node reference and a flag indicating if there are more nodes.
func (ic *iteratorContext) next() (*nodeRef, bool) {
	for {
		idx, ok := ic.tctx.nextChildIdx()
		if !ok {
			break
		}

		if child := ic.children[idx]; child != nil {
			return child, true
		}
	}

	return nil, false
}

// iterator is a struct for tree traversal iteration.
type iterator struct {
	version  int
	tree     *tree
	state    *state
	nextNode *nodeRef
}

// assert that iterator implements the Iterator interface.
var _ Iterator = (*iterator)(nil)

func newTreeIterator(tr *tree, opts int) Iterator {
	state := &state{}
	state.push(newIteratorContext(tr.root))

	it := &iterator{
		version:  tr.version,
		tree:     tr,
		nextNode: tr.root,
		state:    state,
	}

	if opts&TraverseAll == TraverseAll {
		return it
	}

	return &bufferedIterator{
		opts: opts,
		it:   it,
	}
}

// hasConcurrentModification checks if the tree has been modified concurrently.
func (it *iterator) hasConcurrentModification() bool {
	return it.version != it.tree.version
}

// HasNext returns true if there are more nodes to iterate.
func (it *iterator) HasNext() bool {
	return it.nextNode != nil
}

// Next returns the next node and an error if any.
// It returns ErrNoMoreNodes if there are no more nodes to iterate.
// It returns ErrConcurrentModification if the tree has been modified concurrently.
func (it *iterator) Next() (Node, error) {
	if !it.HasNext() {
		return nil, ErrNoMoreNodes
	}

	if it.hasConcurrentModification() {
		return nil, ErrConcurrentModification
	}

	current := it.nextNode
	it.next()

	return current, nil
}

// next moves the iterator to the next node.
func (it *iterator) next() {
	for {
		ctx, ok := it.state.current()
		if !ok {
			it.nextNode = nil // no more nodes to iterate

			return
		}

		nextNode, hasMore := ctx.next()
		if hasMore {
			it.nextNode = nextNode
			it.state.push(newIteratorContext(nextNode))

			return
		}

		it.state.discard() // discard the current context as exhausted
	}
}

// BufferedIterator implements HasNext and Next methods for buffered iteration.
// It allows to iterate over leaf or non-leaf nodes only.
type bufferedIterator struct {
	opts     int
	it       Iterator
	nextNode Node
	nextErr  error
}

func (bit *bufferedIterator) HasNext() bool {
	for bit.hasNext() {
		nxt, err := bit.peek()
		if err != nil {
			return true
		}

		// are we looking for a leaf node?
		if bit.hasLeafIterator() && nxt.Kind() == Leaf {
			return true
		}

		// are we looking for a non-leaf node?
		if bit.hasNodeIterator() && nxt.Kind() != Leaf {
			return true
		}
	}

	bit.resetNext()

	return false
}

func (bit *bufferedIterator) Next() (Node, error) {
	return bit.nextNode, bit.nextErr
}

func (bit *bufferedIterator) hasLeafIterator() bool {
	return bit.opts&TraverseLeaf == TraverseLeaf
}

func (bit *bufferedIterator) hasNodeIterator() bool {
	return bit.opts&TraverseNode == TraverseNode
}

func (bit *bufferedIterator) hasNext() bool {
	return bit.it.HasNext()
}

func (bit *bufferedIterator) peek() (Node, error) {
	bit.nextNode, bit.nextErr = bit.it.Next()

	return bit.nextNode, bit.nextErr
}

func (bit *bufferedIterator) resetNext() {
	bit.nextNode = nil
	bit.nextErr = nil
}
