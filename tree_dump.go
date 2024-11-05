package art

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	printValuesAsChar = 1 << iota
	printValuesAsDecimal
	printValuesAsHex

	printValueDefault = printValuesAsChar
)

// refFormatter is a function that formats an artNodeRef.
type refFormatter func(*dumpNodeRef) string

// RefFullFormatter returns the full address of the node, including the ID and the pointer.
func RefFullFormatter(a *dumpNodeRef) string {
	if a.ptr == nil {
		return "-"
	}
	return fmt.Sprintf("#%d/%p", a.id, a.ptr)
}

// RefShortFormatter returns only the ID of the node.
func RefShortFormatter(a *dumpNodeRef) string {
	if a.ptr == nil {
		return "-"
	}

	return fmt.Sprintf("#%d", a.id)
}

// RefAddrFormatter returns only the pointer address of the node. (legacy)
func RefAddrFormatter(a *dumpNodeRef) string {
	if a.ptr == nil {
		return "-"
	}

	return fmt.Sprintf("%p", a.ptr)
}

// dumpNodeRef represents the address of a nodeRef in the tree,
// composed of a unique, sequential ID and a pointer to the node.
// The ID remains consistent for trees built with the same keys
// while the pointer may change with each build.
// This structure helps identify and compare nodes across different tree instances.
// It is also helpful for debugging and testing.
//
// For example: if you inserted the same keys in two different trees (or rerun the same test),
// you can compare the nodes of the two trees by their IDs.
// The IDs will be the same for the same keys, but the pointers will be different.
type dumpNodeRef struct {
	id  int          // unique ID
	ptr *nodeRef     // pointer to the node
	fmt refFormatter // function to format the address
}

// String returns the string representation of the address.
func (a dumpNodeRef) String() string {
	if a.fmt == nil {
		return RefFullFormatter(&a)
	}
	return a.fmt(&a)
}

// NodeRegistry maintains a mapping between nodeRef pointers and their unique IDs.
type nodeRegistry struct {
	ptrToID   map[*nodeRef]int // Maps a node pointer to its unique ID
	addresses []dumpNodeRef    // List of node references
	formatter refFormatter     // Function to format node references
}

// register adds a nodeRef to the registry and returns its reference.
func (nr *nodeRegistry) register(node *nodeRef) dumpNodeRef {
	// Check if the node is already registered.
	if id, exists := nr.ptrToID[node]; exists {
		return nr.addresses[id]
	}

	// Create a new reference for the node.
	id := len(nr.addresses)
	ref := dumpNodeRef{
		id:  id,
		ptr: node,
		fmt: nr.formatter,
	}

	// Register the node and its reference.
	nr.ptrToID[node] = id
	nr.addresses = append(nr.addresses, ref)

	return ref
}

// depthStorage stores information about the depth of the tree.
type depthStorage struct {
	childNum      int
	childrenTotal int
}

// treeStringer is a helper struct for generating a human-readable representation of the tree.
type treeStringer struct {
	storage      []depthStorage // Storage for depth information
	buf          *bytes.Buffer  // Buffer for building the string representation
	nodeRegistry *nodeRegistry  // Registry for node references
}

// String returns the string representation of the tree.
func (ts *treeStringer) String() string {
	s := ts.buf.String()
	// trim trailing whitespace and newlines.
	s = strings.TrimRight(s, "\n")
	s = strings.TrimRight(s, " ")
	return s
}

// regNode registers a nodeRef and returns its reference.
func (ts *treeStringer) regNode(node *nodeRef) dumpNodeRef {
	addr := ts.nodeRegistry.register(node)
	return addr
}

// regNodes registers a slice of artNodes and returns their references.
func (ts *treeStringer) regNodes(nodes []*nodeRef) []dumpNodeRef {
	if nodes == nil {
		return nil
	}

	var addrs []dumpNodeRef
	for _, n := range nodes {
		addrs = append(addrs, ts.nodeRegistry.register(n))
	}

	return addrs
}

// generatePads generates padding strings for the tree representation.
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

