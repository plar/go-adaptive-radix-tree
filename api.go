package art

// node types
const (
	NODE_4    = Kind(1)
	NODE_16   = Kind(2)
	NODE_48   = Kind(3)
	NODE_256  = Kind(4)
	NODE_LEAF = Kind(5)
)

type Kind int

type Key []byte
type Value interface{}

type Callback func(node Node)

type Node interface {
	Kind() Kind

	// The following method valid only for NODE_LEAF type of node
	Key() Key
	Value() Value
}

type Tree interface {
	Insert(key Key, value Value) (oldValue Value, updated bool)
	Delete(key Key) (oldValue Value, deleted bool)

	Search(key Key) (value Value, found bool)
	ForEach(Callback)

	Minimum() (min Value, found bool)
	Maximum() (max Value, found bool)

	Size() int
}

func New() Tree {
	return newTree()
}
