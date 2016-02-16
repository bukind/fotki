package fotki

import (
    "errors"
    "fmt"
    "os"
)

var Verbose bool = false
var DryRun bool = false

var SameFile = errors.New("the same file found")
var Garbage = errors.New("not an image")


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
