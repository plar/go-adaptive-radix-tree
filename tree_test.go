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

func TestTreeSimple(t *testing.T) {
	tree := New()
	assert.NotNil(t, tree)

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
	}

	for _, s := range strs {
		tree.Insert([]byte(s), []byte(s))
	}

	for _, s := range strs {
		v, found := tree.Search([]byte(s))
		assert.True(t, found)
		if found {
			assert.Equal(t, []byte(s), []byte(v.([]byte)))
		}
	}

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

	l1 := newLeaf([]byte("abcdefg12345678"), "abcdefg12345678").Leaf()
	l2 := newLeaf([]byte("abcdefg!@#$%^&*"), "abcdefg!@#$%^&*").Leaf()
	assert.Equal(t, 7, tree.longestCommonPrefix(l1, l2, 0))
	assert.Equal(t, 3, tree.longestCommonPrefix(l1, l2, 4))

	l1 = newLeaf([]byte("abcdefg12345678"), "abcdefg12345678").Leaf()
	l2 = newLeaf([]byte("defg!@#$%^&*"), "defg!@#$%^&*").Leaf()
	assert.Equal(t, 0, tree.longestCommonPrefix(l1, l2, 0))
}
