package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/idpbuilder/meta/internal/host"
)

func TestSelectApplyMode(t *testing.T) {
	tests := []struct {
		name         string
		hostPlatform string
		homeOnly     bool
		wantMode     string
		wantErr      bool
	}{
		{
			name:         "macos default",
			hostPlatform: "macos",
			wantMode:     applyModeDarwin,
		},
		{
			name:         "nixos default",
			hostPlatform: "nixos",
			wantMode:     applyModeNixOS,
		},
		{
			name:         "linux default",
			hostPlatform: "arch",
			wantMode:     applyModeLinux,
		},
		{
			name:         "nixos home only",
			hostPlatform: "nixos",
			homeOnly:     true,
			wantMode:     applyModeHomeOnly,
		},
		{
			name:         "linux home only",
			hostPlatform: "arch",
			homeOnly:     true,
			wantMode:     applyModeHomeOnly,
		},
		{
			name:         "macos home only unsupported",
			hostPlatform: "macos",
			homeOnly:     true,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMode, err := selectApplyMode(tt.hostPlatform, tt.homeOnly)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected err=%v, got err=%v", tt.wantErr, err)
			}
			if gotMode != tt.wantMode {
				t.Fatalf("expected mode %q, got %q", tt.wantMode, gotMode)
			}
		})
	}
}

// ── resolveSwitchHost ──────────────────────────────────────────────────────

func stubHostLookups(t *testing.T, get func(string, string) (*host.Info, error), detect func(string) (*host.Info, error)) {
	t.Helper()
	origGet, origDetect := getHostFn, detectHostFn
	getHostFn, detectHostFn = get, detect
	t.Cleanup(func() { getHostFn, detectHostFn = origGet, origDetect })
}