// append adds a string representation of a value to the buffer.
// opts is a list of options for formatting the value.
// If no options are provided, the default is to print the value as a character.
// The available options are:
// - printValuesAsChar: print values as characters
// - printValuesAsDecimal: print values as decimal numbers
// - printValuesAsHex: print values as hexadecimal numbers
func (ts *treeStringer) append(v interface{}, opts ...int) *treeStringer {
	options := 0
	for _, opt := range opts {
		options |= opt
	}

	if options == 0 {
		options = printValueDefault
	}

	switch v := v.(type) {

	case string:
		ts.buf.WriteString(v)

	case []byte:
		ts.append("[")
		for i, b := range v {
			if (options & printValuesAsChar) != 0 {
				if b > 0 {
					ts.append(fmt.Sprintf("%c", b))
				} else {
					ts.append("·")
				}

			} else if (options & printValuesAsDecimal) != 0 {
				ts.append(fmt.Sprintf("%d", b))
			}
			if (options&printValuesAsDecimal) != 0 && i+1 < len(v) {
				ts.append(" ")
			}
		}
		ts.append("]")

	case Key:
		ts.append([]byte(v))

	default:
		ts.append("[")
		ts.append(fmt.Sprintf("%#v", v))
		ts.append("]")
	}

	return ts
}

// appendKey adds a string representation of a nodeRef's key to the buffer.
// see append for the list of available options.
func (ts *treeStringer) appendKey(keys []byte, present []byte, opts ...int) *treeStringer {
	options := 0
	for _, opt := range opts {
		options |= opt
	}

	if options == 0 {
		options = printValueDefault
	}

	ts.append("[")
	for i, b := range keys {
		if (options & printValuesAsChar) != 0 {
			if present[i] != 0 {
				ts.append(fmt.Sprintf("%c", b))
			} else {
				ts.append("·")
			}

		} else if (options & printValuesAsDecimal) != 0 {
			if present[i] != 0 {
				ts.append(fmt.Sprintf("%2d", b))
			} else {
				ts.append("·")
			}
		} else if (options & printValuesAsHex) != 0 {
			if present[i] != 0 {
				ts.append(fmt.Sprintf("%2x", b))
			} else {
				ts.append("·")
			}
		}
		if (options&(printValuesAsDecimal|printValuesAsHex)) != 0 && i+1 < len(keys) {
			ts.append(" ")
		}
	}
	ts.append("]")

	return ts
}

// children generates a string representation of the children of a nodeRef.
func (ts *treeStringer) children(children []*nodeRef, _ /*numChildred*/ uint16, keyOffset int, zeroChild *nodeRef) {
	for i, child := range children {
		ts.baseNode(child, keyOffset, i, len(children)+1)
	}

	ts.baseNode(zeroChild, keyOffset, len(children)+1, len(children)+1)
}

// node generates a string representation of a nodeRef.
func (ts *treeStringer) node(pad string, prefixLen uint16, prefix []byte, keys []byte, present []byte, children []*nodeRef, numChildren uint16, keyOffset int, zeroChild *nodeRef) {
	if prefix != nil {
		ts.append(pad).
			append(fmt.Sprintf("prefix(%x): ", prefixLen)).
			append(prefix).
			append(" ").
			append(fmt.Sprintf("%v", prefix)).
			append("\n")
	}

	if keys != nil {
		ts.append(pad).
			append("keys: ").
			appendKey(keys, present, printValuesAsChar).
			append(" ").
			appendKey(keys, present, printValuesAsDecimal).
			append("\n")
	}

	ts.append(pad).
		append(fmt.Sprintf("children(%v): %+v <%v>\n",
			numChildren,
			ts.regNodes(children),
			ts.regNode(zeroChild)))

	ts.children(children, numChildren, keyOffset+1, zeroChild)
}

func (ts *treeStringer) baseNode(an *nodeRef, depth int, childNum int, childrenTotal int) {
	padHeader, pad := ts.generatePads(depth, childNum, childrenTotal)
	if an == nil {
		ts.append(padHeader).
			append("nil").
			append("\n")
		return
	}

	ts.append(padHeader).
		append(fmt.Sprintf("%v (%v)\n",
			an.kind,
			ts.regNode(an)))

	switch an.kind {
	case Node4:
		nn := an.node4()

		ts.node(pad,
			nn.prefixLen,
			nn.prefix[:],
			nn.keys[:],
			nn.present[:],
			nn.children[:node4Max],
			nn.childrenLen,
			depth,
			nn.children[node4Max])

	case Node16:
		nn := an.node16()

		var present []byte
		for i := 0; i < len(nn.keys); i++ {
			var b byte
			if nn.hasChild(i) {
				b = 1
			}
			present = append(present, b)
		}

		ts.node(pad,
			nn.prefixLen,
			nn.prefix[:],
			nn.keys[:],
			present,
			nn.children[:node16Max],
			nn.childrenLen,
			depth,
			nn.children[node16Max])

	case Node48:
		nn := an.node48()

		var present []byte
		for i := 0; i < len(nn.keys); i++ {
			var b byte
			if nn.hasChild(i) {
				b = 1
			}
			present = append(present, b)
		}

		ts.node(pad,
			nn.prefixLen,
			nn.prefix[:],
			nn.keys[:],
			present,
			nn.children[:node48Max],
			nn.childrenLen,
			depth,
			nn.children[node48Max])

	case Node256:
		nn := an.node256()

		ts.node(pad,
			nn.prefixLen,
			nn.prefix[:],
			nil,
			nil,
			nn.children[:node256Max],
			nn.childrenLen,
			depth,
			nn.children[node256Max])

	case Leaf:
		n := an.leaf()

		ts.append(pad).
			append(fmt.Sprintf("key(%d): ", len(n.key))).
			append(n.key).
			append(" ").
			append(fmt.Sprintf("%v", n.key)).
			append("\n")

		if s, ok := n.value.(string); ok {
			ts.append(pad).
				append(fmt.Sprintf("val: %v\n",
					s))
		} else if b, ok := n.value.([]byte); ok {
			ts.append(pad).
				append(fmt.Sprintf("val: %v\n",
					string(b)))
		} else {
			ts.append(pad).
				append(fmt.Sprintf("val: %v\n",
					n.value))
		}

	}

	ts.append(pad).
		append("\n")
}

