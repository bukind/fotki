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
	"time"
)

// the initially found location of the image
type ImageLoc struct {
	path string
	info os.FileInfo
	kind ImageKind
}

func GetImageKind(path string) ImageKind {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".png", ".jpeg":
		return IsImage
	case ".mp4", ".avi", ".thm":
		return IsMovie
	default:
		return NoImage
	}
}

func (self *ImageLoc) ExtractDate() (ImageInfo, error) {
	// try to extract the date from the path itself
	ret := ImageInfo{ImageDate{}, self.kind, self.info}
	reg := regexp.MustCompile(`(20[0123]\d)[-_]?(0[1-9]|1[012])[-_]?(0[1-9]|[123]\d)[^/]*$`)
	groups := reg.FindStringSubmatch(self.path)
	if groups != nil {
		// extract year, month, day w/o check
		ret.date.year, _ = strconv.Atoi(groups[1])
		ret.date.month, _ = strconv.Atoi(groups[2])
		ret.date.day, _ = strconv.Atoi(groups[3])
	}

	var err error
	if self.kind != IsMovie {
		cmd := exec.Command("identify", "-ping", "-verbose", self.path)
		out, errx := cmd.Output()
		err = errx
		if err == nil {
			var value string
			scanner := bufio.NewScanner(bytes.NewBuffer(out))
			for scanner.Scan() {
				str := scanner.Text()
				if strings.HasPrefix(str, "    exif:DateTimeOriginal: ") {
					value = strings.SplitN(str, ": ", 2)[1]
				}
				var dummy int
				fmt.Sscanf(value, "%04d%c%02d%c%02d", &ret.date.year, &dummy, &ret.date.month, &dummy, &ret.date.day)
			}
			err = scanner.Err()
		}
	}

	if ret.date.IsEmpty() {
		// try to convert the path to a timestamp
		err = fmt.Errorf("cannot detect the date")
		fp := filepath.Base(self.path)
		fp = fp[:len(fp)-len(filepath.Ext(fp))]
		if stamp, err2 := strconv.ParseInt(fp, 10, 64); err2 == nil {
			const nms = 1000
			ts := time.Unix(stamp/nms, stamp%nms)
			if ts.After(minstamp) && ts.Before(maxstamp) {
				ret.date.year = ts.Year()
				ret.date.month = int(ts.Month())
				ret.date.day = ts.Day()
				err = nil
			}
		}
	}
	return ret, err
}
