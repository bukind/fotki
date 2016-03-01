package fotki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMakedir(t *testing.T) {
	DryRun = false
	Verbose = true

	temp := os.TempDir()
	tstr := time.Now().Format(time.RFC3339Nano)
	tstr = strings.Replace(tstr, ":", ".", -1)
	dir := filepath.Join(temp, tstr, "hello", "world")
	fmt.Println("making a directory", dir)
	info, err := makedir(dir)
	if err != nil {
		t.Fatalf("cannot create %s: %s\n", dir, err.Error())
	}
	if !info.IsDir() || info.Name() != "world" {
		t.Fatalf("wrong parameters of the directory, %v\n", info)
	}
}

func TestYearDaysIsLeafDir(t *testing.T) {
	self := new(yearDays)
	dir := "/tmp"
	self.basedir = dir

	ensure := func(expected bool, path ...string) {
		full := makePath(dir, makePath(path...))
		res := self.IsLeafDir(full)
		if res != expected {
			t.Errorf("incorrect result of IsLeafDir(%s) = %v, must be %v\n", full, res, expected)
		}
	}

	ensure(false)
	ensure(true, daybase, "2001-01-01")
	ensure(true, daybase, "2001-01-01-hello")
	ensure(false, daybase, "2001")
	ensure(false, daybase, "2001_01_01")
	ensure(false, daybase, "2001-01-01badsuffix")
	ensure(false, monbase, "00")
	ensure(true, monbase, "01")
	ensure(false, monbase, "aa")
}