func (ts *treeStringer) startFromNode(an *nodeRef) {
	ts.baseNode(an, 0, 0, 0)
}

/*
DumpNode returns Tree in the human readable format:

--8<-- // main.go

	package main

	import (
		"fmt"
		art "github.com/plar/go-adaptive-radix-tree"
	)

	func main() {
		tree := art.New()
		terms := []string{"A", "a", "aa"}
		for _, term := range terms {
			tree.Insert(art.Key(term), term)
		}
		fmt.Println(tree)
	}

--8<--

	$ go run main.go

	─── Node4 (0xc00011c2d0)
		prefix(0): [··········] [0 0 0 0 0 0 0 0 0 0]
		keys: [Aa··] [65 97 · ·]
		children(2): [0xc00011c2a0 0xc00011c300 - -] <->
		├── Leaf (0xc00011c2a0)
		│   key(1): [A] [65]
		│   val: A
		│
		├── Node4 (0xc00011c300)
		│   prefix(0): [··········] [0 0 0 0 0 0 0 0 0 0]
		│   keys: [a···] [97 · · ·]
		│   children(1): [0xc00011c2f0 - - -] <0xc00011c2c0>
		│   ├── Leaf (0xc00011c2f0)
		│   │   key(2): [aa] [97 97]
		│   │   val: aa
		│   │
		│   ├── nil
		│   ├── nil
		│   ├── nil
		│   └── Leaf (0xc00011c2c0)
		│       key(1): [a] [97]
		│       val: a
		│
		│
		├── nil
		├── nil
		└── nil
*/
func DumpNode(root *nodeRef) string {
	opts := createTreeStringerOptions(WithRefFormatter(RefAddrFormatter))
	trs := newTreeStringer(opts)
	trs.startFromNode(root)
	return trs.String()
}

// treeStringerOptions contains options for DumpTree function.
type treeStringerOptions struct {
	storageSize int
	formatter   refFormatter
}

// treeStringerOption is a function that sets an option for DumpTree.
type treeStringerOption func(opts *treeStringerOptions)

// WithStorageSize sets the size of the storage for depth information.
func WithStorageSize(size int) treeStringerOption {
	return func(opts *treeStringerOptions) {
		opts.storageSize = size
	}
}

// WithRefFormatter sets the formatter for node references.
func WithRefFormatter(formatter refFormatter) treeStringerOption {
	return func(opts *treeStringerOptions) {
		opts.formatter = formatter
	}
}

// TreeStringer returns the string representation of the tree.
// The tree must be of type *art.tree.
func TreeStringer(t Tree, opts ...treeStringerOption) string {
	tr, ok := t.(*tree)
	if !ok {
		return "expected *art.tree"
	}

	trs := newTreeStringer(createTreeStringerOptions(opts...))
	trs.startFromNode(tr.root)
	return trs.String()
}

func createTreeStringerOptions(opts ...treeStringerOption) treeStringerOptions {
	defOpts := treeStringerOptions{
		storageSize: 4096,
		formatter:   RefShortFormatter,
	}

	for _, opt := range opts {
		opt(&defOpts)
	}

	return defOpts
}

func newTreeStringer(opts treeStringerOptions) *treeStringer {
	return &treeStringer{
		storage: make([]depthStorage, opts.storageSize),
		buf:     bytes.NewBufferString(""),
		nodeRegistry: &nodeRegistry{
			ptrToID:   make(map[*nodeRef]int),
			formatter: opts.formatter,
		},
	}
}

func defaultTreeStringer() *treeStringer {
	return newTreeStringer(createTreeStringerOptions())
}
