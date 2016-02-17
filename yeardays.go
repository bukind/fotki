package fotki

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const daybase = ""
const monbase = "all"

type YearDays struct {
	album   *Album // backlink to album
	year    int
	daydirs map[string]*Directory // relative dirpath -> directory for the day
	day2dir map[ImageDate]*StrSet // mapping date -> set of relative dirpath
	mon2dir map[int]*Directory    // month -> directory
	garbage *StrSet               // relative path
	basedir string
	tomake  []string // list of directories to make
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

func (self *YearDays) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "year:%d => %s\n", self.year, self.basedir)
	dkeys := make([]ImageDate, 0, len(self.day2dir))

	for k, _ := range self.day2dir {
		dkeys = append(dkeys, k)
	}
	sort.Sort(ByImageDate(dkeys))
	for _, k := range dkeys {
		fmt.Fprintf(buf, " %s: %s\n", k.String(), self.day2dir[k].String())
	}
	skeys := make([]string, 0, len(self.daydirs))
	for k, _ := range self.daydirs {
		skeys = append(skeys, k)
	}
	sort.Strings(skeys)
	for _, k := range skeys {
		fmt.Fprintf(buf, " %s: %s\n", k, self.daydirs[k].String())
	}
	for k := 1; k < 12; k++ {
		if _, ok := self.mon2dir[k]; !ok {
			continue
		}
		fmt.Fprintf(buf, " %02d: %s\n", k, self.mon2dir[k].String())
	}
	fmt.Fprintf(buf, " .garbage: %s\n", self.garbage.String())
	return buf.String()
}

// return the directory and the flag if it was just made
func (self *YearDays) get_mondir(month int) (*Directory, bool) {
	mondir := self.mon2dir[month]
	justmade := false
	if mondir == nil {
		mondir = NewDirectory(self.basedir, monbase, fmt.Sprintf("%02d", month))
		self.mon2dir[month] = mondir
		justmade = true
	}
	return mondir, justmade
}

