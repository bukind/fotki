package fotki

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

var minstamp time.Time
var maxstamp time.Time

func init() {
	minstamp = time.Unix(1262264400, 0) // 2010-01-01
	maxstamp = time.Unix(1483189200, 0) // 2017-01-01
}

type Album struct {
	root   string
	images []*ImageInfo          // good image -> their infos
	failed map[string]error      // failed image -> error
	years  map[int]YearDayKeeper // year -> contents
}

// Creates a new empty Album based in the rootdir.
func NewAlbum(rootdir string) *Album {
	self := new(Album)
	self.root = rootdir
	// self.images = make([]*ImageInfo,0)
	self.failed = make(map[string]error)
	self.years = make(map[int]YearDayKeeper)
	return self
}

// String returns the representation of the Album.
func (self *Album) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "root=%s\n", self.root)
	for _, info := range self.images {
		fmt.Fprintf(buf, "%s\n", info.String())
	}
	for img, err := range self.failed {
		fmt.Fprintf(buf, " %s => Error %s\n", img, err.Error())
	}
	for _, data := range self.years {
		fmt.Fprintln(buf, data.String())
	}
	return buf.String()
}

// Scan performs the deep search in scandir to find images/movies.
func (self *Album) Scan(scandir string) error {

	istty := IsTTY(os.Stdout)

	var imagelist []*ImageLoc
	walkFun := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			self.failed[path] = err
			return nil
		}
		if !info.Mode().IsRegular() {
			// we are only interested in the regular files
			return nil
		}
		kind := GetImageKind(info.Name())
		switch kind {
		case NoImage:
			self.failed[path] = Garbage
		default:
			imagelist = append(imagelist, &ImageLoc{path, info, kind})
		}
		return nil
	}

	if err := filepath.Walk(scandir, walkFun); err != nil {
		return err
	}

	type Result struct {
		info *ImageInfo
		err  error
	}

	done := make(chan int) // is not used for now
	feed := make(chan *ImageLoc)
	resc := make(chan Result)

	const ScannerCount = 8
	var wg sync.WaitGroup
	for i := 0; i < ScannerCount; i++ {

		// create a scanning goroutine
		wg.Add(1)
		go func(done <-chan int, feed <-chan *ImageLoc, out chan<- Result) {
			defer wg.Done()
			for image := range feed {
				info, err := image.ExtractDate()
				select {
				case <-done:
					// cancelled
					return
				case resc <- Result{info, err}:
					// continue
				}
			}
		}(done, feed, resc)

	}

	// start a consumer goroutine
	wg.Add(1)
	go func(done <-chan int, resc <-chan Result, total int) {
		defer wg.Done()
		for count := 0; count < total; count++ {
			select {
			case <-done:
				break
			case res := <-resc:
				if istty {
					fmt.Printf("\rScanned %d/%d", count+1, total)
				} else {
					if (count+1)%50 == 0 {
						fmt.Printf("Scanned %d/%d\n", count+1, total)
					}
				}
				if res.err == nil {
					self.images = append(self.images, res.info)
				} else {
					self.failed[res.info.path] = res.err
				}
			}
		}
		fmt.Printf("\n")
	}(done, resc, len(imagelist))

	// send paths to the scanners
	for _, image := range imagelist {
		feed <- image
	}
	close(feed)
	wg.Wait()
	close(resc)
	close(done)

	// load years
	yearmap := make(map[int]bool)
	for _, info := range self.images {
		yearmap[info.date.year] = true
	}
	for year, _ := range yearmap {
		if _, ok := self.years[year]; !ok {
			if ydir, err := MakeYearDays(self.root, year); err != nil {
				return err
			} else {
				self.years[year] = ydir
			}
		}
	}

	return nil
}

// Relocate all found images/movies to their canonical place.
func (self *Album) Relocate() error {
	for _, info := range self.images {

		year := self.years[info.date.year]
		if year == nil {
			fmt.Fprintf(os.Stderr, "year %d is not setup\n", info.date.year)
			os.Exit(1)
		}

		dstfiles, err := year.Adopt(info)
		if err != nil {
			self.failed[info.path] = err
			continue
		} else if len(dstfiles) == 0 {
			continue
		}

		for _, dst := range dstfiles {
			if err := self.linkImage(info.path, dst); err != nil {
				self.failed[info.path] = err
			}
		}
	}

	// normalize years
	for _, year := range self.years {
		if err := year.NormalizeDirs(); err != nil {
			return err
		}
	}
	return nil
}

func (self *Album) linkImage(src string, dst string) error {
	if Verbose {
		fmt.Println("# linking", src, "from", dst)
	}
	if DryRun {
		return nil
	}
	return os.Link(src, dst)
}

func (self *Album) ShowFailed() {
	if len(self.failed) == 0 {
		return
	}
	fails := make([]string, 0, len(self.failed))
	for key, _ := range self.failed {
		fails = append(fails, key)
	}
	sort.Strings(fails)
	fmt.Fprintf(os.Stderr, "The following %d files were not processed\n", len(fails))
	for _, key := range fails {
		fmt.Fprintln(os.Stderr, "FAIL", key, ":", self.failed[key])
	}
}
