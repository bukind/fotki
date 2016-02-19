package fotki

import (
	"testing"
)

func TestMakePath(t *testing.T) {
	if makePath() != "" {
		t.Error("empty makePath() fails")
	}
	if makePath("") != "" {
		t.Error("makePath(\"\") fails")
	}
	if makePath("/") != "/" {
		t.Error("makePath(\"/\") fails")
	}
	if makePath(".") != "." {
		t.Error("makePath(\".\") fails")
	}
	if makePath(".", ".") != "." {
		t.Error("makePath(\"./.\") fails")
	}
	if makePath("hello", "world") != "hello/world" {
		t.Error("makePath(\"hello/world\") fails")
	}
}
