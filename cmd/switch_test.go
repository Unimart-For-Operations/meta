package cmd

import "testing"

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
