package scanner

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/cellstate/box/graph"

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

func TestBottomsUpIndexing(t *testing.T) {
	dir, err := ioutil.TempDir("", "box_test_")
	assert.NoError(t, err, "Creating temporary directory should not fail")
	l := log.New(os.Stderr, "graph/fs: ", log.Ltime|log.Lmicroseconds)
	generateProject(t, dir, 100000, 4)

	s, err := NewScanner(l, dir)
	assert.NoError(t, err, "Creating the filesystem should not fail")

	done := make(chan bool)
	nodes := map[string]graph.Node{}
	go func() {
		for n := range s.Nodes {
			assert.Len(t, n.Key(), 20, "Expect keys of all nodes to be sha1: 20 bytes long")

			//we expect 8 nodes: 3dirs, 2files, 3parts
			nodes[fmt.Sprintf("%x", n.Key())] = n
			if len(nodes) == 11 {
				done <- true
				return
			}
		}
	}()

	err = s.Scan()
	assert.NoError(t, err, "Scanning the filesystem should not fail")

	<-done
	assert.Len(t, nodes, 11, "Expected N nodes after file edit")

	//we expect the following nodes from the scanner
	exp := []string{
		"a67316b4de11d37d722e7da5768d7d22220c2b89",
		"e025982956d87909188cd8b76699711478347de6",
		"104c9da6a7654229304fd77f4479751070453613",
		"8b2a5d310e80ad144819786e36ca4733e26939c9",

		//@todo assert large file existence
	}
	for _, k := range exp {
		if _, ok := nodes[k]; !ok {
			assert.Fail(t, "Expected node with key "+k)
		}
	}

	//we write text to the beginning of the large file
	//a rescan should yield only 3 new nodes (1 new root, 1 new file and 1 new part)
	// writeFileAt(t, filepath.Join(dir, "large_file"), []byte("foobbb"), 0)
	// fi, err := os.Stat(filepath.Join(dir, "large_file"))
	// assert.NoError(t, err, "large_file should be statable")
	// assert.Equal(t, int64(10), fi.Size(), "large_file should have remained the same size")

	// // log.Println("AAAAAAA", fi.Size())

	// go func() {
	// 	for n := range s.Nodes {
	// 		nodes[fmt.Sprintf("%x", n.Key())] = n
	// 		log.Println(len(nodes))
	// 		if len(nodes) == 11 {
	// 			done <- true
	// 			return
	// 		}
	// 	}
	// }()

	// log.Println("\n")
	// err = s.Scan()
	// assert.NoError(t, err, "Rescanning the filesystem should not fail")
	// <-done
	// assert.Len(t, nodes, 11, "Expected N nodes after file edit")

	// //we expect the following nodes from the scanner
	// exp = []string{
	// 	"a67316b4de11d37d722e7da5768d7d22220c2b89",
	// 	"e025982956d87909188cd8b76699711478347de6",
	// 	"104c9da6a7654229304fd77f4479751070453613",
	// 	"8b2a5d310e80ad144819786e36ca4733e26939c9",
	// 	"0c2be78762d41f24231c83067d19d8f505d0c3d4",
	// 	"50dd3df9c5fa56785373b85e4121adccc9b9a849",
	// 	"940c21f904885cdd6047a1052f349ac191340991",
	// 	"a8bbe8aef203b33ed33d82df1a81e9adb26a9842",
	// 	"272686d1a53fadb342181e631982c0e8ce4dce6a", //new: part 1 of large file
	// 	"108a4ba1f91903cd6d1d2d0ac6f95fde9f58b39e", //new: new file node
	// 	"1dbb3d9950363dbb7ec8b14abe052ed5e5202628", //new: root node
	// }
	// for _, k := range exp {
	// 	if _, ok := nodes[k]; !ok {
	// 		assert.Fail(t, "Expected node with key "+k)
	// 	}
	// }

	//
	// @TODO How do we implement a rescan of a subdirectory
	//

	// //rescanning a change to a subdirectory should send new upper nodes
	// //as well
	// writeFileAt(t, filepath.Join(dir, "a", "b", "small_file"), []byte("foo"), 0)
	// go func() {
	// 	for n := range s.Nodes {
	// 		nodes[fmt.Sprintf("%x", n.Key())] = n
	// 		if len(nodes) == 14 {
	// 			done <- true
	// 			return
	// 		}
	// 	}
	// }()

	// err = s.Rescan(filepath.Join(dir, "a", "b"))
	// assert.NoError(t, err, "Rescanning the filesystem should not fail")
	// <-done

}
