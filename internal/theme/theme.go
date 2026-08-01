package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Theme is a minimal model of the exported theme JSON we care about.
type Theme struct {
	Name     string                 `json:"name"`
	Palette  map[string]interface{} `json:"palette"`
	Semantic map[string]string      `json:"semantic"`
	Terminal struct {
		ANSI map[string]string `json:"ansi"`
	} `json:"terminal"`
	Fonts map[string]map[string]interface{} `json:"fonts"`
}

// LoadFromOrg tries to obtain the theme JSON from the org directory.
// It prefers calling the cmdr export script if present, falling back to
// looking for a checked-in theme.json at common locations.
func LoadFromOrg(orgDir, themeName string) (*Theme, error) {
	// Preferred: run cmdr/scripts/theme-export.sh
	script := filepath.Join(orgDir, "cmdr", "scripts", "theme-export.sh")
	if _, err := os.Stat(script); err == nil {
		cmd := exec.Command(script, themeName)
		cmd.Dir = filepath.Join(orgDir, "cmdr")
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("running theme-export: %w", err)
		}
		return parseTheme(out)
	}

	// Fallbacks: orgDir/theme.json or orgDir/cmdr/theme.json
	candidates := []string{
		filepath.Join(orgDir, "theme.json"),
		filepath.Join(orgDir, "cmdr", "theme.json"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			b, err := os.ReadFile(p)
			if err != nil {
				return nil, fmt.Errorf("reading %s: %w", p, err)
			}
			return parseTheme(b)
		}
	}

	return nil, fmt.Errorf("no theme export found in org dir: %s", orgDir)
}

func parseTheme(b []byte) (*Theme, error) {
	var t Theme
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, fmt.Errorf("parsing theme json: %w", err)
	}
	return &t, nil
}

// DMSThemePaths returns the paths where DMS's matugen writes the k9s skin and
// tmux theme on DMS hosts. These are regenerated on every wallpaper/theme
// change, so they reflect the host's active DMS theme. The files may not exist
// yet (e.g. first boot before matugen has run); callers should fall back to the
// static Catppuccin export when they don't.
func DMSThemePaths(homeDir string) (k9s, tmux string) {
	return filepath.Join(homeDir, ".config", "k9s", "skins", "dank.yaml"),
		filepath.Join(homeDir, ".config", "tmux", "dank-theme.conf")
}

// GenerateTmuxStatus returns a simple tmux status-right snippet using
// the theme's semantic colors. This is a minimal example for testing.
func GenerateTmuxStatus(t *Theme) string {
	fg := t.Semantic["fg"]
	accent := t.Semantic["accent"]
	bg := t.Semantic["bgPanel"]
	if fg == "" {
		fg = "white"
	}
	if accent == "" {
		accent = "green"
	}
	if bg == "" {
		bg = "black"
	}
	// tmux uses colour names or #rrggbb
	return fmt.Sprintf("set -g status-right \"#[fg=%s,bg=%s] %s #[fg=%s,bg=%s] %s \"",
		fg, bg, "#(whoami)", fg, accent, "#(date '+%%Y-%%m-%%d')")
}

// GenerateK9sSkin builds a minimal k9s skin mapping using semantic colors.
// Returns YAML suitable for k9s skin files.
func GenerateK9sSkin(t *Theme) string {
	fg := t.Semantic["fg"]
	bg := t.Semantic["bgPanel"]
	accent := t.Semantic["accent"]
	warn := t.Semantic["warn"]
	ok := t.Semantic["ok"]
	if fg == "" {
		fg = "#c6d0f5"
	}
	if bg == "" {
		bg = "#292c3c"
	}
	if accent == "" {
		accent = "#ca9ee6"
	}
	if warn == "" {
		warn = "#ef9f76"
	}
	if ok == "" {
		ok = "#a6d189"
	}

	skin := "background: " + bg + "\n"
	skin += "styles:\n"
	skin += "  default:\n"
	skin += "    fg: " + fg + "\n"
	skin += "  header:\n"
	skin += "    fg: " + accent + "\n"
	skin += "  error:\n"
	skin += "    fg: " + warn + "\n"
	skin += "  success:\n"
	skin += "    fg: " + ok + "\n"
	return skin
}
