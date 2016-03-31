package art

type tree struct {
	root *artNode
	size int
}

func (t *tree) Insert(key Key, value Value) (Value, bool) {
	oldValue, updated := t.recursiveInsert(&t.root, key, value, 0)
	if !updated {
		t.size++
	}
	return oldValue, updated
}

func (t *tree) Delete(key Key) (Value, bool) {
	value, deleted := t.recursiveDelete(&t.root, key, 0)
	if deleted {
		t.size--
		return value, true
	}

	return nil, false
}

func (t *tree) Search(key Key) (Value, bool) {
	node := t.root

	prefixLen := 0
	depth := 0

	for node != nil {
		if node.kind == NODE_LEAF {
			leaf := node.Leaf()
			if leaf.Match(key) {
				return leaf.value, true
			}
			return nil, false
		}

		nodeBase := node.BaseNode()
		if nodeBase.prefixLen > 0 {
			prefixLen = nodeBase.CheckPrefix(key, depth)
			if prefixLen != min(MAX_PREFIX_LENGTH, nodeBase.prefixLen) {
				return nil, false
			}
			depth += nodeBase.prefixLen
		}

		next := node.FindChild(key.charAt(depth))
		if *next != nil {
			node = *next
		} else {
			node = nil
		}
		depth++
	}

	return nil, false
}

func (t *tree) Minimum() (value Value, found bool) {
	if t == nil || t.root == nil {
		return nil, false
	}

	leaf := t.root.Minimum()
	if leaf != nil {
		return leaf.value, true
	}
	return nil, false
}

func (t *tree) Maximum() (value Value, found bool) {
	if t == nil || t.root == nil {
		return nil, false
	}

	leaf := t.root.Maximum()
	if leaf != nil {
		return leaf.value, true
	}
	return nil, false
}

func (t *tree) Size() int {
	if t == nil || t.root == nil {
		return 0
	}

	return t.size
}

func (t *tree) Walk(callback Callback) {
	if t == nil || t.root == nil {
		return
	}

	t.walker(t.root, callback)
}

func (t *tree) walker(current *artNode, callback Callback) {
	if current == nul {
		return
	}

	callback(current)

}

func (t *tree) recursiveInsert(curNode **artNode, key Key, value Value, depth int) (Value, bool) {
	node := *curNode
	if node == nil {
		replaceRef(curNode, factory.newLeaf(key, value))
		return nil, false
	}

	if node.kind == NODE_LEAF {
		leaf := node.Leaf()

		// update exists value
		if leaf.Match(key) {
			oldValue := leaf.value
			leaf.value = value
			return oldValue, true
		}

		// new value, split the leaf newNode into a node4
		newLeaf := factory.newLeaf(key, value)
		leaf2 := newLeaf.Leaf()
		lcp := t.longestCommonPrefix(leaf, leaf2, depth)

		newNode := factory.newNode4()
		newNode.SetPrefix(key[depth:], lcp)
		depth += lcp

		newNode.AddChild(leaf.key.charAt(depth), node)
		newNode.AddChild(leaf2.key.charAt(depth), newLeaf)
		replaceRef(curNode, newNode)
		return nil, false
	}

	nodeBase := node.BaseNode()
	if nodeBase.prefixLen > 0 {
		prefixDiff := node.PrefixMismatch(key, depth)
		if prefixDiff >= nodeBase.prefixLen {
			depth += nodeBase.prefixLen
			goto NEXT_NODE
		}
		newNode := factory.newNode4()
		newNodeBase := newNode.BaseNode()
		newNodeBase.prefixLen = prefixDiff
		for j := 0; j < min(MAX_PREFIX_LENGTH, prefixDiff); j++ {
			newNodeBase.prefix[j] = nodeBase.prefix[j]
		}

		if nodeBase.prefixLen <= MAX_PREFIX_LENGTH {
			newNode.AddChild(nodeBase.prefix[prefixDiff], node)
			nodeBase.prefixLen -= (prefixDiff + 1)

			for j, limit := 0, min(MAX_PREFIX_LENGTH, nodeBase.prefixLen); j < limit; j++ {
				nodeBase.prefix[j] = nodeBase.prefix[prefixDiff+1+j]
			}

		} else {
			nodeBase.prefixLen -= (prefixDiff + 1)
			leaf := node.Minimum()
			newNode.AddChild(leaf.key.charAt(depth+prefixDiff), node)

			for j, limit := 0, min(MAX_PREFIX_LENGTH, nodeBase.prefixLen); j < limit; j++ {
				nodeBase.prefix[j] = leaf.key.charAt(depth + prefixDiff + 1 + j)
			}
		}

		// Insert the new leaf
		newNode.AddChild(key.charAt(depth+prefixDiff), factory.newLeaf(key, value))
		replaceRef(curNode, newNode)
		return nil, false
	}

NEXT_NODE:

	// Find a child to recursive to
	next := node.FindChild(key.charAt(depth))
	if *next != nil {
		return t.recursiveInsert(next, key, value, depth+1)
	}

	// No Child, artNode goes with us
	node.AddChild(key.charAt(depth), factory.newLeaf(key, value))
	return nil, false
}

func (t *tree) recursiveDelete(curNode **artNode, key Key, depth int) (Value, bool) {
	if t == nil || *curNode == nil || len(key) == 0 {
		return nil, false
	}

	node := *curNode

	if node.kind == NODE_LEAF {
		leaf := node.Leaf()
		if leaf.Match(key) {
			replaceRef(curNode, nil)

			return leaf.value, true
		}

		return nil, false
	}

	nodeBase := node.BaseNode()
	if nodeBase.prefixLen > 0 {

		prefixLen := nodeBase.CheckPrefix(key, depth)
		if prefixLen != min(MAX_PREFIX_LENGTH, nodeBase.prefixLen) {
			return nil, false
		}

		depth += nodeBase.prefixLen
	}

	next := node.FindChild(key.charAt(depth))
	if *next == nil {
		return nil, false
	}

	if (*next).kind == NODE_LEAF {
		leaf := (*next).Leaf()
		if leaf.Match(key) {
			node.DeleteChild(key.charAt(depth))
			return leaf.value, true
		}
		return nil, false

	}

	return t.recursiveDelete(next, key, depth+1)
}

func (t *tree) longestCommonPrefix(l1 *leaf, l2 *leaf, depth int) int {
	l1key, l2key := l1.key, l2.key
	idx, limit := depth, min(len(l1key), len(l2key))

	//fmt.Printf("%+v %+v %v %v\n", l1, l2, depth, limit)
	for ; idx < limit; idx++ {
		if l1key[idx] != l2key[idx] {
			break
		}
	}
	return idx - depth
}
