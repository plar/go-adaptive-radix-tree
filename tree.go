package art

// treeOpResult represents the result of the tree operation.
type treeOpResult int

const (
	// treeOpNoChange indicates that the key was not found.
	treeOpNoChange treeOpResult = iota

	// treeOpInserted indicates that the key/value was inserted.
	treeOpInserted

	// treeOpUpdated indicates that the existing key was updated with a new value.
	treeOpUpdated

	// treeOpDeleted indicates that the key was deleted.
	treeOpDeleted
)

// keyChar stores the key character and an flag
// to indicate if the key char is invalid.
type keyChar struct {
	ch      byte
	invalid bool
}

// singleton keyChar instance to indicate
// that the key char is invalid.
//
//nolint:gochecknoglobals
var keyCharInvalid = keyChar{ch: 0, invalid: true}

// charAt returns the character at the given index.
// If the index is out of bounds, it returns 0 and false.
func (k Key) charAt(idx int) keyChar {
	if k.isValid(idx) {
		return keyChar{ch: k[idx]}
	}

	return keyCharInvalid
}

// isValid checks if the given index is within the bounds of the key.
func (k Key) isValid(idx int) bool {
	return idx >= 0 && idx < len(k)
}

// tree is the main data structure of the ART tree.
type tree struct {
	version int      // version is used to detect concurrent modifications
	size    int      // size is the number of elements in the tree
	root    *nodeRef // root is the root node of the tree
}

// make sure that tree implements all methods from the Tree interface.
var _ Tree = (*tree)(nil)

// Insert inserts the given key and value into the tree.
// If the key already exists, it updates the value and
// returns the old value with second return value set to true.
func (tr *tree) Insert(key Key, value Value) (Value, bool) {
	oldVal, status := tr.insertRecursively(&tr.root, key, value, 0)
	if status == treeOpInserted {
		tr.version++
		tr.size++
	}

	return oldVal, status == treeOpUpdated
}

// Delete deletes the given key from the tree.
func (tr *tree) Delete(key Key) (Value, bool) {
	val, status := tr.deleteRecursively(&tr.root, key, 0)
	if status == treeOpDeleted {
		tr.version++
		tr.size--

		return val, true
	}

	return nil, false
}

// Search searches for the given key in the tree.
func (tr *tree) Search(key Key) (Value, bool) {
	keyOffset := 0

	current := tr.root
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
			if prefixLen != minInt(int(curNode.prefixLen), maxPrefixLen) {
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

// Minimum returns the minimum key in the tree.
func (tr *tree) Minimum() (Value, bool) {
	if tr == nil || tr.root == nil {
		return nil, false
	}

	return tr.root.minimum().value, true
}

// Maximum returns the maximum key in the tree.
func (tr *tree) Maximum() (Value, bool) {
	if tr == nil || tr.root == nil {
		return nil, false
	}

	return tr.root.maximum().value, true
}

// Size returns the number of elements in the tree.
func (tr *tree) Size() int {
	if tr == nil || tr.root == nil {
		return 0
	}

	return tr.size
}

// ForEach iterates over all keys in the tree and calls the callback function.
func (tr *tree) ForEach(callback Callback, opts ...int) {
	options := traverseOptions(opts...)
	tr.forEachRecursively(tr.root, traverseFilter(options, callback))
}

// ForEachPrefix iterates over all keys with the given prefix.
func (tr *tree) ForEachPrefix(key Key, callback Callback) {
	tr.forEachPrefix(key, callback)
}

// Iterator returns a new tree iterator.
func (tr *tree) Iterator(opts ...int) Iterator {
	return newTreeIterator(tr, traverseOptions(opts...))
}

// String returns tree in the human readable format, see DumpNode for examples.
func (tr *tree) String() string {
	return DumpNode(tr.root)
}
