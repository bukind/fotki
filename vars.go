package fotki

import (
    "errors"
)

var Verbose bool = false
var DryRun bool = false

var SameFile = errors.New("the same file found")
var Garbage = errors.New("not an image")
