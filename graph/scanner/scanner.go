package fsreverse

import (
	"bufio"
	"crypto/sha1"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/cellstate/box/graph"
	"github.com/cellstate/errwrap"
)

// @todo make sure new (upper) nodes are also send when lower directories are rescanned
// @todo switch to sha256 for better crypto

// represents a directory, a directory
// contains 0-N items (dirs or files) it
// is assumed that the order of items
// is the same between scans to yield the
// same hash unless some underlying item
// changed
type Dir struct {
	hash  []byte
	items [][]byte
}

func (d *Dir) calcHash() error {
	sha := sha1.New()
	for _, item := range d.items {
		_, err := sha.Write(item)
		if err != nil {
			return errwrap.Wrapf("Failed to write item hash '%x' to dir hash: {{err}}", err, item)
		}
	}

	d.hash = sha.Sum(nil)
	return nil
}

func (dir *Dir) Key() graph.Key        { return graph.Key(dir.hash) }
func (dir *Dir) Data() ([]byte, error) { return []byte{}, nil }
func (dir *Dir) Links() ([]graph.Key, error) {
	keys := []graph.Key{}

	for _, item := range dir.items {
		keys = append(keys, item)
	}

	return keys, nil
}

// represents a file, files contain
// 1-N parts, it is assumed that the order
// of parts is the same between scans so scans
// yield the same hash unless an underlying part
// changed
type File struct {
	hash  []byte
	parts [][]byte
}

func (f *File) calcHash() error {
	if len(f.parts) < 1 {
		return errors.New("File should have at least one part, got < 1")
	}

	sha := sha1.New()
	for _, p := range f.parts {
		_, err := sha.Write(p)
		if err != nil {
			return errwrap.Wrapf("Failed to write part hash '%x' to file hash: {{err}}", err, p)
		}
	}

	f.hash = sha.Sum(nil)
	return nil
}

func (f *File) Key() graph.Key        { return graph.Key(f.hash) }
func (f *File) Data() ([]byte, error) { return []byte{}, nil }
func (f *File) Links() ([]graph.Key, error) {
	keys := []graph.Key{}
	for _, p := range f.parts {
		keys = append(keys, p)
	}

	return keys, nil
}

// represents a file part
type Part struct {
	hash  []byte
	start int64
	end   int64
	bits  int
}

func (p *Part) Key() graph.Key              { return graph.Key(p.hash) }
func (p *Part) Links() ([]graph.Key, error) { return []graph.Key{}, nil }
func (p *Part) Data() ([]byte, error)       { return []byte{}, nil }

// A scanner for a specific directory, repeated
// scans of the same root are deterministic. resulting
// structure is send as graph nodes over a channel.
type Scanner struct {
	Nodes chan graph.Node

	root string
	*log.Logger
}

func NewScanner(l *log.Logger, root string) (*Scanner, error) {
	return &Scanner{
		root: root,

		Nodes:  make(chan graph.Node),
		Logger: l,
	}, nil
}

func (s *Scanner) SplitFile(p string, fi os.FileInfo) ([]*Part, error) {
	parts := []*Part{}

	f, err := os.Open(p)
	if err != nil {
		return parts, errwrap.Wrapf("Failed to open file '%s' for splitting: {{err}}", err, p)
	}

	defer f.Close()

	//buffer and sum each byte
	sha := sha1.New()
	buff := bufio.NewReader(f)
	rs := NewRollsum()
	pos := int64(0)
	last := pos
	for {
		b, err := buff.ReadByte()
		if err != nil {
			if err == io.EOF {
				if pos != last {
					parts = append(parts, &Part{start: last, end: pos, hash: sha.Sum(nil)})
				}
				break
			} else {
				return parts, errwrap.Wrapf("Failed to read byte from file '%s': {{err}}", err, p)
			}
		}

		_, err = sha.Write([]byte{b})
		if err != nil {
			return parts, errwrap.Wrapf("Failed to write byte to hash: {{err}}", err)
		}

		pos++

		rs.Roll(b)
		if rs.OnSplit() {
			parts = append(parts, &Part{bits: rs.Bits(), start: last, end: pos, hash: sha.Sum(nil)})
			sha.Reset()
			last = pos
		}
	}

	return parts, nil
}

// (re)scan the root directory, recursively calling
// scanDir(), depth-first
func (s *Scanner) Scan() error {
	_, err := s.scanDir(s.root)
	if err != nil {
		return err
	}

	return nil
}

//recursively scans a given directory depth-first
//in an memory efficient manner.
func (s *Scanner) scanDir(dirp string) (*Dir, error) {
	dirh, err := os.Open(dirp)
	if err != nil {
		return nil, err
	}

	defer dirh.Close()
	names, err := dirh.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	//sort names to make dir hashes consistent between scans
	sort.Strings(names)

	//stat stuff in directory
	dir := &Dir{}
	for _, n := range names {
		path := filepath.Join(dirh.Name(), n)

		//@todo check index for rapid rescans

		fi, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}

		if fi.IsDir() {
			//it is a dir, traverse
			ndir, err := s.scanDir(path)
			if err != nil {
				return nil, err
			}

			dir.items = append(dir.items, ndir.hash)

		} else {
			//it is a file, split into parts and create file
			parts, err := s.SplitFile(path, fi)
			if err != nil {
				return nil, err
			}

			f := &File{}
			for i, part := range parts {
				f.parts = append(f.parts, part.hash)

				s.Printf("found part %d of %s (%x)", i, path, part.hash)
				s.Nodes <- part
			}

			err = f.calcHash()
			if err != nil {
				return nil, err
			}

			s.Printf("%d parts of file %s (%x)", len(f.parts), path, f.hash)
			s.Nodes <- f

			dir.items = append(dir.items, f.hash)
		}
	}

	err = dir.calcHash()
	if err != nil {
		return nil, err
	}

	s.Printf("%d items in dir '%s', (%x)", len(dir.items), dirp, dir.hash)
	s.Nodes <- dir
	return dir, nil
}
