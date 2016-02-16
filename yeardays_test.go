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
		t.Fatal("cannot create %s: %s", dir, err.Error())
	}
	if !info.IsDir() || info.Name() != "world" {
		t.Fatal("wrong parameters of the directory, %v", info)
	}
}
