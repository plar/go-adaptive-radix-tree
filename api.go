package art

type kind int
type Kind int

type Key []byte
type Value interface{}
type Callback func(node Node)

type Node interface {
	Kind() Kind
	Key() Key
	Value() Value
}

type Tree interface {
	Insert(key Key, value Value) (oldValue Value, updated bool)
	Delete(key Key) (oldValue Value, deleted bool)

	Search(key Key) (value Value, found bool)
	Walk(Callback)

	Minimum() (min Value, found bool)
	Maximum() (max Value, found bool)

	Size() int
}

func New() Tree {
	return newTree()
}
