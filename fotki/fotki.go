package main

import (
	"flag"
	"fmt"
	"github.com/bukind/fotki"
	"os"
)

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail: %s\n", err.Error())
		os.Exit(1)
	}
}

func main() {
	flag.BoolVar(&fotki.Verbose, "v", false, "Be verbose")
	flag.BoolVar(&fotki.DryRun, "n", false, "Dry-run")
	scandir := flag.String("scan", "", "The directory to scan")
	rootdir := flag.String("root", "", "The root directory of the album")

	flag.Parse()

	if fotki.Verbose {
		fmt.Println("# scandir=", *scandir)
		fmt.Println("# rootdir=", *rootdir)
	}

	if *scandir == "" {
		fmt.Fprintln(os.Stderr, "scan flag is required")
		os.Exit(1)
	}

	if *rootdir == "" {
		fmt.Fprintln(os.Stderr, "root flag is required")
		os.Exit(1)
	}

	album := fotki.NewAlbum(*rootdir)
	check(album.Scan(*scandir))

	check(album.Relocate())
	album.ShowFailed()
}
