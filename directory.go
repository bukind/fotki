package fotki

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Directory struct {
	path  string          // absolute path of the dir
	items map[string]bool // relative paths of the contents
}

func makePath(elts ...string) string {
	return filepath.Join(elts...)
}

// NewDirectory creates a Directory from the list of paths.
func NewDirectory(path ...string) *Directory {
	self := new(Directory)
	self.path = makePath(path...)
	self.items = make(map[string]bool)
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

func (self *Directory) Contents() []string {
	result := make([]string, 0, len(self.items))
	for key := range self.items {
		result = append(result, key)
	}
	return result
}

func (self *Directory) String() string {
	return fmt.Sprintf("%v", self.Contents())
}

func (self *Directory) Add(item string) {
	self.items[item] = true
}

func (self *Directory) Has(item string) bool {
	_, ok := self.items[item]
	return ok
}

func (self *Directory) Stat(item string) (os.FileInfo, error) {
	if !self.Has(item) {
		return nil, &os.PathError{Op: "Stat", Path: self.Path(item), Err: errors.New("not found")}
	}
	return os.Stat(self.Path(item))
}
