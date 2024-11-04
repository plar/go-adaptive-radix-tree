package art

// treeOpResult represents the result of the tree operation.
type treeOpResult int

const (
	// treeOpInserted indicates that the key/value was inserted.
	treeOpInserted treeOpResult = iota

	// treeOpUpdated indicates that the existing key
	// was updated with a new value.
	treeOpUpdated

	// treeOpDeleted indicates that the key was deleted.
	treeOpDeleted

	// treeOpNotFound indicates that the key was not found.
	treeOpNoChange
)

// tree is the main data structure of the ART tree.
type tree struct {
	version int      // version is used to detect concurrent modifications
	size    int      // size is the number of elements in the tree
	root    *nodeRef // root is the root node of the tree
}

// make sure that tree implements all methods from the Tree interface
var _ Tree = (*tree)(nil)

func (tr *tree) Insert(key Key, value Value) (Value, bool) {
	oldVal, status := tr.insertRecursively(&tr.root, key, value, 0)
	if status == treeOpInserted {
		tr.version++
		tr.size++
	}

	return oldVal, status == treeOpUpdated
}

func (tr *tree) Delete(key Key) (Value, bool) {
	val, status := tr.deleteRecursively(&tr.root, key, 0)
	if status == treeOpDeleted {
		tr.version++
		tr.size--
		return val, true
	}

	return nil, false
}

func (tr *tree) Search(key Key) (Value, bool) {
	current := tr.root
	keyOffset := 0
	for current != nil {
		if current.isLeaf() {
			leaf := current.leaf()
			if leaf.match(key) {
				return leaf.value, true
			}

			return nil, false
		}

		curNode := current.node()

		if curNode.prefixLen > 0 {
			prefixLen := current.match(key, keyOffset)
			if prefixLen != min(int(curNode.prefixLen), MaxPrefixLen) {
				return nil, false
			}
			keyOffset += int(curNode.prefixLen)
		}

		next := current.findChildByKey(key, keyOffset)
		if *next != nil {
			current = *next
		} else {
			current = nil
		}
		keyOffset++
	}

	return nil, false
}

func (tr *tree) Minimum() (value Value, found bool) {
	if tr == nil || tr.root == nil {
		return nil, false
	}

	leaf := tr.root.minimum()

	return leaf.value, true
}

func (tr *tree) Maximum() (value Value, found bool) {
	if tr == nil || tr.root == nil {
		return nil, false
	}

	leaf := tr.root.maximum()

	return leaf.value, true
}

func (tr *tree) Size() int {
	if tr == nil || tr.root == nil {
		return 0
	}

	return tr.size
}

func (tr *tree) ForEach(callback Callback, opts ...int) {
	options := traverseOptions(opts...)
	tr.forEachRecursively(tr.root, traverseFilter(options, callback))
}

func (tr *tree) ForEachPrefix(key Key, callback Callback) {
	tr.forEachPrefix(tr.root, key, callback)
}

// Iterator pattern
func (tr *tree) Iterator(opts ...int) Iterator {
	options := traverseOptions(opts...)

	it := newTreeIterator(tr)
	if options&TraverseAll == TraverseAll {
		return it
	}

	bti := &bufferedIterator{
		options: options,
		it:      it,
	}
	return bti
}

// String returns tree in the human readable format, see DumpNode for examples
func (tr *tree) String() string {
	return DumpNode(tr.root)
}
