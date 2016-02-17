package fotki

import (
	"os"
	"testing"
)

func TestGetImageKind(t *testing.T) {
	type testRes struct {
		path string
		res  ImageKind
	}

	tests := []testRes{
		{"", NoImage},
		{".jpg", IsImage},
		{"A.JpG", IsImage},
		{"B.JPEG", IsImage},
		{"C.JPEG_", NoImage},
		{"d.mp4", IsMovie},
		{"e.avi", IsMovie},
	}

	for _, tr := range tests {
		res := GetImageKind(tr.path)
		if res != tr.res {
			t.Errorf("Wrong GetImageKind(%s) = %v, must be %v", tr.path, res, tr.res)
		}
	}
}

func TestImageLocExtractDate(t *testing.T) {
	type testRes struct {
		path string
		date ImageDate
	}

	tests := []testRes{
		{"testdata/tmp.jpg", ImageDate{2012, 12, 10}},
		{"testdata/img2016_02_17.jpg", ImageDate{2016, 02, 17}},
		{"testdata/1455684141000.jpg", ImageDate{2016, 02, 17}},
	}
	for _, tr := range tests {
		osinfo, err := os.Stat(tr.path)
		if err != nil {
			t.Errorf("image is not found %s\n", tr.path)
			continue
		}
		loc := &ImageLoc{tr.path, osinfo, GetImageKind(tr.path)}
		info, err := loc.ExtractDate()
		if err != nil {
			t.Errorf("could not get the date for %s: %s\n", tr.path, err.Error())
			continue
		}
		if !os.SameFile(loc.info, info.info) {
			t.Errorf("different info for %s\n", tr.path)
			continue
		}
		if info.date != tr.date {
			t.Errorf("wrong date found for %s: %v != %v\n", tr.path, info.date, tr.date)
			continue
		}
	}
}
