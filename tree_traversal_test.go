package art

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTreeTraversalPreordered(t *testing.T) {
	t.Parallel()

	tree := newTree()
	tree.Insert(Key("1"), 1)
	tree.Insert(Key("2"), 2)

	traversal := []Node{}

	tree.ForEach(func(node Node) bool {
		traversal = append(traversal, node)

		return true
	}, TraverseAll)

	assert.Equal(t, 2, tree.size)
	assert.Equal(t, Node4, traversal[0].Kind())
	assert.Equal(t, tree.root, traversal[0])
	assert.Nil(t, traversal[0].Key())
	assert.Nil(t, traversal[0].Value())

	assert.Equal(t, Key("1"), traversal[1].Key())
	assert.Equal(t, 1, traversal[1].Value())
	assert.Equal(t, Leaf, traversal[1].Kind())

	assert.Equal(t, Key("2"), traversal[2].Key())
	assert.Equal(t, 2, traversal[2].Value())
	assert.Equal(t, Leaf, traversal[2].Kind())

	tree.ForEach(func(node Node) bool {
		assert.Equal(t, Node4, node.Kind())

		return true
	}, TraverseNode)
}

func TestTreeTraversalNode48(t *testing.T) {
	t.Parallel()

	tree := newTree()
	for i := 48; i > 0; i-- {
		tree.Insert(Key{byte(i)}, i)
	}

	traversal := []Node{}

	tree.ForEach(func(node Node) bool {
		traversal = append(traversal, node)

		return true
	}, TraverseAll)

	// Ensure all nodes are inserted and traversed in order.
	assert.Equal(t, 48, tree.size)
	assert.Equal(t, tree.root, traversal[0])
	assert.Equal(t, Node48, traversal[0].Kind())

	for i := 1; i <= 48; i++ {
		assert.Equal(t, Key{byte(i)}, traversal[i].Key())
		assert.Equal(t, Leaf, traversal[i].Kind())
	}
}

func TestTreeTraversalCancelEarly(t *testing.T) {
	t.Parallel()

	tree := newTree()
	for i := 0; i < 10; i++ {
		tree.Insert(Key{byte(i)}, i)
	}

	assert.Equal(t, 10, tree.Size())

	count := 0

	tree.ForEach(func(Node) bool {
		count++

		return count < 5
	}, TraverseAll)

	assert.Equal(t, 5, count)
}

func TestTreeTraversalWordsStats(t *testing.T) {
	t.Parallel()

	tree, _ := treeWithData("test/assets/words.txt")
	stats := collectStats(tree.Iterator(TraverseAll))

	assert.Equal(t, treeStats{235886, 113419, 10433, 403, 1}, stats)
}

