package fotki

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
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

type YearDayKeeper interface {
	// represent the inner state
	String() string
	// adopt a new image into the year
	Adopt(info *ImageInfo) ([]string, bool, error)

	NormalizeDirs() error
}

type yearDays struct {
	basedir string
	dirs    map[string]*Directory // absolute path -> directory
	day2dir map[ImageDate]*StrSet // mapping date -> set of path
	garbage *StrSet               // absolute path
}

func MakeYearDays(rootdir string, year int) (YearDayKeeper, error) {
	self := new(yearDays)
	self.basedir = filepath.Join(rootdir, strconv.Itoa(year))
	self.dirs = make(map[string]*Directory)
	self.day2dir = make(map[ImageDate]*StrSet)
	self.garbage = NewStrSet()
	return self, self.scan()
}

func (self *yearDays) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "year:%s\n", self.basedir)
	dkeys := make([]ImageDate, 0, len(self.day2dir))

	for k, _ := range self.day2dir {
		dkeys = append(dkeys, k)
	}
	sort.Sort(ByImageDate(dkeys))
	for _, k := range dkeys {
		fmt.Fprintf(buf, " %s: %s\n", k.String(), self.day2dir[k].String())
	}
	skeys := make([]string, 0, len(self.dirs))
	for k, _ := range self.dirs {
		skeys = append(skeys, k)
	}
	sort.Strings(skeys)
	for _, k := range skeys {
		fmt.Fprintf(buf, " %s: %s\n", k, self.dirs[k].String())
	}
	fmt.Fprintf(buf, " .garbage: %s\n", self.garbage.String())
	return buf.String()
}

// return the directory and the flag if it was just made
func (self *yearDays) get_mondir(month int) (*Directory, bool) {
	path := makePath(self.basedir, monbase, fmt.Sprintf("%02d", month))
	mondir, ok := self.dirs[path]
	if !ok {
		mondir = NewDirectory(path)
		self.dirs[path] = mondir
	}
	return mondir, !ok
}

// This one is called from constructor
func (self *yearDays) scan() error {

	// scanning day dirs

	dayscandir := makePath(self.basedir, daybase)
	monscandir := makePath(self.basedir, monbase)

	yearPerDayRegex := regexp.MustCompile(`^(20[0123]\d)[-_](0[1-9]|1[012])[-_](0[1-9]|[123]\d)`)

	if Verbose {
		fmt.Println("# scanning year at ", self.basedir)
	}

	// the function to scan
	collectPerDayWalk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // failed - ignore
		}
		root, last := filepath.Split(path)
		root = filepath.Clean(root)

		if info.Mode().IsDir() {
			if Verbose {
				fmt.Println("# dir", path)
			}
			if root == dayscandir {
				match := yearPerDayRegex.FindStringSubmatch(last)
				if match != nil {
					// a match found, can proceed w/o error checking
					var date ImageDate
					date.year, _ = strconv.Atoi(match[1])
					date.month, _ = strconv.Atoi(match[2])
					date.day, _ = strconv.Atoi(match[3])
					if _, ok := self.day2dir[date]; !ok {
						self.day2dir[date] = NewStrSet()
					}
					self.dirs[path] = NewDirectory(path)
					self.day2dir[date].Add(path)
					return nil
				}
			} else if root == monscandir {
				month, err := strconv.Atoi(last)
				if err == nil {
					_, _ = self.get_mondir(month)
					return nil
				}
			}
			if path == self.basedir || path == dayscandir || path == monscandir {
				return nil // the scandir itself - ignore
			}
			// garbage dir remains
			self.garbage.Add(path)
			return filepath.SkipDir
		} else {
			// normal file
			if dir, ok := self.dirs[root]; ok {
				// TODO: save info along
				dir.Add(last, info)
			} else {
				self.garbage.Add(path)
			}
		}
		return nil
	}

	if err := filepath.Walk(self.basedir, collectPerDayWalk); err != nil {
		return err
	}

	if Verbose {
		fmt.Println("# Scanned", self.String())
	}
	return nil
}

