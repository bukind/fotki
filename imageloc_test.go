package fotki

import "testing"

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
