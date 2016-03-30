package art

type tree struct {
	root *artNode
	size uint64
}

func (t *tree) Insert(key Key, value Value) (Value, bool) {
	// fmt.Printf("========== INSERT %v (%v)\n", key, string(key))
	oldValue, updated := t.recursiveInsert(t.root, &t.root, key, value, 0)
	if !updated {
		t.size++
	}
	//fmt.Printf("========== %+v\n%+v\n\n\n", t.root, t.size)
	return oldValue, updated
}

func (t *tree) Delete(key Key) (Value, bool) {
	return nil, false
}

func (t *tree) Search(key Key) (Value, bool) {
	var child **artNode

	n := t.root

	prefixLen := 0
	depth := 0

	for n != nil {
		if n.kind == NODE_LEAF {
			leaf := n.Leaf()
			if leaf.Match(key, depth) {
				return leaf.value, true
			}

			return nil, false
		}

		base := n.BaseNode()
		if base.prefixLen > 0 {
			prefixLen = base.CheckPrefix(key, depth)
			if prefixLen != min(MAX_PREFIX_LENGTH, base.prefixLen) {
				return nil, false
			}
			depth += base.prefixLen
		}

		child = n.FindChild(keyChar(key, depth))
		if *child != nil {
			n = *child
		} else {
			n = nil
		}
		depth++
	}

	return nil, false
}

func (t *tree) recursiveInsert(n *artNode, ref **artNode, key Key, value Value, depth int) (Value, bool) {
	// fmt.Printf("-> recursiveInsert\n-> node: %+v\n-> ref: %+v\n-> key: %+v\n-> depth: %v\n", n, ref, string(key), depth)

	if n == nil {
		*ref = newLeaf(key, value)
		return nil, false
	}

	if n.kind == NODE_LEAF {
		leaf := n.Leaf()

		// update exists value
		if leaf.Match(key, depth) {
			oldValue := leaf.value
			leaf.value = value
			return oldValue, true
		}

		newNode := newNode4()
		newNode4 := newNode.Node4()

		// new value, split the leaf newNode into a node4
		leafNew := newLeaf(key, value)
		leaf2 := leafNew.Leaf()
		longestPrefix := t.longestCommonPrefix(leaf, leaf2, depth)
		newNode4.prefixLen = longestPrefix
		if longestPrefix > 0 {
			newNode4.SetKey(key[depth:], min(MAX_PREFIX_LENGTH, longestPrefix))
		}
		depth += longestPrefix
		newNode.AddChild(&newNode, keyChar(leaf.key, depth), n)
		newNode.AddChild(&newNode, keyChar(leaf2.key, depth), leafNew)
		*ref = newNode
		return nil, false
	}

	bn := n.BaseNode()
	if bn.prefixLen > 0 {
		prefixDiff := n.PrefixMismatch(key, depth)
		if prefixDiff >= bn.prefixLen {
			depth += bn.prefixLen
			goto RECURSION_SEARCH
		}
		newNode := newNode4()
		nbn := newNode.BaseNode()
		nbn.prefixLen = prefixDiff
		for j := 0; j < min(MAX_PREFIX_LENGTH, prefixDiff); j++ {
			nbn.prefix[j] = bn.prefix[j]
		}

		// Insert the new leaf
		leaf := newLeaf(key, value)
		newNode.AddChild(&newNode, keyChar(key, depth+prefixDiff), leaf)

		if bn.prefixLen <= MAX_PREFIX_LENGTH {
			newNode.AddChild(&newNode, prefixChar(bn.prefix, prefixDiff), n)
			bn.prefixLen -= (prefixDiff + 1)
			for j, limit := 0, min(MAX_PREFIX_LENGTH, bn.prefixLen); j < limit; j++ {
				bn.prefix[j] = bn.prefix[prefixDiff+1+j]
			}

		} else {
			bn.prefixLen -= (prefixDiff + 1)
			leaf := n.Minimum()
			c := keyChar(leaf.key, depth+prefixDiff)
			newNode.AddChild(&newNode, c, n)
			for j, limit := 0, min(MAX_PREFIX_LENGTH, bn.prefixLen); j < limit; j++ {
				bn.prefix[j] = leaf.key[depth+prefixDiff+1+j]
			}
		}

		*ref = newNode
		return nil, false
	}

RECURSION_SEARCH:

	// Find a child to recursive to
	child := n.FindChild(keyChar(key, depth))
	if *child != nil {
		return t.recursiveInsert(*child, child, key, value, depth+1)
	}

	// No Child, artNode goes with us
	n.AddChild(ref, keyChar(key, depth), newLeaf(key, value))
	return nil, false
}

func (t *tree) recursiveDelete(n *artNode, ref **artNode, key Key, depth int) (Value, bool) {
	if n == nil || key == nil {
		return nil, false
	}

	return nil, false
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
