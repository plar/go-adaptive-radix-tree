package art

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var words [][]byte

func loadTestFile(path string) [][]byte {
	file, err := os.Open(path)
	if err != nil {
		panic("Couldn't open " + path)
	}
	defer file.Close()

	var words [][]byte
	reader := bufio.NewReader(file)
	for {
		if line, err := reader.ReadBytes(byte('\n')); err != nil {
			break
		} else {
			if len(line) > 0 {
				words = append(words, line[:len(line)-1])
			}
		}
	}
	return words
}

func TestTreeLongestCommonPrefix(t *testing.T) {
	tree := &tree{}

	l1 := factory.newLeaf([]byte("abcdefg12345678"), "abcdefg12345678").Leaf()
	l2 := factory.newLeaf([]byte("abcdefg!@#$%^&*"), "abcdefg!@#$%^&*").Leaf()
	assert.Equal(t, 7, tree.longestCommonPrefix(l1, l2, 0))
	assert.Equal(t, 3, tree.longestCommonPrefix(l1, l2, 4))

	l1 = factory.newLeaf([]byte("abcdefg12345678"), "abcdefg12345678").Leaf()
	l2 = factory.newLeaf([]byte("defg!@#$%^&*"), "defg!@#$%^&*").Leaf()
	assert.Equal(t, 0, tree.longestCommonPrefix(l1, l2, 0))
}

func TestTreeInit(t *testing.T) {
	tree := New()
	assert.NotNil(t, tree)
}

func TestPoolFactory(t *testing.T) {
	factory := newPoolObjFactory()
	n4 := factory.newNode48()
	assert.NotNil(t, n4)
	factory.releaseNode(n4)

	n4v2 := factory.newNode48()
	assert.True(t, n4 == n4v2)

	n4v3 := factory.newNode48()
	assert.NotEqual(t, n4v2, n4v3)
}

func TestObjFactory(t *testing.T) {
	factory := newObjFactory()
	n4 := factory.newNode48()
	assert.NotNil(t, n4)
	n4v2 := factory.newNode48()
	assert.True(t, n4 != n4v2)
}

func TestTreeUpdate(t *testing.T) {
	tree := newTree()

	key := []byte("key")

	ov, updated := tree.Insert(key, "value")
	assert.Nil(t, ov)
	assert.False(t, updated)
	assert.Equal(t, 1, tree.size)
	assert.Equal(t, kind(NODE_LEAF), tree.root.kind)

	v, found := tree.Search(key)
	assert.True(t, found)
	assert.Equal(t, "value", v.(string))

	ov, updated = tree.Insert(key, "otherValue")
	assert.Equal(t, "value", ov.(string))
	assert.True(t, updated)
	assert.Equal(t, 1, tree.size)

	v, found = tree.Search(key)
	assert.True(t, found)
	assert.Equal(t, "otherValue", v.(string))
}

func TestTreeInsertSimilarPrefix(t *testing.T) {
	tree := newTree()

	tree.Insert([]byte{1}, 1)
	tree.Insert([]byte{1, 1}, 11)

	v, found := tree.Search([]byte{1, 1})
	assert.True(t, found)
	assert.Equal(t, 11, v.(int))

}

// An Art Node with a similar prefix should be split into new nodes accordingly
// And should be searchable as intended.
func TestTreeInsert3AndSearchWords(t *testing.T) {
	tree := newTree()

	searchTerms := []string{"A", "a", "aa"}

	for _, term := range searchTerms {
		tree.Insert([]byte(term), term)
	}

	for _, term := range searchTerms {
		v, found := tree.Search([]byte(term))
		assert.True(t, found)
		assert.Equal(t, term, v.(string))
	}
}

func TestTreeInsertAndGrowToBiggerNode(t *testing.T) {
	var testData = []struct {
		totalNodes byte
		expected   kind
	}{
		{5, NODE_16},
		{17, NODE_48},
		{49, NODE_256},
	}

	for _, data := range testData {
		tree := newTree()
		for i := byte(0); i < data.totalNodes; i++ {
			tree.Insert([]byte{i}, i)
		}
		assert.Equal(t, int(data.totalNodes), tree.size)
		assert.Equal(t, data.expected, tree.root.kind)
	}
}

