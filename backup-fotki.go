// Dmitry Bukin
// +build ignore

package main

import (
    "bufio"
    "bytes"
    "flag"
    "fmt"
    "fotki"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "sort"
    "strconv"
    "strings"
)


var verbose bool = false

func (self *YearAllDaysDir) Scan(perdaydir string) error {
    // nShards := 0
    // isReady := make(chan int)
    yearPerDayRegex := regexp.MustCompile(`^(20[0123]\d)-(0[1-9]|1[012])-(0[1-9]|[123]\d)`)

    if verbose {
        fmt.Println("# scanning", perdaydir)
    }

    // the function to scan
    collectPerDayWalk := func (path string, info os.FileInfo, err error) error {
        if err != nil {
            // failed - ignore
            return nil
        }
        if info.Mode().IsDir() {
            if verbose {
                fmt.Println("# dir",path)
            }
            match := yearPerDayRegex.FindStringSubmatch(info.Name())
            if match != nil {
                // a match found, conversion can be done w/o error checking
                var date ImageDate
                date.year, _ = strconv.Atoi(match[1])
                date.month, _ = strconv.Atoi(match[2])
                date.day, _ = strconv.Atoi(match[3])
                if _, ok := self.date2dir[date]; !ok {
                    self.date2dir[date] = new(StringSet)
                }
                self.alldirs[info.Name()] = NewDirectory()
                strset := self.date2dir[date]
                (*strset)[info.Name()] = true
                // TODO: start subwalk

            } else {
                // invalid subdirectory!
                // TODO: add it to the garbage
            }
            return filepath.SkipDir
        } else {
            // TODO: not a dir at the first level, add to the garbage
        }
        return nil
    }

    return filepath.Walk(perdaydir, collectPerDayWalk)
}


// ===============================================================
// check and add an image to the imagelist
// scandir the full path to the directory to scan
func (self *ImageList) Scan(scandir string) error {

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
                date, err := self.ExtractImageDate(path);
                if err == nil {
                    self.images[path] = date
                } else {
                    self.failed[path] = err
                }
        }
        return nil
    }

    return filepath.Walk(scandir, walkFun)
}


func (self *ImageList) FillAllYearCache() {
    year_set := make(map[int]int)
    for _, date := range self.images {
        year_set[date.year] = 1
    }
    if verbose {
        fmt.Println("all years:", year_set)
    }
    for year := range year_set {
        if verbose {
            fmt.Println("year:", year)
        }
        yearall := NewYearAllDaysDir()
        yearall.Scan(self.YearDir(year))
        if verbose {
            fmt.Println(yearall.String())
        }
    }
}


func (self *ImageList) ExtractImageDate(path string) (ImageDate, error) {
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


// to be replaced by os.Link
func os_Link(oldname, newname string) error {
    fmt.Printf("link %s <- %s\n", oldname, newname)
    return nil
}


func (self *ImageList) dummy_RelocateAll() {
    var tofail []string
next_image:
    for image, date := range self.images {
        srcdir, name := filepath.Split(image)
        fmt.Println(srcdir, name, date)
        yearstr := date.YearString()
        datestr := date.DayString()

        dstname := strings.Replace(strings.ToLower(name), " ", "_", -1)

        alldir := filepath.Join(self.root, yearstr, "all")
        // if _, ok := self.perday[alldir]; !ok {
            // we have not scanned the dir yet, do it now.
            // the directory is necessary to find the possible presence of image
            // fmt.Println(" checking", alldir)
            // self.perday[alldir] = NewSubdirSet()
            // if err := filepath.Walk(alldir, self.perday[alldir].collectSubdirWalk); err != nil {
            // fmt.Fprintln(os.Stderr, "  cannot scan", alldir, ":", err)
            // }
        // }

        // check if the image already exists in one of perday subdir
        // TODO: cache the contents of subdirs later
        perday_content := NewStringSet() // self.perday[alldir].subdirs[date]

        fmt.Printf(" perdays@%s:%s#%d: %v\n", alldir, datestr, perday_content.len(), perday_content.keys())
        for subdir := range *perday_content {
            tocheck := filepath.Join(alldir, subdir, dstname)
            fmt.Printf("  checking %s\n", tocheck)
            if _, err := os.Stat(tocheck); err == nil {
                // the file exists, skip creation
                self.failed[image] = fmt.Errorf("file exists at %s", tocheck)
                tofail = append(tofail, image)
                continue next_image
            }
        }

        // check if the image already exists in per month dir
        tocheck := filepath.Join(self.root, yearstr, date.MonthString(),
                                 dstname)
        if _, err := os.Stat(tocheck); err == nil {
            // the file exists, skip creation
            self.failed[image] = fmt.Errorf("file exists at %s", tocheck)
            tofail = append(tofail, image)
            continue next_image
        }

        // override the destination if needed
        dstdir := filepath.Join(alldir, datestr)
        if len(*perday_content) == 1 {
            // if exactly one directory exists for the given date, then use it
            keys := perday_content.keys()
            dstdir = filepath.Join(alldir, keys[0])
            fmt.Printf(" using single day dst %s\n", dstdir)
        }

        // link the original image from per day and per month location
        if err := os_Link(image, tocheck); err != nil {
            // cannot link from month
            self.failed[image] = fmt.Errorf("cannot link: %s", err.Error())
            tofail = append(tofail, image)
            continue next_image
        }

        tocheck = filepath.Join(dstdir, dstname)
        if err := os_Link(image, tocheck); err != nil {
            // cannot link from day
            self.failed[image] = fmt.Errorf("cannot link: %s", err.Error())
            tofail = append(tofail, image)
            continue next_image
        }
    }

    // remove elements that failed
    for _, image := range tofail {
        delete(self.images, image)
    }

    if len(self.failed) > 0 {
        fmt.Fprintln(os.Stderr, "The following images failed:")
        // print everything that failed
        keys := make([]string, 0, len(self.failed))
        for image := range self.failed {
            keys = append(keys, image)
        }
        sort.Strings(keys)
        for _, image := range keys {
            fmt.Fprintln(os.Stderr, image, self.failed[image].Error())
        }
    }
}


// A test of scanner
func _testScanner() {
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        fmt.Printf("<%s>\n", scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "error:", err)
        os.Exit(1)
    }
}


func main() {
    // get the scanroot directory
    flag.BoolVar(&verbose, "v", false, "Be verbose")
    scandir := flag.String("scan", "", "The directory to scan")
    rootdir := flag.String("root", "", "The root directory of the album")

    flag.Parse()

    if verbose {
        fmt.Println("scandir:", *scandir)
        fmt.Println("rootdir:", *rootdir)
    }

    // validate the input
    if *scandir == "" {
        fmt.Fprintln(os.Stderr, "scan flag is required")
        os.Exit(1)
    }
    if *rootdir == "" {
        fmt.Fprintln(os.Stderr, "root flag is required")
        os.Exit(1)
    }

    // scan the directory
    images := NewImageList(*rootdir)
    if err := images.Scan(*scandir); err != nil {
        // failed to collect
        fmt.Fprintln(os.Stderr, "failed to collect:", err)
        os.Exit(2)
    }

    if verbose {
        fmt.Println("the list of images follows:")
        fmt.Println(images.String())
    }

    images.FillAllYearCache()
}
