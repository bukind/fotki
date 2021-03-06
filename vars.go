package fotki

import (
	"errors"
	"fmt"
	"os"
)

// Print some verbose info if true.
var Verbose bool = false

// Do not modify filesystem if true.
var DryRun bool = false

// Do rescan of valid directories under root if true.
var Rescan bool = false

// Remove the original files
var RemoveOld bool = false

var SameFileError = errors.New("the same file as detected by OS")
var AlreadyExistError = errors.New("file already exists at dst")
var IdenticalError = errors.New("the files are identical")
var GarbageError = errors.New("not an image")

type ImageKind int

const (
	NoImage ImageKind = iota
	IsImage
	IsMovie
)

/// detect if the output is a tty, default is not
func IsTTY(fd *os.File) bool {
	if info, err := fd.Stat(); err != nil {
		fmt.Fprintf(os.Stderr, "cannot stat: %s\n", err.Error())
		return false
	} else if (info.Mode() & os.ModeDevice) != 0 {
		return true
	}
	return false
}
