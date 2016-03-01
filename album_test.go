package fotki

import "testing"

func TestAlbumIsLeafDir(t *testing.T) {
	self := NewAlbum("/tmp")

	ensure := func(expected bool, path ...string) {
		full := makePath(path...)
		res := self.IsLeafDir(full)
		if res != expected {
			t.Errorf("incorrect result of IsLeafDir(%s) = %v, must be %v\n", full, res, expected)
		}
	}

	r := self.root
	ensure(false, "")
	ensure(false, "/bar")
	ensure(false, "bar")
	ensure(false, r)
	ensure(false, r, "bar")
	ensure(false, r, "2001", "bar")
	ensure(true, r, "2001", daybase, "2001-01-01")
	ensure(false, r, "2001", daybase, "2003-01-01")
	ensure(true, r, "2001", daybase, "2001-01-01-hello")
	ensure(false, r, "2001", daybase, "2001")
	ensure(false, r, "2001", daybase, "2001_01_01")
	ensure(false, r, "2001", daybase, "2001-01-01badsuffix")
	ensure(false, r, "2001", monbase, "00")
	ensure(true, r, "2001", monbase, "01")
	ensure(false, r, "2001", monbase, "aa")
}