func TestTreeInsertWordsAndMinMax(t *testing.T) {
	words := loadTestFile("test/assets/words.txt")
	tree := newTree()
	for _, w := range words {
		tree.Insert(w, w)
	}

	for _, w := range words {
		v, found := tree.Search(w)
		assert.Equal(t, w, v)
		assert.True(t, found)
	}

	minimum := tree.root.Minimum()
	assert.Equal(t, []byte("A"), minimum.value.([]byte))
	maximum := tree.root.Maximum()
	assert.Equal(t, []byte("zythum"), maximum.value.([]byte))
}

func TestTreeInsertUUIDsAndMinMax(t *testing.T) {
	words := loadTestFile("test/assets/uuid.txt")
	tree := newTree()
	for _, w := range words {
		tree.Insert(w, w)
	}

	for _, w := range words {
		v, found := tree.Search(w)
		assert.Equal(t, w, v)
		assert.True(t, found)
	}

	minimum := tree.root.Minimum()
	assert.Equal(t, []byte("00026bda-e0ea-4cda-8245-522764e9f325"), minimum.value.([]byte))
	maximum := tree.root.Maximum()
	assert.Equal(t, []byte("ffffcb46-a92e-4822-82af-a7190f9c1ec5"), maximum.value.([]byte))
}

type testData struct {
	message string
	insert  interface{}
	delete  interface{}
	size    int
	root    interface{}
}

type treeTestCustom func(data *testData, tree *tree)

