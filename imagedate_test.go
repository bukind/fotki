package fotki

import "testing"

func TestImageDateIsEmpty(t *testing.T) {
	var img ImageDate
	if !img.IsEmpty() {
		t.Error("image after construction is not empty")
	}
	img = ImageDate{2016, 02, 12}
	if img.IsEmpty() {
		t.Error("image is empty while it shouldn't be")
	}
}


func TestImageStrings(t *testing.T) {
	img := ImageDate{2016, 02, 12}
	expected := "2016-02-12"
	if img.String() != expected {
		t.Errorf("incorrect representation of %v: %s != %s", img, img.String(), expected)
	}
}


func TestImageDateLess(t *testing.T) {
    a := ImageDate{}
	b := ImageDate{2016, 02, 17}
	c := ImageDate{2016, 02, 18}
	d := ImageDate{2017, 02, 17}
	e := ImageDate{2016, 03, 17}

	type testRes struct {
	    a ImageDate
		b ImageDate
		res1 bool
		res2 bool
	}

	tests := []testRes{
	    {a, a, false, false},
		{a, b, true, false},
		{b, b, false, false},
		{b, c, true, false},
		{b, d, true, false},
		{b, e, true, false},
	}

	for _, tr := range tests {
	    res1 := tr.a.Less(tr.b)
	    if res1 != tr.res1 {
		    t.Errorf("bad comparison %v.Less(%v) = %v must be %v",
			         tr.a, tr.b, res1, tr.res1)
		}
	    res2 := tr.b.Less(tr.a)
	    if res2 != tr.res2 {
		    t.Errorf("bad comparison %v.Less(%v) = %v must be %v",
			         tr.b, tr.a, res2, tr.res2)
		}
	}
}