func TestResolveSwitchHostPrecedence(t *testing.T) {
	mk := func(name, platform string) *host.Info {
		return &host.Info{Name: name, Platform: platform}
	}

	t.Run("positional arg wins over flag", func(t *testing.T) {
		var requested string
		stubHostLookups(t,
			func(dir, name string) (*host.Info, error) {
				requested = name
				return mk(name, "nixos"), nil
			},
			func(dir string) (*host.Info, error) {
				t.Fatal("detect should not be called")
				return nil, nil
			},
		)

		info, err := resolveSwitchHost("/org", []string{"arg-host"}, "flag-host")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if requested != "arg-host" || info.Name != "arg-host" {
			t.Errorf("expected arg-host, got requested=%q info=%q", requested, info.Name)
		}
	})

	t.Run("flag used when no arg", func(t *testing.T) {
		stubHostLookups(t,
			func(dir, name string) (*host.Info, error) { return mk(name, "arch"), nil },
			func(dir string) (*host.Info, error) {
				t.Fatal("detect should not be called")
				return nil, nil
			},
		)

		info, err := resolveSwitchHost("/org", nil, "flag-host")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != "flag-host" {
			t.Errorf("expected flag-host, got %q", info.Name)
		}
	})

	t.Run("auto-detect fallback", func(t *testing.T) {
		stubHostLookups(t,
			func(dir, name string) (*host.Info, error) {
				t.Fatal("getHost should not be called")
				return nil, nil
			},
			func(dir string) (*host.Info, error) { return mk("detected", "macos"), nil },
		)

		info, err := resolveSwitchHost("/org", nil, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != "detected" {
			t.Errorf("expected detected, got %q", info.Name)
		}
	})

	t.Run("unknown host error", func(t *testing.T) {
		stubHostLookups(t,
			func(dir, name string) (*host.Info, error) { return nil, errors.New("nope") },
			func(dir string) (*host.Info, error) { return nil, errors.New("nope") },
		)

		if _, err := resolveSwitchHost("/org", []string{"ghost"}, ""); err == nil {
			t.Error("expected error for unknown host")
		} else if !strings.Contains(err.Error(), `host "ghost" not found`) {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("detect failure error", func(t *testing.T) {
		stubHostLookups(t,
			func(dir, name string) (*host.Info, error) { return nil, errors.New("nope") },
			func(dir string) (*host.Info, error) { return nil, errors.New("no match") },
		)

		if _, err := resolveSwitchHost("/org", nil, ""); err == nil {
			t.Error("expected error for detect failure")
		} else if !strings.Contains(err.Error(), "auto-detection failed") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

// ── switch apply paths (exec seams) ────────────────────────────────────────

type execRecorder struct {
	calls [][]string
	dirs  []string
	err   error
}

func (r *execRecorder) runVisibleDir(dir, name string, args ...string) error {
	r.dirs = append(r.dirs, dir)
	r.calls = append(r.calls, append([]string{name}, args...))
	return r.err
}

func (r *execRecorder) runVisible(name string, args ...string) error {
	return r.runVisibleDir("", name, args...)
}

func stubExec(t *testing.T, rec *execRecorder, cmdExists func(string) bool) {
	t.Helper()
	origDir, origVis, origExists := runVisibleDirFn, runVisibleFn, commandExistsFn
	runVisibleDirFn = rec.runVisibleDir
	runVisibleFn = rec.runVisible
	if cmdExists != nil {
		commandExistsFn = cmdExists
	}
	t.Cleanup(func() {
		runVisibleDirFn, runVisibleFn, commandExistsFn = origDir, origVis, origExists
	})
}

func argvString(call []string) string { return strings.Join(call, " ") }

func TestSwitchNixOS(t *testing.T) {
	rec := &execRecorder{}
	stubExec(t, rec, nil)

	if err := switchNixOS("/org/cmdr", ".#strix-nix"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rec.calls) != 1 {
		t.Fatalf("expected 1 exec call, got %d", len(rec.calls))
	}
	want := "sudo nixos-rebuild switch --flake .#strix-nix"
	if got := argvString(rec.calls[0]); got != want {
		t.Errorf("argv = %q, want %q", got, want)
	}
	if rec.dirs[0] != "/org/cmdr" {
		t.Errorf("dir = %q, want /org/cmdr", rec.dirs[0])
	}
}

func TestSwitchNixOSError(t *testing.T) {
	rec := &execRecorder{err: errors.New("boom")}
	stubExec(t, rec, nil)

	err := switchNixOS("/org/cmdr", ".#strix-nix")
	if err == nil || !strings.Contains(err.Error(), "nixos-rebuild switch failed") {
		t.Errorf("expected wrapped nixos-rebuild error, got %v", err)
	}
}

func TestSwitchLinux(t *testing.T) {
	rec := &execRecorder{}
	stubExec(t, rec, nil)

	if err := switchLinux("/org/cmdr", ".#cachyos"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "home-manager switch --flake .#cachyos"
	if got := argvString(rec.calls[0]); got != want {
		t.Errorf("argv = %q, want %q", got, want)
	}
}

func TestSwitchLinuxError(t *testing.T) {
	rec := &execRecorder{err: errors.New("boom")}
	stubExec(t, rec, nil)

	err := switchLinux("/org/cmdr", ".#cachyos")
	if err == nil || !strings.Contains(err.Error(), "home-manager switch failed") {
		t.Errorf("expected wrapped home-manager error, got %v", err)
	}
}

func TestSwitchHomeOnlyApply(t *testing.T) {
	rec := &execRecorder{}
	stubExec(t, rec, nil)

	if err := switchHomeOnlyApply("/org/cmdr", "strix-nix"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "nix run .#homeConfigurations.strix-nix.activationPackage"
	if got := argvString(rec.calls[0]); got != want {
		t.Errorf("argv = %q, want %q", got, want)
	}
}

func TestSwitchHomeOnlyApplyError(t *testing.T) {
	rec := &execRecorder{err: errors.New("boom")}
	stubExec(t, rec, nil)

	err := switchHomeOnlyApply("/org/cmdr", "strix-nix")
	if err == nil || !strings.Contains(err.Error(), "home-only apply failed") {
		t.Errorf("expected wrapped home-only error, got %v", err)
	}
}

func TestSwitchDarwinWithRebuild(t *testing.T) {
	rec := &execRecorder{}
	stubExec(t, rec, func(name string) bool { return name == "darwin-rebuild" })

	if err := switchDarwin("/org/cmdr", ".#studio"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "sudo darwin-rebuild switch --flake .#studio"
	if got := argvString(rec.calls[0]); got != want {
		t.Errorf("argv = %q, want %q", got, want)
	}
}

func TestSwitchDarwinBootstrap(t *testing.T) {
	rec := &execRecorder{}
	stubExec(t, rec, func(name string) bool { return false })

	if err := switchDarwin("/org/cmdr", ".#studio"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Last call must be the nix-darwin bootstrap invocation (an /etc/bashrc
	// move may precede it depending on the test host).
	last := rec.calls[len(rec.calls)-1]
	want := "sudo nix run nix-darwin/master#darwin-rebuild -- switch --flake .#studio"
	if got := argvString(last); got != want {
		t.Errorf("argv = %q, want %q", got, want)
	}
}

func TestSwitchDarwinError(t *testing.T) {
	rec := &execRecorder{err: errors.New("boom")}
	stubExec(t, rec, func(name string) bool { return name == "darwin-rebuild" })

	err := switchDarwin("/org/cmdr", ".#studio")
	if err == nil || !strings.Contains(err.Error(), "darwin-rebuild switch failed") {
		t.Errorf("expected wrapped darwin-rebuild error, got %v", err)
	}
}

// ── runSwitch end-to-end dispatch ──────────────────────────────────────────

func TestRunSwitchDispatch(t *testing.T) {
	// Fake org dir so resolveOrgDir succeeds
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "cmdr"), 0o755)
	os.WriteFile(filepath.Join(tmp, ".gitmodules"), []byte("[submodule]"), 0o644)
	os.WriteFile(filepath.Join(tmp, "cmdr", "flake.nix"), []byte("{}"), 0o644)
	orgDir = tmp
	t.Cleanup(func() { orgDir = "" })

	tests := []struct {
		name     string
		platform string
		homeOnly bool
		wantArgv string
	}{
		{"nixos default", "nixos", false, "sudo nixos-rebuild switch --flake .#test-host"},
		{"arch default", "arch", false, "home-manager switch --flake .#test-host"},
		{"nixos home-only", "nixos", true, "nix run .#homeConfigurations.test-host.activationPackage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubHostLookups(t,
				func(dir, name string) (*host.Info, error) {
					return &host.Info{Name: name, Platform: tt.platform}, nil
				},
				func(dir string) (*host.Info, error) {
					t.Fatal("detect should not be called")
					return nil, nil
				},
			)
			rec := &execRecorder{}
			stubExec(t, rec, nil)

			switchHomeOnly = tt.homeOnly
			t.Cleanup(func() { switchHomeOnly = false })

			if err := runSwitch(switchCmd, []string{"test-host"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(rec.calls) != 1 {
				t.Fatalf("expected 1 exec call, got %d", len(rec.calls))
			}
			if got := argvString(rec.calls[0]); got != tt.wantArgv {
				t.Errorf("argv = %q, want %q", got, tt.wantArgv)
			}
		})
	}

	t.Run("macos home-only rejected", func(t *testing.T) {
		stubHostLookups(t,
			func(dir, name string) (*host.Info, error) {
				return &host.Info{Name: name, Platform: "macos"}, nil
			},
			func(dir string) (*host.Info, error) { return nil, errors.New("unused") },
		)
		rec := &execRecorder{}
		stubExec(t, rec, nil)

		switchHomeOnly = true
		t.Cleanup(func() { switchHomeOnly = false })

		err := runSwitch(switchCmd, []string{"test-host"})
		if err == nil || !strings.Contains(err.Error(), "not supported on macOS") {
			t.Errorf("expected macOS home-only rejection, got %v", err)
		}
		if len(rec.calls) != 0 {
			t.Errorf("expected no exec calls, got %d", len(rec.calls))
		}
	})
}
