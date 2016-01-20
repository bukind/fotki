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
    alldirs  map[string]*Directory
		date2dir map[ImageDate]*StrSet
		garbage  *StrSet
}


func NewYearDays() *YearDays {
    self := new(YearDays)
		self.alldirs = make(map[string]*Directory)
		self.date2dir = make(map[ImageDate]*StrSet)
		self.garbage = new(StrSet)
		return self
}


func (self *YearDays) String() string {
    buf := new(bytes.Buffer)
		for k, v := range self.date2dir {
		    fmt.Fprintf(buf, " %s: %s\n", k.String(), v.String())
		}
		fmt.Fprintf(buf, " .garbage: %s\n", self.garbage.String())
		return buf.String()
}


func (self *YearDays) Scan(scandir string) error {
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
                if _, ok := self.date2dir[date]; !ok {
                    self.date2dir[date] = new(StrSet)
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

    return filepath.Walk(scandir, collectPerDayWalk)
}
