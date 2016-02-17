package fotki

import (
	"fmt"
	"os"
)

type ImageInfo struct {
	path string
	date ImageDate
	kind ImageKind
	info os.FileInfo
}

func (self *ImageInfo) String() string {
	return fmt.Sprintf("%s %s", self.path, self.date.String())
}
