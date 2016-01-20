package fotki

import (
    "fmt"
)

type Directory struct {
    items map[string]bool
}


func NewDirectory() *Directory {
    self := new(Directory)
    self.items = make(map[string]bool)
    return self
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
