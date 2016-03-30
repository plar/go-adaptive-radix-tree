package art

import (
	"bufio"
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

// func init() {
// 	words = make([][]byte, 0, 300000)

// 	file, err := os.Open("test/assets/words3.txt")
// 	if err != nil {
// 		panic("Couldn't open words.txt")
// 	}

// 	defer file.Close()

// 	reader := bufio.NewReader(file)

// 	for {
// 		if line, err := reader.ReadBytes(byte('\n')); err != nil {
// 			break
// 		} else {
// 			var word []byte
// 			l := len(line)
// 			if l > 0 && (line[l-1] == '\n' || line[l-1] == '\r') {
// 				word = line[:len(line)-1]
// 			} else {
// 				word = line
// 			}

// 			if len(word) > 0 {
// 				words = append(words, []byte("/"+string(word)))
// 			}

// 			//fmt.Printf("Line: '%v'\n", string(word))
// 		}
// 	}

// }

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

func TestTreeSimple(t *testing.T) {
	tree := New()

	strs := []string{
		"bacteria",
		"Bacteriaceae",
		"bacteriaceous",
		"bacterial",
		"bacterially",
		"bacterian",
		"bacteric",
		"bacterian1",
		"bacteric2",
		"bacbacteria",
		"BacBacteriaceae",
		"bacbacteriaceous",
		"b1cbacteria",
		"Ba2Bacteriaceae",
		"bac3acteriaceous",
		"bacb4cteria",
		"BacBa5teriaceae",
		"bacbac6eriaceous",
		"bacbact7ria",
		"BacBacte8iaceae",
		"bacbacter9aceous",
		"bacbacteri0",
		"BacBacteria1eae",
		"bacbacteriac1ous",
		"1b1cbacteria",
		"2Ba2Bacteriaceae",
		"3bac3acteriaceous",
		"4bacb4cteria",
		"5BacBa5teriaceae",
		"6bacbac6eriaceous",
		"7bacbact7ria",
		"8BacBacte8iaceae",
		"9bacbacter9aceous",
		"1bacbacteri0",
		"11BacBacteria1eae",
		"111bacbacteriac1ous",
	}

	for _, s := range strs {
		tree.Insert([]byte(s), []byte(s))
		// fmt.Println("=========================================== ", s)
		// fmt.Println(tree)
	}

	for _, s := range strs {
		v, found := tree.Search([]byte(s))
		assert.True(t, found)
		if found {
			assert.Equal(t, []byte(s), []byte(v.([]byte)))
		}
	}
}

func TestTreeUpdate(t *testing.T) {
	tree := New()

	key := []byte("key")

	ov, updated := tree.Insert(key, "value")
	assert.Nil(t, ov)
	assert.False(t, updated)

	v, found := tree.Search(key)
	assert.True(t, found)
	assert.Equal(t, "value", v.(string))

	ov, updated = tree.Insert(key, "otherValue")
	assert.Equal(t, "value", ov.(string))
	assert.True(t, updated)

	v, found = tree.Search(key)
	assert.True(t, found)
	assert.Equal(t, "otherValue", v.(string))
}

func TestTreeInsertWords(t *testing.T) {
	words := loadTestFile("test/assets/words.txt")

	tree := New()

	for _, w := range words {
		tree.Insert(w, w)
	}

	for _, w := range words {
		v, found := tree.Search(w)
		assert.Equal(t, w, v)
		assert.True(t, found)
	}
	//fmt.Println(tree)
}

func TestTreeInsertUUIDs(t *testing.T) {

	words := loadTestFile("test/assets/uuid.txt")

	tree := New()
	for _, w := range words {
		tree.Insert(w, w)
	}

	for _, w := range words {
		v, found := tree.Search(w)
		assert.Equal(t, w, v)
		assert.True(t, found)
	}
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
