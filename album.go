package fotki

import (
    "bytes"
    "fmt"
)

type Album struct {
    root string
		images map[string]ImageDate // good image -> their dates
		failed map[string]error     // failed image -> error
}


func NewAlbum(rootdir string) *Album {
    self := new(Album)
		self.root = rootdir
		self.images = make(map[string]ImageDate)
		self.failed = make(map[string]error)
		return self
}


func (self *Album) String() string {
    buf := new(bytes.Buffer)
    fmt.Fprintf(buf, "root=%s\n", self.root)
    for img, date := range self.images {
        fmt.Fprintf(buf, " %s => %s\n", img, date.String())
    }
    for img, err := range self.failed {
        fmt.Fprintf(buf, "\n %s => Error %s\n", img, err.Error())
    }
    return buf.String()
}


func (self *Album) Scan(scandir string) error {
    _ = scandir
    return nil
}