func (self *YearDays) Scan() error {

	// scanning day dirs

	dayscandir := NewDirectory(self.basedir, daybase).Path()
	monscandir := NewDirectory(self.basedir, monbase).Path()

	yearPerDayRegex := regexp.MustCompile(`^(20[0123]\d)[-_](0[1-9]|1[012])[-_](0[1-9]|[123]\d)`)

	if Verbose {
		fmt.Println("# scanning", dayscandir)
	}

	// the function to scan
	collectPerDayWalk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// failed - ignore
			return nil
		}
		if info.Mode().IsDir() {
			if path == dayscandir {
				// the scandir itself - ignore
				return nil
			} else if path == monscandir {
				// the basedir for month - skip
				return filepath.SkipDir
			}
			if Verbose {
				fmt.Println("# dir", path)
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
				self.daydirs[info.Name()] = NewDirectory(path)
				strset := self.day2dir[date]
				strset.Add(info.Name())
			} else {
				// invalid subdirectory!
				self.garbage.Add(path)
			}
			return filepath.SkipDir
		} else {
			// not a dir at the first level, add to the garbage
			self.garbage.Add(path)
		}
		return nil
	}

	if err := filepath.Walk(dayscandir, collectPerDayWalk); err != nil {
		return err
	}

	// scanning individual day directories
	for _, dir := range self.daydirs {

		scanpath := dir.Path()
		if Verbose {
			fmt.Println("# scanning", scanpath)
		}
		collectDayDirWalk := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				// failed - ignore
				return nil
			}
			if info.Mode().IsDir() {
				if path == scanpath {
					// the scanpath itself - ignore
					return nil
				}
				// else garbage
				self.garbage.Add(path)
				return filepath.SkipDir
			} else {
				// not a dir
				dir.Add(info.Name())
			}
			return nil
		}

		if err := filepath.Walk(scanpath, collectDayDirWalk); err != nil {
			return err
		}
	}

	// scanning month dirs
	for mon := 1; mon <= 12; mon++ {
		scandir := NewDirectory(self.basedir, monbase, fmt.Sprintf("%02d", mon)).Path()

		if Verbose {
			fmt.Println("# scanning", scandir)
		}

		collectPerMonWalk := func(path string, info os.FileInfo, err error) error {
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

	if Verbose {
		fmt.Println("# Scanned", self.String())
	}
	return nil
}

/// find a suitable location to place an image, or return error
func (self *YearDays) FindDay(date ImageDate, dstname string, srcinfo os.FileInfo) (string, error) {
	var dirset *StrSet
	var ok bool
	if dirset, ok = self.day2dir[date]; ok {
		// a dirset is found, check the contents
		var found string
		if Verbose {
			fmt.Printf("# dirset is %s\n", dirset.String())
		}
		for _, dirname := range dirset.Keys() {
			found = dirname
			dir := self.daydirs[dirname]
			if Verbose {
				fmt.Printf("# checking %s\n", dir.Path())
			}
			if dir.Has(dstname) {
				if Verbose {
					fmt.Println("# file is found in day", dir.Path(dstname))
				}
				return self.compareInfo(dir.Path(dstname), srcinfo)
			}
		}
		// file is not found in subdirs
		if dirset.Len() == 1 {
			// only one subdir
			dir := self.daydirs[found]
			dir.Add(dstname) // update the output
			if Verbose {
				fmt.Printf("# file %s is not found in a single daydir %s\n", dstname, dir.Path(dstname))
			}
			return dir.Path(dstname), nil
		}
	} else {
		// no dirset exists
		dirset = NewStrSet()
		self.day2dir[date] = dirset
	}

	dirname := date.String()
	dirset.Add(dirname)
	dir := self.daydirs[dirname]
	if dir == nil {
		// just has been made
		dir = NewDirectory(self.basedir, daybase, dirname)
		self.daydirs[dirname] = dir
		self.tomake = append(self.tomake, dir.Path())
	}
	dir.Add(dstname)
	if Verbose {
		fmt.Printf("# dayset has no file: %v\n", dirset)
	}
	return dir.Path(dstname), nil
}

func (self *YearDays) FindMonth(date ImageDate, dstname string, srcinfo os.FileInfo) (string, error) {
	dir, justmade := self.get_mondir(date.month)
	if dir.Has(dstname) {
		return self.compareInfo(dir.Path(dstname), srcinfo)
	}
	dir.Add(dstname)
	if justmade {
		self.tomake = append(self.tomake, dir.Path())
	}
	return dir.Path(dstname), nil
}

// check if the destination is the same as origin
func (self *YearDays) compareInfo(dst string, srcinfo os.FileInfo) (string, error) {
	dstinfo, err := os.Stat(dst)
	if err != nil {
		return "", fmt.Errorf("cannot stat %s: %s", dst, err.Error())
	}
	if os.SameFile(dstinfo, srcinfo) {
		return dst, SameFile
	}
	return "", fmt.Errorf("file exists %s", dst)
}

func makedir(dir string) (os.FileInfo, error) {
	info, err := os.Lstat(dir)
	if err == nil {
		if info.IsDir() {
			return info, err
		} else {
			return info, fmt.Errorf("%s is not a dir", dir)
		}
	}
	// get the parent dir
	info2, err := makedir(filepath.Dir(dir))
	if err != nil {
		// could not create the root directory
		return info2, err
	}
	if Verbose {
		fmt.Println("# mkdir", dir)
	}
	if DryRun {
		return info, nil
	}
	err = os.Mkdir(dir, info2.Mode())
	if err != nil {
		return info, err
	}
	return os.Lstat(dir)
}

func (self *YearDays) MakeAllDirs() error {
	for _, dir := range self.tomake {
		_, err := makedir(dir)
		if err != nil {
			return err
		}

		if DryRun {
			continue
		}

		// directory exists, try to create a file
		timestr := time.Now().Format(time.RFC3339Nano)
		fname := filepath.Join(dir, strings.Replace(timestr, ":", ".", -1))
		fd, err := os.Create(fname)

		mkerr := func(action string) error {
			return fmt.Errorf("cannot %s a file in %s: %s", action, dir, err.Error())
		}
		if err != nil {
			return mkerr("create")
		}
		if err = fd.Close(); err != nil {
			return mkerr("close")
		}
		if err = os.Remove(fname); err != nil {
			return mkerr("remove")
		}
	}
	self.tomake = nil
	return nil
}

func (self *YearDays) NormalizeDirs() error {
	dirs := make([]string, 0, len(self.daydirs))
	for dir, _ := range self.daydirs {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	for _, src := range dirs {
		dst := strings.Replace(src, "_", "-", -1)
		if src == dst {
			continue
		}
		src = NewDirectory(self.basedir, daybase, src).Path()
		dst = NewDirectory(self.basedir, daybase, dst).Path()
		if Verbose {
			fmt.Println("# normalize", src, "->", dst)
		}
		if DryRun {
			continue
		}
		if err := os.Rename(src, dst); err != nil {
			return err
		}
	}
	return nil
}
