package art

import (
	"bytes"
	"fmt"
)

const (
	printValuesAsChar = 1 << iota
	printValuesAsDecimal

	printValueDefault = printValuesAsChar
)

type depthStorage struct {
	childNum      int
	childrenTotal int
}

type treeStringer struct {
	storage []depthStorage
	buf     *bytes.Buffer
}

// String returns tree in the human readable format, see DumpNode for examples
func (t *tree) String() string {
	return DumpNode(t.root)
}

func (ts *treeStringer) generatePads(depth int, childNum int, childrenTotal int) (pad0, pad string) {
	ts.storage[depth] = depthStorage{childNum, childrenTotal}

	for d := 0; d <= depth; d++ {
		if d < depth {
			if ts.storage[d].childNum+1 < ts.storage[d].childrenTotal {
				pad0 += "│   "
			} else {
				pad0 += "    "
			}
		} else {
			if childrenTotal == 0 {
				pad0 += "─"
			} else if ts.storage[d].childNum+1 < ts.storage[d].childrenTotal {
				pad0 += "├"
			} else {
				pad0 += "└"
			}
			pad0 += "──"
		}

	}
	pad0 += " "

	for d := 0; d <= depth; d++ {
		if childNum+1 < childrenTotal && childrenTotal > 0 {
			if ts.storage[d].childNum+1 < ts.storage[d].childrenTotal {
				pad += "│   "
			} else {
				pad += "    "
			}
		} else if d < depth && ts.storage[d].childNum+1 < ts.storage[d].childrenTotal {
			pad += "│   "
		} else {
			pad += "    "
		}

	}

	return
}

func (ts *treeStringer) append(v interface{}, opts ...int) *treeStringer {
	options := 0
	for _, opt := range opts {
		options |= opt
	}

	if options == 0 {
		options = printValueDefault
	}

	switch v.(type) {

	case string:
		str, _ := v.(string)
		ts.buf.WriteString(str)

	case []byte:
		arr, _ := v.([]byte)
		ts.append("[")
		for i, b := range arr {
			if (options & printValuesAsChar) != 0 {
				if b > 0 {
					ts.append(fmt.Sprintf("%c", b))
				} else {
					ts.append("·")
				}

			} else if (options & printValuesAsDecimal) != 0 {
				ts.append(fmt.Sprintf("%d", b))
			}
			if (options&printValuesAsDecimal) != 0 && i+1 < len(arr) {
				ts.append(" ")
			}
		}
		ts.append("]")

	case Key:
		k, _ := v.(Key)
		ts.append("[").append(string(k)).append("]")

	default:
		ts.append("[")
		ts.append(fmt.Sprintf("%#v", v))
		ts.append("]")
	}
	return ts
}

func (ts *treeStringer) children(children []*artNode, numChildred int, depth int) {
	for i := 0; i < len(children); i++ {
		ts.baseNode(children[i], depth, i, len(children))
	}
}

func (ts *treeStringer) node(pad string, prefixLen int, prefix []byte, keys []byte, children []*artNode, numChildren int, depth int) {
	if prefix != nil {
		ts.append(pad).append(fmt.Sprintf("prefix(%x): %v", prefixLen, prefix))
		ts.append(prefix).append("\n")
	}

	if keys != nil {
		ts.append(pad).append("keys: ").append(keys, printValuesAsDecimal).append(" ")
		ts.append(keys, printValuesAsChar).append("\n")
	}

	ts.append(pad).append(fmt.Sprintf("children(%v): %+v\n", numChildren, children))
	ts.children(children, numChildren, depth+1)
}

func (ts *treeStringer) baseNode(an *artNode, depth int, childNum int, childrenTotal int) {
	padHeader, pad := ts.generatePads(depth, childNum, childrenTotal)
	if an == nil {
		ts.append(padHeader).append("nil").append("\n")
		return
	}

	ts.append(padHeader)
	ts.append(fmt.Sprintf("%v (%p)\n", an.kind, an))
	switch an.kind {
	case Node4:
		n := an.node()
		nn := an.node4()
		ts.node(pad, n.prefixLen, n.prefix[:], nn.keys[:], nn.children[:], n.numChildren, depth)

	case Node16:
		n := an.node()
		nn := an.node16()
		ts.node(pad, n.prefixLen, n.prefix[:], nn.keys[:], nn.children[:], n.numChildren, depth)

	case Node48:
		n := an.node()
		nn := an.node48()
		ts.node(pad, n.prefixLen, n.prefix[:], nn.keys[:], nn.children[:], n.numChildren, depth)

	case Node256:
		n := an.node()
		nn := an.node256()
		ts.node(pad, n.prefixLen, n.prefix[:], nil, nn.children[:], n.numChildren, depth)

	case Leaf:
		n := an.leaf()
		ts.append(pad).append(fmt.Sprintf("key: %v ", n.key)).append(n.key[:]).append("\n")

		if s, ok := n.value.(string); ok {
			ts.append(pad).append(fmt.Sprintf("val: %v\n", s))
		} else if b, ok := n.value.([]byte); ok {
			ts.append(pad).append(fmt.Sprintf("val: %v\n", string(b)))
		} else {
			ts.append(pad).append(fmt.Sprintf("val: %v\n", n.value))
		}

	}

	ts.append(pad).append("\n")

}

func (ts *treeStringer) rootNode(an *artNode) {
	ts.baseNode(an, 0, 0, 0)
}

/*
DumpNode returns Tree in the human readable format:
 package main

 import (
	"fmt"
	"github.com/plar/go-adaptive-radix-tree"
 )

 func main() {
 	tree := art.New()
	terms := []string{"A", "a", "aa"}
	for _, term := range terms {
		tree.Insert(art.Key(term), term)
	}
	fmt.Println(tree)
 }

 Output:
 ─── Node4 (0xc8200f3b30)
     prefix(0): [0 0 0 0 0 0 0 0 0 0] [··········]
     keys: [65 97 0 0] [Aa··]
     children(2): [0xc8200f3af0 0xc8200f3b70 <nil> <nil>]
     ├── Leaf (0xc8200f3af0)
     │   key: [65] [A]
     │   val: A
     │
     ├── Node4 (0xc8200f3b70)
     │   prefix(0): [0 0 0 0 0 0 0 0 0 0] [··········]
     │   keys: [0 97 0 0] [·a··]
     │   children(2): [0xc8200f3b20 0xc8200f3b60 <nil> <nil>]
     │   ├── Leaf (0xc8200f3b20)
     │   │   key: [97] [a]
     │   │   val: a
     │   │
     │   ├── Leaf (0xc8200f3b60)
     │   │   key: [97 97] [aa]
     │   │   val: aa
     │   │
     │   ├── nil
     │   └── nil
     │
     ├── nil
     └── nil
*/
func DumpNode(root *artNode) string {
	ts := &treeStringer{make([]depthStorage, 4096), bytes.NewBufferString("")}
	ts.rootNode(root)
	return ts.buf.String()
}
