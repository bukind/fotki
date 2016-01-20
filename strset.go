package fotki

import "fmt"

type StrSet map[string]bool

func NewStrSet() *StrSet {
    self := new(StrSet)
    *self = make(map[string]bool)
    return self
}


func (self *StrSet) Keys() []string {
    result := make([]string, 0, len(*self))
    for key := range *self {
        result = append(result,key)
    }
    return result
}


func (self *StrSet) Len() int {
    return len(*self)
}


func (self *StrSet) String() string {
    return fmt.Sprintf("%v", self.Keys())
}


func (self *StrSet) Add(item string) {
    (*self)[item] = true
}
