package fotki

import (
	"os"
)


type ImageInfo struct {
    date ImageDate
	info os.FileInfo
}


func (self ImageInfo) String() string {
    return self.date.String()
}