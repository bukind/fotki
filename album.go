package fotki

import (
    "bufio"
    "bytes"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "sync"
    "time"
)

var minstamp time.Time
var maxstamp time.Time


func init() {
    minstamp = time.Unix(1262264400, 0)  // 2010-01-01
    maxstamp = time.Unix(1483189200, 0)  // 2017-01-01
}


type Album struct {
    root string
    images map[string]ImageInfo // good image -> their infos
    failed map[string]error     // failed image -> error
    years  map[int]*YearDays    // year -> contents
}


func NewAlbum(rootdir string) *Album {
    self := new(Album)
    self.root = rootdir
    self.images = make(map[string]ImageInfo)
    self.failed = make(map[string]error)
    self.years  = make(map[int]*YearDays)
    return self
}


func (self *Album) String() string {
    buf := new(bytes.Buffer)
    fmt.Fprintf(buf, "root=%s\n", self.root)
    for img, info := range self.images {
        fmt.Fprintf(buf, " %s => %s\n", img, info.String())
    }
    for img, err := range self.failed {
        fmt.Fprintf(buf, " %s => Error %s\n", img, err.Error())
    }
    for _, data := range self.years {
        fmt.Fprintln(buf, data.String())
    }
    return buf.String()
}


func (self *Album) Scan(scandir string) error {

    istty := IsTTY(os.Stdout)

    type Image struct {
        path string
        info os.FileInfo
        kind ImageKind
    }
    var imagelist []Image
    walkFun := func (path string, info os.FileInfo, err error) error {
        if err != nil {
            self.failed[path] = err
            return nil
        }
        if !info.Mode().IsRegular() {
            // we are only interested in the regular files
            return nil
        }
        filetype := GetImageKind(info.Name())
        switch filetype {
            case NoImage:
                self.failed[path] = Garbage
            default:
                imagelist = append(imagelist, Image{path, info, filetype})
        }
        return nil
    }

    if err := filepath.Walk(scandir, walkFun); err != nil {
        return err
    }

    type Result struct {
        path string
        info ImageInfo
        err  error
    }

    done := make(chan int) // is not used for now
    feed := make(chan Image)
    resc := make(chan Result)

    const ScannerCount = 8
    var wg sync.WaitGroup
    for i := 0; i < ScannerCount; i++ {

        // create a scanning goroutine
        wg.Add(1)
        go func(done <-chan int, feed <-chan Image, out chan<-Result) {
            defer wg.Done()
            for image := range feed {
                date, err := self.ExtractImageDate(image.path)
                select {
                case <-done:
                    // cancelled
                    return
                case resc <- Result{image.path,
                                    ImageInfo{date,
                                              image.kind,
                                              image.info}, err}:
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
                    if (count+1) % 50 == 0 {
                        fmt.Printf("Scanned %d/%d\n", count+1, total)
                    }
                }
                if res.err == nil {
                    self.images[res.path] = res.info
                } else {
                    self.failed[res.path] = res.err
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
            ydir := NewYearDays(self, year)
            self.years[year] = ydir
            if err := ydir.Scan(); err != nil {
                return err
            }
        }
    }

    return nil
}


func (self *Album) ExtractImageDate(path string) (ImageDate, error) {
    out, err := exec.Command("identify", "-ping", "-verbose", path).Output()
    ret := ImageDate{}
    if err == nil {
        // try to extract the date from the path itself
        reg := regexp.MustCompile(`(20[0123]\d)[-_]?(0[1-9]|1[012])[-_]?(0[1-9]|[123]\d)[^/]*$`)
        // tag := "file"
        groups := reg.FindStringSubmatch(path)
        if groups != nil {
            var year, month, day int
            year, err = strconv.Atoi(groups[1])
            if err == nil {
                month, err = strconv.Atoi(groups[2])
                if err == nil {
                    day, err = strconv.Atoi(groups[3])
                }
            }
            if err == nil {
                ret.year = year
                ret.month = month
                ret.day = day
                // fmt.Printf("%s: %s\n", tag, ret)
            }
        }
        var value string
        scanner := bufio.NewScanner(bytes.NewBuffer(out))
        for scanner.Scan() {
            str := scanner.Text()
            switch {
                // disable modification time for now
                // case strings.HasPrefix(str, "    date:modify: "):
                // tag = "date"
                // value = strings.SplitN(str, ": ", 2)[1]
                case strings.HasPrefix(str, "    exif:DateTimeOriginal: "):
                    // tag = "exif"
                    value = strings.SplitN(str, ": ", 2)[1]
                default:
                    continue
            }
            var dummy int
            fmt.Sscanf(value, "%04d%c%02d%c%02d", &ret.year, &dummy, &ret.month, &dummy, &ret.day)
            // fmt.Printf("%s: %s\n", tag, ret)
        }
        err = scanner.Err()
    }
    if ret.IsEmpty() {
        // try to convert the path to a timestamp
        err = fmt.Errorf("cannot detect the date")
        fp := filepath.Base(path)
        fp = fp[:len(fp)-len(filepath.Ext(fp))]
        if stamp, err2 := strconv.ParseInt(fp, 10, 64); err2 == nil {
            const nms = 1000
            ts := time.Unix(stamp / nms, stamp % nms)
            if ts.After(minstamp) && ts.Before(maxstamp) {
                ret.year = ts.Year()
                ret.month = int(ts.Month())
                ret.day = ts.Day()
                err = nil
            }
        }
    }
    return ret, err
}


// relocate found images
func (self *Album) Relocate() error {
    for image, info := range self.images {
        // The image must be normalized to have only lowercase letters.
        // There must be two destination, per-day and per-month.
        // The per-day destination may have optional suffix:
        // YEARDIR/all/YYYY-MM/pic.jpg
        // YEARDIR/YYYY-MM-DD[-suffix]/pic.jpg

        srcdir, srcname := filepath.Split(image)
        dstname := strings.Replace(strings.ToLower(srcname), " ", "_", -1)

        if Verbose {
            fmt.Println("# processing", srcdir, srcname, info.date, "->", dstname)
        }

        dstdirs := make([]string,0)

        year := self.years[info.date.year]
        if year == nil {
            fmt.Fprintf(os.Stderr, "year %d is not setup\n", info.date.year)
            os.Exit(1)
        }
        var errx error
        if dstdir, err := year.FindMonth(info.date, dstname, info.info); err == nil {
            dstdirs = append(dstdirs, dstdir)
        } else if err == SameFile {
            if Verbose {
                fmt.Printf("# same files %s and %s\n", image, dstdir)
            }
            errx = nil
        } else {
            errx = err
        }
        if dstdir, err := year.FindDay(info.date, dstname, info.info); err == nil {
            dstdirs = append(dstdirs, dstdir)
        } else if err == SameFile {
            if Verbose {
                fmt.Printf("# same files %s and %s\n", image, dstdir)
            }
            errx = nil
        } else {
            errx = err
        }
        if len(dstdirs) == 0 {
            // both are failed
            if errx != nil {
                self.failed[image] = errx
            }
            continue
        }

        if err := year.MakeAllDirs(); err != nil {
            // failed to create dir
            return err
        }

        for _, dst := range dstdirs {
            if err := self.LinkImage(image, dst); err != nil {
                errx = err
            }
        }
        if errx != nil {
            self.failed[image] = errx
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


func (self *Album) LinkImage(src string, dst string) error {
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