func TestTreeTraversalPrefix(t *testing.T) { //nolint:funlen
	t.Parallel()

	dataSet := []struct {
		prefix   string
		keys     []string
		expected []string
	}{
		{
			"empty",
			[]string{},
			[]string{},
		},
		{
			"api",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "api.foo", "api"},
		},
		{
			"a",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
		}, {
			"b",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{},
		},
		{
			"api.",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "api.foo"},
		},
		{
			"api.foo.bar",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{"api.foo.bar"},
		},
		{
			"api.end",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{},
		},
		{
			"",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
		},
		{
			"this:key:has",
			[]string{
				"this:key:has:a:long:prefix:3",
				"this:key:has:a:long:common:prefix:2",
				"this:key:has:a:long:common:prefix:1",
			},
			[]string{
				"this:key:has:a:long:prefix:3",
				"this:key:has:a:long:common:prefix:2",
				"this:key:has:a:long:common:prefix:1",
			},
		},
		{
			"ele",
			[]string{"elector", "electibles", "elect", "electible"},
			[]string{"elector", "electibles", "elect", "electible"},
		},
		{
			"long.api.url.v1",
			[]string{"long.api.url.v1.foo", "long.api.url.v1.bar", "long.api.url.v2.foo"},
			[]string{"long.api.url.v1.foo", "long.api.url.v1.bar"},
		},
	}

	for _, tt := range dataSet {
		tt := tt
		t.Run("Prefix-"+tt.prefix, func(t *testing.T) {
			t.Parallel()

			tree := newTree()
			for _, k := range tt.keys {
				tree.Insert(Key(k), k)
			}

			actual := []string{}

			tree.ForEachPrefix(Key(tt.prefix), func(node Node) bool {
				if node.Kind() == Leaf {
					actual = append(actual, string(node.Key()))
				}

				return true
			})

			sort.Strings(tt.expected)
			sort.Strings(actual)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestTreeTraversalForEachPrefixWithSimilarKey(t *testing.T) {
	t.Parallel()

	tree := newTree()
	tree.Insert(Key("abc0"), "0")
	tree.Insert(Key("abc1"), "1")
	tree.Insert(Key("abc2"), "2")

	totalKeys := 0

	tree.ForEachPrefix(Key("abc"), func(node Node) bool {
		if node.Kind() == Leaf {
			totalKeys++
		}

		return true
	})

	assert.Equal(t, 3, totalKeys)
}

func TestTreeTraversalForEachPrefixConditionalCallback(t *testing.T) {
	t.Parallel()

	tree := newTree()
	tree.Insert(Key("America#California#Irvine"), 1)
	tree.Insert(Key("America#California#Sanfrancisco"), 2)
	tree.Insert(Key("America#California#LosAngeles"), 3)

	count := 0

	tree.ForEachPrefix(Key("America#"), func(node Node) bool {
		if node.Kind() == Leaf {
			count++
		}

		return true
	})
	assert.Equal(t, 3, count)

	count = 0

	tree.ForEachPrefix(Key("America#"), func(node Node) bool {
		if node.Kind() == Leaf {
			count++

			if string(node.Key()) == "America#California#Irvine" {
				return false
			}

			count++ // should not be called
		}

		return true
	})
	assert.Equal(t, 1, count)
}

func TestEarlyPrefixTraversalStop(t *testing.T) {
	t.Parallel()

	totalCalls := 0

	tree := New()
	tree.Insert(Key("0"), "0")
	tree.Insert(Key("1"), "1")
	tree.Insert(Key("11"), "11")
	tree.Insert(Key("111"), "111")
	tree.ForEachPrefix(Key("11"), func(Node) bool {
		totalCalls++

		return false
	})
	assert.Equal(t, 1, totalCalls)
}

func TestTreeTraversalForEachPrefixCallbackStop(t *testing.T) {
	t.Parallel()

	totalCalls := 0

	tree := New()
	tree.Insert(Key("0"), "0")
	tree.Insert(Key("1"), "1")
	tree.Insert(Key("11"), "11")
	tree.Insert(Key("111"), "111")
	tree.Insert(Key("1111"), "1111")
	tree.Insert(Key("11111"), "11111")
	tree.ForEachPrefix(Key("0"), func(Node) /*cont*/ bool {
		totalCalls++

		return false
	})
	assert.Equal(t, 1, totalCalls)

	totalCalls = 0

	tree.ForEachPrefix(Key("11"), func(Node) /*cont*/ bool {
		totalCalls++

		return false
	})
	assert.Equal(t, 1, totalCalls)

	totalCalls = 0

	tree.ForEachPrefix(Key("nokey"), func(Node) /*cont*/ bool {
		totalCalls++ // should be never called

		return false
	})
	assert.Equal(t, 0, totalCalls)
}

func TestPrefixTraversalWords(t *testing.T) {
	t.Parallel()

	var found []string

	tree, _ := treeWithData("test/assets/words.txt")
	tree.ForEachPrefix(Key("antisa"), func(node Node) bool {
		if node.Kind() == Leaf {
			val, ok := node.Value().([]byte)
			assert.True(t, ok)

			found = append(found, string(val))
		}

		return true
	})

	expected := []string{
		"antisacerdotal",
		"antisacerdotalist",
		"antisaloon",
		"antisalooner",
		"antisavage",
	}
	assert.Equal(t, expected, found)
}

func TestTreeIterator(t *testing.T) {
	t.Parallel()

	tree := newTree()
	tree.Insert(Key("2"), []byte{2})
	tree.Insert(Key("1"), []byte{1})

	it := tree.Iterator(TraverseAll)
	assert.NotNil(t, it)
	assert.True(t, it.HasNext())

	n4, err := it.Next()
	require.NoError(t, err)
	assert.Equal(t, Node4, n4.Kind())

	assert.True(t, it.HasNext())
	v1, err := it.Next()
	require.NoError(t, err)
	assert.Equal(t, []byte{1}, v1.Value())

	assert.True(t, it.HasNext())
	v2, err := it.Next()
	require.NoError(t, err)
	assert.Equal(t, []byte{2}, v2.Value())

	assert.False(t, it.HasNext())
	bad, err := it.Next()
	assert.Nil(t, bad)
	assert.Equal(t, ErrNoMoreNodes, err)
}

func TestTreeIteratorConcurrentModification(t *testing.T) {
	t.Parallel()

	tree := newTree()
	tree.Insert(Key("2"), []byte{2})
	tree.Insert(Key("1"), []byte{1})

	it1 := tree.Iterator(TraverseAll)
	assert.NotNil(t, it1)
	assert.True(t, it1.HasNext())

	// simulate concurrent modification
	tree.Insert(Key("3"), []byte{3})

	bad, err := it1.Next()
	assert.Nil(t, bad)
	assert.Equal(t, ErrConcurrentModification, err)

	it2 := tree.Iterator(TraverseAll)
	assert.NotNil(t, it2)
	assert.True(t, it2.HasNext())

	tree.Delete([]byte("3"))

	bad, err = it2.Next()
	assert.Nil(t, bad)
	assert.Equal(t, ErrConcurrentModification, err)

	// test buffered ConcurrentModification
	it3 := tree.Iterator(TraverseNode)
	assert.NotNil(t, it3)
	tree.Insert(Key("3"), []byte{3})
	assert.True(t, it3.HasNext())
	bad, err = it3.Next()
	assert.Nil(t, bad)
	assert.Equal(t, ErrConcurrentModification, err)
}

func TestTreeIterateWordsStats(t *testing.T) {
	t.Parallel()

	tree, _ := treeWithData("test/assets/words.txt")
	stats := collectStats(tree.Iterator(TraverseAll))
	assert.Equal(t, treeStats{235886, 113419, 10433, 403, 1}, stats)

	stats = collectStats(tree.Iterator(TraverseLeaf))
	assert.Equal(t, treeStats{235886, 0, 0, 0, 0}, stats)

	stats = collectStats(tree.Iterator(TraverseNode))
	assert.Equal(t, treeStats{0, 113419, 10433, 403, 1}, stats)

	// by default Iterator traverses only leaf nodes
	stats = collectStats(tree.Iterator())
	assert.Equal(t, treeStats{235886, 0, 0, 0, 0}, stats)
}

func TestIteratorHasNextDoesNotAdvanceState(t *testing.T) {
	t.Parallel()

	tree := newTree()
	tree.Insert(Key("1"), []byte{1})
	tree.Insert(Key("2"), []byte{2})

	iter := tree.Iterator()

	// HasNext should not advance the iterator state
	assert.True(t, iter.HasNext())
	assert.True(t, iter.HasNext())

	// change the iterator state
	n, err := iter.Next()
	require.NoError(t, err)
	assert.Equal(t, Key("1"), n.Key())

	// HasNext remains idempotent
	assert.True(t, iter.HasNext())
	assert.True(t, iter.HasNext())

	// advance to the second key
	n, err = iter.Next()
	require.NoError(t, err)
	assert.Equal(t, Key("2"), n.Key())

	// HasNext returns false at the end
	assert.False(t, iter.HasNext())
	assert.False(t, iter.HasNext())

	// calling Next after the iterator is exhausted
	for i := 0; i < 2; i++ {
		n, err = iter.Next()
		assert.Nil(t, n, "Next() should return nil after exhaustion")
		assert.Equal(t, ErrNoMoreNodes, err, "Next() should return ErrNoMoreNodes after exhaustion")
	}
}

func TestIteratorEmptyTreeBehavior(t *testing.T) {
	t.Parallel()

	tree := New()
	iter := tree.Iterator()

	// HasNext should return false for an empty tree
	assert.False(t, iter.HasNext())
	assert.False(t, iter.HasNext())

	// Next should return nil and ErrNoMoreNodes for an empty tree
	n, err := iter.Next()
	assert.Nil(t, n)
	assert.Equal(t, ErrNoMoreNodes, err)
}
