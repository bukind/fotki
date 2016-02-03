package fotki

import (
    "bufio"
    "bytes"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "strconv"
    "strings"
)

type Album struct {
    root string
    images map[string]ImageDate // good image -> their dates
    failed map[string]error     // failed image -> error
    years  map[int]*YearDays    // year -> contents
}


func NewAlbum(rootdir string) *Album {
    self := new(Album)
    self.root = rootdir
    self.images = make(map[string]ImageDate)
    self.failed = make(map[string]error)
    self.years  = make(map[int]*YearDays)
    return self
}


func (self *Album) String() string {
    buf := new(bytes.Buffer)
    fmt.Fprintf(buf, "root=%s\n", self.root)
    for img, date := range self.images {
        fmt.Fprintf(buf, " %s => %s\n", img, date.String())
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

    var images []string
    walkFun := func (path string, info os.FileInfo, err error) error {
        if !info.Mode().IsRegular() {
            // we are only interested in the regular files
            return nil
        }
        if err != nil {
            self.failed[path] = err
            return nil
        }
        ext := strings.ToLower(filepath.Ext(info.Name()))
        switch ext {
            case ".jpg", ".png", ".jpeg":
                images = append(images, path)
            // TODO: should we add something to garbage here
        }
        return nil
    }

    if err := filepath.Walk(scandir, walkFun); err != nil {
        return err
    }

    for i, path := range images {
         // scan is going to be here
         if Verbose {
             fmt.Printf("# scanning %d/%d %s\n", i, len(images), path)
         }
         date, err := self.ExtractImageDate(path)
         if err == nil {
              self.images[path] = date
         } else {
              self.failed[path] = err
         }
    }

    // load years
    yearmap := make(map[int]bool)
    for _, d := range self.images {
        yearmap[d.year] = true
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
        err = fmt.Errorf("cannot detect the date")
    }
    return ret, err
}


// relocate found images
func (self *Album) Relocate() error {
    for image, date := range self.images {
        // The image must be normalized to have only lowercase letters.
        // There must be two destination, per-day and per-month.
        // The per-day destination may have optional suffix:
        // YEARDIR/YYYY-MM/pic.jpg
        // YEARDIR/all/YYYY-MM-DD[-suffix]/pic.jpg

        srcdir, srcname := filepath.Split(image)
        dstname := strings.Replace(strings.ToLower(srcname), " ", "_", -1)

        if Verbose {
            fmt.Println("# processing", srcdir, srcname, date, "->", dstname)
        }

        dstdirs := make([]string,0)

        year := self.years[date.year]
        if year == nil {
            fmt.Fprintf(os.Stderr, "year %d is not setup\n", date.year)
            os.Exit(1)
        }
        var errx error
        if dstdir, err := year.FindMonth(date, dstname); err == nil {
            dstdirs = append(dstdirs, dstdir)
        } else {
            errx = err
        }
        if dstdir, err := year.FindDay(date, dstname); err == nil {
            dstdirs = append(dstdirs, dstdir)
        } else {
            errx = err
        }
        if len(dstdirs) == 0 && errx != nil {
            // both are failed
            self.failed[image] = errx
            continue
        }

        errx = nil
        for _, dst := range dstdirs {
            if err := self.MoveImage(image, dst); err != nil {
                errx = err
            }
        }
        if errx != nil {
            self.failed[image] = errx
        }
    }
    return nil
}


func (self *Album) MoveImage(src string, dst string) error {
    if Verbose {
        fmt.Println("# moving", src, "to", dst)
    }
    if DryRun {
        return nil
    }
    return os.Rename(src, dst)
}
