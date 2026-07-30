package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateK9sSkin(t *testing.T) {
	path := filepath.Join("fixtures", "theme-sample.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tObj, err := parseTheme(b)
	if err != nil {
		t.Fatal(err)
	}
	skin := GenerateK9sSkin(tObj)
	if skin == "" {
		t.Fatal("empty skin")
	}
}
