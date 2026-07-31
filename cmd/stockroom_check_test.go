package cmd

import "testing"

func TestClassifySubmoduleRemote(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		ok    bool
		label string
	}{
		{
			name:  "relative remote",
			url:   "../cmdr.git",
			ok:    true,
			label: "→ ../cmdr.git (relative)",
		},
		{
			name:  "legacy idpbuilder ssh remote",
			url:   "git@github.com:idpbuilder/cmdr.git",
			ok:    false,
			label: "",
		},
		{
			name:  "unimart for operations ssh remote",
			url:   "git@github.com:Unimart-For-Operations/cmdr.git",
			ok:    true,
			label: "→ Unimart-For-Operations org",
		},
		{
			name: "unexpected github org",
			url:  "git@github.com:example/cmdr.git",
			ok:   false,
		},
		{
			name: "non-github remote",
			url:  "git@gitlab.com:Unimart-For-Operations/cmdr.git",
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label, ok := classifySubmoduleRemote(tt.url)
			if ok != tt.ok {
				t.Fatalf("expected ok=%v, got %v", tt.ok, ok)
			}
			if label != tt.label {
				t.Fatalf("expected label %q, got %q", tt.label, label)
			}
		})
	}
}
