package fotki

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Directory struct {
	path  string          // absolute path of the dir
	items map[string]os.FileInfo // relative paths of the contents
}

func makePath(elts ...string) string {
	return filepath.Join(elts...)
}

// NewDirectory creates a Directory from the list of paths.
func NewDirectory(path ...string) *Directory {
	self := new(Directory)
	self.path = makePath(path...)
	self.items = make(map[string]os.FileInfo)
	return self
}

// Path returns the path of the directory joined with optional arguments.
func (self *Directory) Path(path ...string) string {
	if len(path) > 0 {
		full := make([]string, len(path)+1)
		full[0] = self.path
		copy(full[1:], path)
		return makePath(full...)
	}
	return self.path
}

func (self *Directory) contents() []string {
	result := make([]string, 0, len(self.items))
	for key := range self.items {
		result = append(result, key)
	}
	return result
}

func (self *Directory) String() string {
	return fmt.Sprintf("%v", self.contents())
}

// Add an item w/o creating it in the directory.
func (self *Directory) Add(item string, info os.FileInfo) {
	self.items[item] = info
}

// Check if the item exists in the directory.
// The item may not exist on disk yet.
func (self *Directory) Has(item string) bool {
	_, ok := self.items[item]
	return ok
}

func (self *Directory) Stat(item string) (os.FileInfo, error) {
    if fi, ok := self.items[item]; !ok {
		return nil, &os.PathError{Op: "Stat", Path: self.Path(item), Err: errors.New("not found")}
	} else {
	    if fi == nil {
		    var err error
		    if fi, err = os.Stat(self.Path(item)); err != nil {
			    fmt.Println("# negative hit Stat:", self.Path(item))
			    return nil, err
			}
			self.items[item] = fi
		}
	    return fi, nil
	}
}
