// +build ignore

package main

import (
    "fmt"
		"bukind/fotki"
		"os"
		"path/filepath"
)

func main() {
    // temporary hardcoded
		fotki.Verbose = true
    rootdir := filepath.Join(os.Getenv("HOME"), "Downloads")
    album := fotki.NewAlbum(rootdir)
		album.Scan(rootdir)
		fmt.Println(album.String())
}
