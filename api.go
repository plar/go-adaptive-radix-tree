// Go implementation of Adaptive Radix Tree (ART)

package art

// node types
const (
	NODE_4    = Kind(1)
	NODE_16   = Kind(2)
	NODE_48   = Kind(3)
	NODE_256  = Kind(4)
	NODE_LEAF = Kind(5)
)

// Traverse Options
const (
	//
	TRAVERSE_LEAF = 1
	TRAVERSE_NODE = 2
	TRAVERSE_ALL  = TRAVERSE_LEAF | TRAVERSE_NODE
)

type Kind int

type Key []byte
type Value interface{}

type Callback func(node Node) (cont bool)

type Node interface {
	Kind() Kind

	// The following methods are valid only for Leaf node
	Key() Key
	Value() Value
}

type Iterator interface {
	HasNext() bool
	Next() (Node, error)
}

type Tree interface {
	Insert(key Key, value Value) (oldValue Value, updated bool)
	Delete(key Key) (oldValue Value, deleted bool)

	Search(key Key) (value Value, found bool)

	ForEach(callback Callback, options ...int)
	ForEachPrefix(keyPrefix Key, callback Callback)

	Iterator(options ...int) Iterator
	//IteratorPrefix() Iterator

	Minimum() (min Value, found bool)
	Maximum() (max Value, found bool)

	Size() int
}

func New() Tree {
	return newTree()
}
