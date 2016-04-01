package art

import (
	"bytes"
	"fmt"
	"strings"
)

type depthStorage struct {
	childNum      int
	childrenTotal int
}

type treeStringer struct {
	storage []depthStorage
	buf     *bytes.Buffer
}

func (t *tree) String() string {
	return DumpNode(t.root)
}

func (ts *treeStringer) generatePads(depth int, childNum int, childrenTotal int) (pad0, pad string) {
	for d := 0; d <= depth; d++ {
		if d < depth {
			pad0 += "│   "
		} else {
			pad0 += "├──"
		}

	}
	pad0 += " "
	pad = strings.Repeat("│   ", depth+1)
	return
}

func (ts *treeStringer) generatePadsV2(depth int, childNum int, childrenTotal int) (pad0, pad string) {
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

func (ts *treeStringer) append(str string) *treeStringer {
	ts.buf.WriteString(str)
	return ts
}

func (ts *treeStringer) array(arr []byte) *treeStringer {
	ts.append(" [")
	for _, b := range arr {
		if b > 0 {
			ts.append(string(b))
		} else {
			ts.append("·")
		}
	}
	ts.append("]")
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
		ts.array(prefix).append("\n")
	}

	if keys != nil {
		ts.append(pad).append(fmt.Sprintf("keys: %v", keys))
		ts.array(keys).append("\n")
	}

	ts.append(pad).append(fmt.Sprintf("children(%v): %+v\n", numChildren, children))
	ts.children(children, numChildren, depth+1)
}

func (ts *treeStringer) baseNode(an *artNode, depth int, childNum int, childrenTotal int) {
	//padHeader, pad := ts.generatePads(depth, childNum, childrenTotal)
	padHeader, pad := ts.generatePadsV2(depth, childNum, childrenTotal)
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
		ts.append(pad).append(fmt.Sprintf("key: %v", n.key)).array(n.key[:]).append("\n")

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
	ts := &treeStringer{make([]depthStorage, 256), bytes.NewBufferString("")}
	ts.rootNode(root)
	return ts.buf.String()
}