func TestTreeInsertAndDeleteOperations(t *testing.T) {

	var data = []testData{
		{
			"Insert 1 Delete 1",
			[]string{"test"},
			[]string{"test"},
			0,
			nil},
		{
			"Insert 2 Delete 1",
			[]string{"test1", "test2"},
			[]string{"test2"},
			1,
			NODE_LEAF},
		{
			"Insert 2 Delete 2",
			[]string{"test1", "test2"},
			[]string{"test1", "test2"},
			0,
			nil},
		{
			"Insert 5 Delete 1",
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1"},
			4,
			NODE_4},
		{
			"Insert 5 Delete 5",
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			nil},
		{
			"Insert 17 Delete 1",
			[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17"},
			[]string{"2"},
			16,
			NODE_16},
		{
			"Insert 17 Delete 17",
			[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17"},
			[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17"},
			0,
			nil},
		{
			"Insert 49 Delete 1",
			treeTestCustom(func(data *testData, tree *tree) {
				for i := byte(0); i < 49; i++ {
					tree.Insert([]byte{i}, []byte{i})
				}
			}),
			[]byte{byte(2)},
			48,
			NODE_48},
		{
			"Insert 49 Delete 49",
			treeTestCustom(func(data *testData, tree *tree) {
				for i := byte(0); i < 49; i++ {
					tree.Insert([]byte{i}, []byte{i})
				}
			}),
			treeTestCustom(func(data *testData, tree *tree) {
				for i := byte(0); i < 49; i++ {
					term := []byte{i}
					val, deleted := tree.Delete(term)
					assert.True(t, deleted, data.message)
					assert.Equal(t, term, val)
				}
			}),
			0,
			nil},
		{
			"Insert 256 Delete 1",
			treeTestCustom(func(data *testData, tree *tree) {
				for i := 0; i < 256; i++ {
					term := bytes.NewBuffer([]byte{})
					term.WriteByte(byte(i))
					tree.Insert(term.Bytes(), term.Bytes())
				}
			}),
			[]byte{2},
			255,
			NODE_256},
		{
			"Insert 256 Delete 256",
			treeTestCustom(func(data *testData, tree *tree) {
				for i := 0; i < 256; i++ {
					term := fmt.Sprintf("%d", i)
					tree.Insert([]byte(term), term)
				}
			}),
			treeTestCustom(func(data *testData, tree *tree) {
				for i := 0; i < 256; i++ {
					term := fmt.Sprintf("%d", i)
					val, deleted := tree.Delete([]byte(term))
					assert.True(t, deleted, data.message)
					assert.Equal(t, term, val.(string))
				}
			}),
			0,
			nil},
	}

	for _, ds := range data {
		tree := newTree()

		// insert test data...
		if strs, ok := ds.insert.([]string); ok {
			for _, term := range strs {
				tree.Insert([]byte(term), term)
			}
		} else if builder, ok := ds.insert.(treeTestCustom); ok {
			builder(&ds, tree)
		}

		// delete test data and check...
		if strs, ok := ds.delete.([]string); ok {
			for _, term := range strs {
				val, deleted := tree.Delete([]byte(term))
				assert.True(t, deleted, ds.message)

				if s, ok := val.(string); ok {
					assert.Equal(t, term, s, ds.message)
				}

				_, found := tree.Search([]byte(term))
				assert.False(t, found, ds.message)
			}
		} else if bts, ok := ds.delete.([]byte); ok {
			for _, term := range bts {

				val, deleted := tree.Delete([]byte{term})
				assert.True(t, deleted, ds.message)

				assert.Equal(t, []byte{term}, val, ds.message)

				_, found := tree.Search([]byte{term})
				assert.False(t, found, ds.message)
			}
		} else if builder, ok := ds.delete.(treeTestCustom); ok {
			builder(&ds, tree)
		}

		// general assertions...
		assert.Equal(t, ds.size, tree.size, ds.message)

		if ds.root == nil {
			assert.Nil(t, tree.root, ds.message)
		} else if k, ok := ds.root.(kind); ok {
			assert.Equal(t, k, tree.root.kind, ds.message)
		} else if an, ok := ds.root.(*artNode); ok {
			assert.Equal(t, an, tree.root, ds.message)
		}

		//fmt.Println("===================================\n", DumpNode(tree.root))
	}
}

// Inserting a single value into the tree and removing it should result in a nil tree root.
func TestInsertAndDeleteOne(t *testing.T) {
	tree := newTree()
	tree.Insert([]byte("test"), []byte("data"))
	v, deleted := tree.Delete([]byte("test"))
	assert.True(t, deleted)
	assert.Equal(t, []byte("data"), v.([]byte))
	assert.Equal(t, 0, tree.size)
	assert.Nil(t, tree.root)
}

func TestInsertTwoAndDeleteOne(t *testing.T) {
	tree := newTree()
	tree.Insert([]byte("2"), []byte{2})
	tree.Insert([]byte("1"), []byte{1})

	_, found := tree.Search([]byte("2"))
	assert.True(t, found)
	_, found = tree.Search([]byte("1"))
	assert.True(t, found)

	v, deleted := tree.Delete([]byte("2"))
	assert.True(t, deleted)
	if deleted {
		assert.Equal(t, []byte{2}, v.([]byte))
	}

	assert.Equal(t, 1, tree.size)
	assert.Equal(t, NODE_LEAF, tree.root.kind)
}

func TestInsertTwoAndDeleteTwo(t *testing.T) {
	tree := newTree()
	tree.Insert([]byte("2"), []byte{2})
	tree.Insert([]byte("1"), []byte{1})

	_, found := tree.Search([]byte("2"))
	assert.True(t, found)
	_, found = tree.Search([]byte("1"))
	assert.True(t, found)

	v, deleted := tree.Delete([]byte("2"))
	assert.True(t, deleted)
	assert.Equal(t, []byte{2}, v.([]byte))

	v, deleted = tree.Delete([]byte("1"))
	assert.True(t, deleted)
	assert.Equal(t, []byte{1}, v.([]byte))

	assert.Equal(t, 0, tree.size)
	assert.Nil(t, tree.root)
}

func BenchmarkTreeInsertWords(b *testing.B) {
	b.StopTimer()
	words := loadTestFile("test/assets/words.txt")
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		tree := New()
		for _, w := range words {
			tree.Insert(w, w)
		}
	}
}

func BenchmarkTreeSearchWords(b *testing.B) {
	b.StopTimer()
	words := loadTestFile("test/assets/words.txt")
	tree := New()
	for _, w := range words {
		tree.Insert(w, w)
	}
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		for _, w := range words {
			tree.Search(w)
		}
	}
}

func BenchmarkTreeInsertUUIDs(b *testing.B) {
	b.StopTimer()
	words := loadTestFile("test/assets/uuid.txt")
	b.StartTimer()

	// var m runtime.MemStats
	// var makes int
	for n := 0; n < b.N; n++ {
		tree := New()
		for _, w := range words {
			tree.Insert(w, w)
		}

		// runtime.ReadMemStats(&m)
		// fmt.Printf("\n====== MemStats: %+v\n", m)
		// fmt.Printf("HeapSys: %d, HeapAlloc: %d, HeapIdle: %d, HeapReleased: %d, HeapObjs: %d, # %d\n", m.HeapSys, m.HeapAlloc,
		// 	m.HeapIdle, m.HeapReleased, m.HeapObjects, makes)

		// makes++
	}
}

func BenchmarkTreeSearchUUIDs(b *testing.B) {
	b.StopTimer()
	words := loadTestFile("test/assets/uuid.txt")
	tree := New()
	for _, w := range words {
		tree.Insert(w, w)
	}
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		for _, w := range words {
			tree.Search(w)
		}
	}
}
