package fotki

import (
    "bytes"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strconv"
)


type YearDays struct {
    album    *Album  // backlink to album
    year     int
    daydirs  map[string]*Directory  // directory for the day
    day2dir  map[ImageDate]*StrSet  // mapping date -> set of dirs
    mon2dir  map[int]*Directory     // month -> directory
    garbage  *StrSet
    basedir  string
    tomake   []string               // list of directories to make
}


func NewYearDays(album *Album, year int) *YearDays {
    self := new(YearDays)
    self.album = album
    self.year = year
    self.daydirs = make(map[string]*Directory)
    self.day2dir = make(map[ImageDate]*StrSet)
    self.mon2dir = make(map[int]*Directory)
    self.garbage = NewStrSet()
    self.basedir = filepath.Join(self.album.root, strconv.Itoa(self.year))
    return self
}


func (self *YearDays) MakePath(elts ...string) string {
    s := []string{self.basedir}
    s = append(s, elts...)
    return filepath.Join(s...)
}


func (self *YearDays) String() string {
    buf := new(bytes.Buffer)
    fmt.Fprintf(buf, "year:%d => %s\n", self.year, self.MakePath("all"))
    for k, v := range self.day2dir {
        fmt.Fprintf(buf, " %s: %s\n", k.String(), v.String())
    }
    fmt.Fprintf(buf, " .garbage: %s\n", self.garbage.String())
    return buf.String()
}


// return the directory and the flag if it was just made
func (self *YearDays) get_mondir(month int) (*Directory, bool) {
    mondir := self.mon2dir[month]
    justmade := false
    if mondir == nil {
        mondir = NewDirectory()
        self.mon2dir[month] = mondir
        justmade = true
    }
    return mondir, justmade
}


func (self *YearDays) Scan() error {

    // scanning day dirs

    scandir := self.MakePath("all")
    yearPerDayRegex := regexp.MustCompile(`^(20[0123]\d)-(0[1-9]|1[012])-(0[1-9]|[123]\d)`)

    if Verbose {
        fmt.Println("# scanning", scandir)
    }

    // the function to scan
    collectPerDayWalk := func (path string, info os.FileInfo, err error) error {
        if err != nil {
            // failed - ignore
            return nil
        }
        if info.Mode().IsDir() {
            if path == scandir {
                // the scandir itself - ignore
                return nil
            }
            if Verbose {
                fmt.Println("# dir",path)
            }
            match := yearPerDayRegex.FindStringSubmatch(info.Name())
            if match != nil {
                // a match found, conversion can be done w/o error checking
                var date ImageDate
                date.year, _ = strconv.Atoi(match[1])
                date.month, _ = strconv.Atoi(match[2])
                date.day, _ = strconv.Atoi(match[3])
                if _, ok := self.day2dir[date]; !ok {
                    self.day2dir[date] = NewStrSet()
                }
                self.daydirs[info.Name()] = NewDirectory()
                strset := self.day2dir[date]
                strset.Add(info.Name())
                // TODO: start subwalk
            } else {
                // invalid subdirectory!
                // TODO: add it to the garbage
                self.garbage.Add(info.Name())
            }
            return filepath.SkipDir
        } else {
            // TODO: not a dir at the first level, add to the garbage
        }
        return nil
    }

    if err := filepath.Walk(scandir, collectPerDayWalk); err != nil {
        return err
    }

    // scanning month dirs
    for mon := 1; mon <= 12; mon++ {
        mondate := ImageDate{self.year, mon, 1}
        scandir := self.MakePath(mondate.MonthString())

        if Verbose {
            fmt.Println("# scanning", scandir)
        }

        collectPerMonWalk := func (path string, info os.FileInfo, err error) error {
            if err != nil {
                // failed - ignore
                return nil
            }

            if info.Mode().IsDir() {
                if path == scandir {
                    // the scandir itself - ignore
                    return nil
                } else {
                    // all enclosed dirs - skip
                    return filepath.SkipDir
                }
            }

            dir, _ := self.get_mondir(mon)
            dir.Add(info.Name())
            return nil
        }

        if err := filepath.Walk(scandir, collectPerMonWalk); err != nil {
            return err
        }
    }

    return nil
}


/// find a suitable location to place an image, or return error
func (self *YearDays) FindDay(date ImageDate, dstname string) (string, error) {
    var dirset *StrSet
    var ok bool
    if dirset, ok = self.day2dir[date]; ok {
        // a dirset is found, check the contents
        var found string
        for _, dirname := range dirset.Keys() {
            found = dirname
            dir := self.daydirs[dirname]
            if dir.Has(dstname) {
                return "", fmt.Errorf("file exists @ %s", dirname)
            }
        }
        // file is not found in subdirs
        if dirset.Len() == 1 {
            // only one subdir
            self.daydirs[found].Add(dstname)  // update the output
            return self.MakePath("all", found, dstname), nil
        }
    } else {
        // no dirset exists
        dirset = NewStrSet()
        self.day2dir[date] = dirset
    }

    dirname := date.DayString()
    dirset.Add(dirname)
    dir := self.daydirs[dirname]
    if dir == nil {
        dir = NewDirectory()
        self.daydirs[dirname] = dir
        self.tomake = append(self.tomake, self.MakePath("all", dirname))
    }
    dir.Add(dstname)
    return self.MakePath("all", dirname, dstname), nil
}


func (self *YearDays) FindMonth(date ImageDate, dstname string) (string, error) {
    dir, justmade := self.get_mondir(date.month)
    if dir.Has(dstname) {
        return "", fmt.Errorf("file exists @ %d-%d", self.year, date.month)
    }
    dir.Add(dstname)
    if justmade {
        self.tomake = append(self.tomake, self.MakePath(date.MonthString()))
    }
    return self.MakePath(date.MonthString(), dstname), nil
}


func (self *YearDays) MakeAllDirs() error {
    for _, dir := range self.tomake {
        if Verbose {
            fmt.Println("# mkdir", dir)
        }
        if DryRun {
            continue
        }
        if err := os.MkdirAll(dir, 0755); err != nil {
            return err
        }
    }
    self.tomake = nil
    return nil
}