/// find a suitable location to place an image, or return error
func (self *yearDays) findDayDir(info *ImageInfo, dstname string) (*Directory, bool, error) {
	var res error
	var dirset *StrSet
	var ok bool
	if dirset, ok = self.day2dir[info.date]; ok {
		// a dirset is found, check the contents
		var found string
		if Verbose {
			fmt.Printf("# dirset is %s\n", dirset.String())
		}
		for _, dirname := range dirset.Keys() {
			found = dirname
			dir := self.dirs[dirname]
			if Verbose {
				fmt.Printf("# checking %s\n", dir.Path())
			}
			if dstinfo, err := dir.Stat(dstname); err == nil {
				if os.SameFile(dstinfo, info.info) {
					res = SameFileError
				} else {
					// compare the md5sum
					res = AlreadyExistError
					if info.info.Size() == dstinfo.Size() {
						if srchash, err := Md5sum(info.path); err == nil {
							if dsthash, err := Md5sum(dir.Path(dstname)); err == nil {
								if bytes.Compare(srchash, dsthash) == 0 {
									res = IdenticalError
								}
							}
						}
					}
				}
				return dir, false, res
			}
		}
		// file is not found in subdirs
		if dirset.Len() == 1 {
			// only one subdir
			dir := self.dirs[found]
			dir.Add(dstname, nil) // update the output
			if Verbose {
				fmt.Printf("# file %s is not found in a single daydir %s\n", dstname, dir.Path(dstname))
			}
			return dir, false, nil
		}
	} else {
		// no dirset exists
		dirset = NewStrSet()
		self.day2dir[info.date] = dirset
	}

	dirname := makePath(self.basedir, daybase, info.date.String())
	dirset.Add(dirname)
	dir := self.dirs[dirname]
	var justmade bool
	if dir == nil {
		// just has been made
		dir = NewDirectory(dirname)
		self.dirs[dirname] = dir
		justmade = true
	}
	return dir, justmade, nil
}

func Md5sum(path string) ([]byte, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	const bufsize = 0x10000
	buf := make([]byte, bufsize)
	h := md5.New()
	for {
		count, err := fd.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if count == 0 {
			break
		}
		h.Write(buf[:count])
	}
	return h.Sum(nil), nil
}

// Compare the destination and the original.
func (self *yearDays) findMonthDir(info *ImageInfo, dstname string) (*Directory, bool, error) {
	var res error
	dir, justmade := self.get_mondir(info.date.month)
	if dstinfo, err := dir.Stat(dstname); err == nil {
		if os.SameFile(dstinfo, info.info) {
			res = SameFileError
		} else {
			res = AlreadyExistError
		}
	}
	return dir, justmade, res
}

// check if the destination is the same as origin
/*
func (self *yearDays) compareInfo(dst string, srcinfo os.FileInfo) (string, error) {
	dstinfo, err := os.Stat(dst)
	if err != nil {
		return "", fmt.Errorf("cannot stat %s: %s", dst, err.Error())
	}
	if os.SameFile(dstinfo, srcinfo) {
		return dst, SameFileError
	}
	return "", fmt.Errorf("file exists %s", dst)
}
*/

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

func (self *yearDays) makeAllDirs(tomake []*Directory) error {
	for _, dir := range tomake {
		_, err := makedir(dir.Path())
		if err != nil {
			return err
		}

		if DryRun {
			continue
		}

		// directory exists, try to create a file
		timestr := time.Now().Format(time.RFC3339Nano)
		fname := dir.Path(strings.Replace(timestr, ":", ".", -1))
		fd, err := os.Create(fname)

		mkerr := func(action string) error {
			return fmt.Errorf("cannot %s a file in %s: %s", action, dir.Path(), err.Error())
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
	return nil
}

func (self *yearDays) NormalizeDirs() error {
	dirs := make([]string, 0, len(self.dirs))
	for _, dir := range self.dirs {
		dirs = append(dirs, dir.Path())
	}
	sort.Strings(dirs)
	for _, src := range dirs {
		root, last := filepath.Split(src)
		dst := filepath.Join(root, strings.Replace(last, "_", "-", -1))
		if src == dst {
			continue
		}
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

func (self *yearDays) Adopt(info *ImageInfo) ([]string, bool, error) {

	srcname := filepath.Base(info.path)
	dstname := strings.Replace(strings.ToLower(srcname), " ", "_", -1)

	if Verbose {
		fmt.Println("# processing", info.path, info.date, "->", dstname)
	}

	var dstfiles []string
	var tomake []*Directory
	funcs := [...]func(*ImageInfo, string) (*Directory, bool, error){self.findMonthDir, self.findDayDir}

	tokill := true // allow to kill the original file
	for _, f := range funcs {
		dir, justmade, err := f(info, dstname)
		if err != nil {
			if err == SameFileError {
				if srcname == dstname && info.path == dir.Path(dstname) {
					// we cannot delete the file as the path are the same as well
					tokill = false
				}
				continue
			}
			return dstfiles, false, err
		} else {
			// the file does not exist yet
			dir.Add(dstname, nil)
			dstfiles = append(dstfiles, dir.Path(dstname))
			if justmade {
				tomake = append(tomake, dir)
			}
		}
	}

	// create all dirs
	return dstfiles, tokill, self.makeAllDirs(tomake)
}
