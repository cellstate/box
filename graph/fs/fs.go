package fs

import (
	"bufio"
	"crypto/sha1"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/cellstate/box/config"
	"github.com/cellstate/box/graph"
	"github.com/cellstate/errwrap"
)

var LargeFileLimit = int64(5000)

// generate a graph from a file system
func FromFS(p string, l *log.Logger) (*FS, error) {
	fs := &FS{
		dirs:   map[string]*FSDir{},
		root:   p,
		Logger: l,
	}

	err := filepath.Walk(p, fs.walk)
	if err != nil {
		return nil, errwrap.Wrapf("Failed to walk fs at '%s': {{err}}", err, p)
	}

	return fs, nil
}

// represents a file system directory, dir
// structures typically have no data but (many)
// links
type FSDir struct {
	links []graph.Node
}

func (dir *FSDir) Key() (graph.Key, error) {
	return graph.Key(""), nil
}

func (dir *FSDir) Metadata() map[string]string {
	return map[string]string{}
}

func (dir *FSDir) Links() []graph.Key {
	return []graph.Key{}
}

func (dir *FSDir) Link(n graph.Node) error {
	if dir, ok := n.(*FSDir); ok {
		dir.links = append(dir.links, dir)
	} else if f, ok := n.(*FSDir); ok {
		dir.links = append(dir.links, f)
	} else {
		return errors.New("Cannot link from dir to type '%T'")
	}

	return nil
}

func (dir *FSDir) Data() ([]byte, error) {
	return []byte{}, nil
}

// represents a file system file, files may have
// links to other files in case of a large file
// that is chucked
type FSFile struct {
	links []graph.Node
}

func (f *FSFile) Key() (graph.Key, error) {
	return graph.Key(""), nil
}

func (f *FSFile) Metadata() map[string]string {
	return map[string]string{}
}

func (f *FSFile) Links() []graph.Key {
	return []graph.Key{}
}

func (f *FSFile) Link(n graph.Node) error {
	if part, ok := n.(*FSPart); ok {
		f.links = append(f.links, part)
	} else {
		return errors.New("Cannot link from file to type '%T'")
	}

	return nil
}

func (f *FSFile) Data() ([]byte, error) {
	return []byte{}, nil
}

// represents a file part, used to split
// up large files using a rolling checksum
// should not contain links
type FSPart struct {
	hash  []byte
	start int64
	end   int64
	bits  int
}

func (p *FSPart) Key() (graph.Key, error) {
	return graph.Key(""), nil
}

func (p *FSPart) Metadata() map[string]string {
	return map[string]string{}
}

func (p *FSPart) Links() []graph.Key {
	return []graph.Key{}
}

func (p *FSPart) Link(n graph.Node) error {
	return errors.New("Cannot link from part to other nodes")
}

func (p *FSPart) Data() ([]byte, error) {
	return []byte{}, nil
}

// represents a file system as a
// (DA) graph. node keys are hashes that can
// be used for efficient comparison using
// a merkle tree. Large files are split up
// using rsync link rolling hash
type FS struct {
	root  string
	dirs  map[string]*FSDir
	nodes []graph.Node
	*log.Logger
}

func (fs *FS) rel(p string) string {
	res, err := filepath.Rel(fs.root, p)
	if err != nil {
		fs.Printf("Warning: Failed to determine relative path from root: '%s' to target: '%s': %s", fs.root, p, err)
		return fs.root
	}

	return res
}

//use an adler-32 rolling checksum to do content-based
//splitting of a (large) file. when we are reading bytes from disk
//we can hash them with a more secure hash as well
//
//@todo this is called during stat traversal, is that the right time?
func (fs *FS) split(fsf *FSFile, p string, fi os.FileInfo) ([]*FSPart, error) {
	parts := []*FSPart{}

	f, err := os.Open(p)
	if err != nil {
		return parts, errwrap.Wrapf("Failed to open file '%s' for splitting: {{err}}", err, fs.rel(p))
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
					parts = append(parts, &FSPart{start: last, end: pos, hash: sha.Sum(nil)})
				}
				break
			} else {
				return parts, errwrap.Wrapf("Failed to read byte from file '%s': {{err}}", err, fs.rel(p))
			}
		}

		_, err = sha.Write([]byte{b})
		if err != nil {
			return parts, errwrap.Wrapf("Failed to write byte to hash: {{err}}", err)
		}

		pos++

		rs.Roll(b)
		if rs.OnSplit() {
			parts = append(parts, &FSPart{bits: rs.Bits(), start: last, end: pos, hash: sha.Sum(nil)})
			sha.Reset()
			last = pos
		}
	}

	return parts, nil
}

//@todo, this implementation can be pretty heavy on memory
//we keep track of all directories in .dirs and also stack all
//nodes of the graph in .nodes
func (fs *FS) walk(p string, fi os.FileInfo, err error) error {

	//do not plot the .box directory
	if fs.rel(p) == config.BoxDirName {
		return filepath.SkipDir
	}

	var n graph.Node
	relp := fs.rel(p)
	parentp := filepath.Base(relp)
	if fi.IsDir() {
		fs.Printf("Plotting dir %s", relp)
		dir := &FSDir{}
		n = dir

		//are we inside another node we want to link from
		if parent, ok := fs.dirs[parentp]; ok {
			err := parent.Link(n)
			if err != nil {
				return errwrap.Wrapf("Could not link dir '%s' from '%s': {{err}}", err, relp, parentp)
			}
		}

		fs.dirs[relp] = dir
	} else {
		fs.Printf("Plotting file %s", fs.rel(p))
		f := &FSFile{}
		n = f

		if parent, ok := fs.dirs[parentp]; ok {
			err := parent.Link(n)
			if err != nil {
				return errwrap.Wrapf("Could not link file '%s' from '%s': {{err}}", err, relp, parentp)
			}
		}

		if fi.Size() > LargeFileLimit {
			fs.Printf("File '%s' is considered large (%d > %d), splitted up in parts", relp, fi.Size(), LargeFileLimit)
			parts, err := fs.split(f, p, fi)
			if err != nil {
				return errwrap.Wrapf("Failed to split file '%s': {{err}}", err, relp)
			}

			for i, p := range parts {
				err := f.Link(p)
				if err != nil {
					return errwrap.Wrapf("Could not link part '%d' from '%s': {{err}}", err, i, relp)
				}

				fs.Printf("Linked part %d starts at byte %d ends at %d, hash: (%x)", i, p.start, p.end, p.hash)
			}
		}
	}

	fs.nodes = append(fs.nodes, n)
	return nil
}

func (fs *FS) Compare(b graph.Graph) ([]graph.Node, error) {
	return []graph.Node{}, nil
}

func (fs *FS) List() ([]graph.Node, error) {
	return []graph.Node{}, nil
}
