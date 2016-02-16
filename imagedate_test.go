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

func TestImageString(t *testing.T) {
	img := ImageDate{2016, 02, 12}
	expected := "2016-02-12"
	if img.String() != expected {
		t.Errorf("incorrect representation of %v: %s != %s", img, img.String(), expected)
	}
}
