package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromOrg_Fallback(t *testing.T) {
	tmp := t.TempDir()
	// create a fake cmdr/theme.json
	cmdrDir := filepath.Join(tmp, "cmdr")
	if err := os.MkdirAll(cmdrDir, 0755); err != nil {
		t.Fatal(err)
	}
	sample := `{"name":"catppuccin-frappe","semantic":{"fg":"#c6d0f5","accent":"#ca9ee6"}}`
	if err := os.WriteFile(filepath.Join(cmdrDir, "theme.json"), []byte(sample), 0644); err != nil {
		t.Fatal(err)
	}

	tObj, err := LoadFromOrg(tmp, "catppuccin-frappe")
	if err != nil {
		t.Fatalf("LoadFromOrg failed: %v", err)
	}
	if tObj.Name != "catppuccin-frappe" {
		t.Fatalf("unexpected name: %s", tObj.Name)
	}
	if tObj.Semantic["fg"] != "#c6d0f5" {
		t.Fatalf("unexpected fg: %s", tObj.Semantic["fg"])
	}
}
