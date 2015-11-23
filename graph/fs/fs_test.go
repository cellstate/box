package fs

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateProject(t *testing.T, dir string, size int, seed int64) {
	rnd := rand.New(rand.NewSource(seed))

	//write a large file to /
	fpath := filepath.Join(dir, "large_file")
	f, err := os.Open(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create(fpath)
			if err != nil {
				t.Fatal(err)
			}

			total := 0
			for {
				n, err := f.Write([]byte{byte(rnd.Intn(256))})
				if err != nil {
					t.Fatal(err)
				}

				total += n
				if total >= size {
					f.Seek(0, 0)
					break
				}
			}

		} else {
			t.Fatal(err)
		}
	}

	defer f.Close()

	//write non-random file to subdir
	err = os.MkdirAll(filepath.Join(dir, "a", "b"), 0777)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(dir, "a", "b", "small_file"), []byte("i'm small"), 0666)
	if err != nil {
		t.Fatal(err)
	}
}

func writeFileAt(t *testing.T, p string, data []byte, pos int64) {
	f, err := os.OpenFile(p, os.O_RDWR, 0666)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	_, err = f.WriteAt(data, pos)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGraphInterface(t *testing.T) {
	dir, err := ioutil.TempDir("", "box_test_")
	assert.NoError(t, err, "Creating temporary directory should not fail")
	l := log.New(os.Stderr, "graph/fs: ", log.Ltime|log.Lmicroseconds)
	generateProject(t, dir, 10000, 4)

	fsg, err := NewGraph(dir, l)
	assert.NoError(t, err, "Creating fs graph should not fail")

	err = fsg.Index()
	assert.NoError(t, err, "Scanning fs into graph should not fail")

	nodes, err := fsg.List()
	assert.NoError(t, err, "Retrieving all nodes in the graph should not fail")
	assert.Len(t, nodes, 7, "Graph should have this number of nodes")

	for _, n := range nodes {
		k, err := n.Key()
		assert.NoError(t, err, "Should be able to get keys of all nodes")
		assert.InDelta(t, 0, len(k), 16, "Key should be longer then 16 bytes")

		nn, err := fsg.Get(k)
		assert.NoError(t, err, "Should be able to get a node for all keys")
		assert.Equal(t, n, nn, "And retrieved node should be equal")
	}
}
