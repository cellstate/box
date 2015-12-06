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

func TestScan(t *testing.T) {
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
			if len(nodes) == 20 {
				done <- true
				return
			}
		}
	}()

	err = s.Scan()
	assert.NoError(t, err, "Scanning the filesystem should not fail")

	<-done
	assert.Len(t, nodes, 20, "Expected N nodes after file edit")

	//we expect the following nodes from the first scan
	exp1 := []string{
		"05a82b3c331c68f097b0151dafe016ac6d7a05f0", // /
		"8b2a5d310e80ad144819786e36ca4733e26939c9", // +-/a
		"104c9da6a7654229304fd77f4479751070453613", //    +-/b
		"e025982956d87909188cd8b76699711478347de6", //       +-/small_file
		"a67316b4de11d37d722e7da5768d7d22220c2b89", //         +-.1 (last part)
		"3bd1b79ca92f423b6c54943e3ffc2f47bd349059", // +-/large_file
		"7b2f18510ef8919ec40a1084b4215e5008fdcd63", //   +-.1
		"a9713219f8ddbb7b07102798812456b659c7dd8b", //     +-.1
		"c8f9d95c38323d0ced6d6f4347aeb216bd952fc3", //     +-.2
		"68b4872687de1307293529b54b1304d8bc9752f3", //     +-.3
		"4e510ea21ba1686b3cae35ac4d70b146acb9ecea", //     +-.4
		"10e127809e3a524df036eff4a57a1f8f0a487f1f", //     +-.5
		"0c2be78762d41f24231c83067d19d8f505d0c3d4", //     +-.6
		"09caed6d1d68a045d4ac8900fd2c342124adfa6c", //   +-.2
		"b153af3ed57a87fa17bcd96a537767cf9132ef40", //     +-.1
		"300317a09449cf1ce0064f6c6483a8ff8facf316", //       +-.1
		"5cc88fded228359d87580c1c1c5f9fcd5812dcf2", //       +-.2
		"d1871116f30bf0ae8e610dff3a94470dcce2f811", //   +-.3
		"717cbf439a773491fda1d80784a93bd87f3973c0", //   +-.4
		"2a8b53c4798a72b2f84156031d281fbe2578902b", //   +-.5 (last part)
	}
	for _, k := range exp1 {
		if _, ok := nodes[k]; !ok {
			assert.Fail(t, "Expected node with key "+k)
		}
	}

	//we rewrite bytes at the beginning of the large file
	writeFileAt(t, filepath.Join(dir, "large_file"), []byte("foobbb"), 0)
	fi, err := os.Stat(filepath.Join(dir, "large_file"))
	assert.NoError(t, err, "large_file should be statable")
	assert.Equal(t, int64(100000), fi.Size(), "large_file should have remained the same size")

	go func() {
		for n := range s.Nodes {
			nodes[fmt.Sprintf("%x", n.Key())] = n
			if len(nodes) == 24 {
				done <- true
				return
			}
		}
	}()

	err = s.Scan()
	assert.NoError(t, err, "Rescanning the filesystem should not fail")
	<-done
	assert.Len(t, nodes, 24, "Expected N nodes after file edit")

	//a rescan should yield only 4 *new* nodes (1 new root, 1 new file and 2 new (sub) parts)
	exp2 := make([]string, len(exp1))
	copy(exp2, exp1)
	exp2 = append(exp2, []string{
		"a52e66bf904287782b8f9edd1f0da2914ae4d533", // /
		"3507fd760c8e0f71ae189bf815ac6287198fbfc3", // /large_file
		"c8c375b27bc367854fc937527ef9d228350e6e31", // /large_file.1
		"7e7ff5e5deed0f08cd4d38e866233b6d65f321fc", // /large_file.1.1
	}...)

	for _, k := range exp2 {
		if _, ok := nodes[k]; !ok {
			assert.Fail(t, "Expected node with key "+k)
		}
	}

	//
	// @TODO How do we implement a rescan of a subdirectory
	//

}
