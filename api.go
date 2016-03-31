package art

type Key []byte
type Value interface{}

type Tree interface {
	Insert(key Key, value Value) (oldValue Value, updated bool)
	Delete(key Key) (oldValue Value, deleted bool)
	Search(key Key) (value Value, found bool)
	Minimum() (value Value, found bool)
	Maximum() (value Value, found bool)
}

func New() Tree {
	return newTree()
}
