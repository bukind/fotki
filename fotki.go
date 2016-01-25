// +build ignore

package main

import (
    "fmt"
		"bukind/fotki"
		"os"
		"path/filepath"
)


func check(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "fail:", err.Error())
        os.Exit(1)
    }
}


func main() {
    // temporary hardcoded
		fotki.Verbose = true
    fotki.DryRun = true
    rootdir := filepath.Join(os.Getenv("HOME"), "Downloads")
    album := fotki.NewAlbum(rootdir)
    check(album.Scan(rootdir))

    check(album.Relocate())

    if fotki.Verbose {
        fmt.Println(album.String())
    }
}
