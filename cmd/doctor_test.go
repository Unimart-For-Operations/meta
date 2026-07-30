package cmd

import "testing"

func TestPlatformToolChecks(t *testing.T) {
	tests := []struct {
		platform  string
		wantNames []string
	}{
		{"macos", []string{"darwin-rebuild", "brew"}},
		{"nixos", []string{"nixos-rebuild"}},
		{"arch", []string{"home-manager"}},
		{"ubuntu", []string{"home-manager"}},
		{"linux", []string{"home-manager"}},
		{"", []string{"home-manager"}},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			checks := platformToolChecks(tt.platform)
			if len(checks) != len(tt.wantNames) {
				t.Fatalf("expected %d checks, got %d", len(tt.wantNames), len(checks))
			}
			for i, want := range tt.wantNames {
				if checks[i].name != want {
					t.Errorf("check[%d] = %q, want %q", i, checks[i].name, want)
				}
				if checks[i].check == nil {
					t.Errorf("check[%d] %q has nil check func", i, want)
				}
			}
		})
	}
}
