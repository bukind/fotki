package fotki

import (
	"os"
	"path/filepath"
	"strings"
)


type ImageKind int

const (
    NoImage ImageKind = iota
    IsImage
    IsMovie
)


func GetImageKind(path string) ImageKind {
    ext := strings.ToLower(filepath.Ext(path))
    switch ext {
        case ".jpg", ".png", ".jpeg":
            return IsImage
        case ".mp4", ".avi", ".thm":
            return IsMovie
        default:
            return NoImage
    }
}


type ImageInfo struct {
    date ImageDate
	kind ImageKind
	info os.FileInfo
}


func (self ImageInfo) String() string {
    return self.date.String()
}
