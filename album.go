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
)

type Album struct {
    root string
		images map[string]ImageDate // good image -> their dates
		failed map[string]error     // failed image -> error
}


func NewAlbum(rootdir string) *Album {
    self := new(Album)
		self.root = rootdir
		self.images = make(map[string]ImageDate)
		self.failed = make(map[string]error)
		return self
}


func (self *Album) String() string {
    buf := new(bytes.Buffer)
    fmt.Fprintf(buf, "root=%s\n", self.root)
    for img, date := range self.images {
        fmt.Fprintf(buf, " %s => %s\n", img, date.String())
    }
    for img, err := range self.failed {
        fmt.Fprintf(buf, "\n %s => Error %s\n", img, err.Error())
    }
    return buf.String()
}


func (self *Album) Scan(scandir string) error {

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
                date, err := self.ExtractImageDate(path)
                if err == nil {
                    self.images[path] = date
                } else {
                    self.failed[path] = err
                }
        }
        return nil
    }

    if err := filepath.Walk(scandir, walkFun); err != nil {
        return err
    }

    // load years
    years := self.Years()
    for _, year := range years {
        ydir := NewYearDays(self, year)
        if err := ydir.Scan(); err != nil {
            return err
        }
        if Verbose {
            fmt.Println(ydir.String())
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


func (self *Album) Years() []int {
    years := make(map[int]bool)
    for _, d := range self.images {
        years[d.year] = true
    }
    res := make([]int, 0, len(years))
    for year, _ := range years {
        res = append(res,year)
    }
    sort.Ints(res)
    return res
}
