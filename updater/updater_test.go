package updater

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.4.0", "1.3.0", 1},
		{"1.3.0", "1.4.0", -1},
		{"1.4.0", "1.4.0", 0},
		{"2.0.0", "1.9.9", 1},
		{"1.0.0", "2.0.0", -1},
		{"1.10.0", "1.9.0", 1},
		{"0.0.1", "0.0.0", 1},
	}
	for _, tt := range tests {
		got := compareVersions(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCurrentVersionFormat(t *testing.T) {
	if CurrentVersion == "" {
		t.Error("CurrentVersion must not be empty")
	}
	parts := splitVersion(CurrentVersion)
	if len(parts) != 3 {
		t.Errorf("CurrentVersion %q must have 3 parts (major.minor.patch), got %d", CurrentVersion, len(parts))
	}
}

func TestSplitVersion(t *testing.T) {
	parts := splitVersion("1.4.0")
	if len(parts) != 3 || parts[0] != "1" || parts[1] != "4" || parts[2] != "0" {
		t.Errorf("splitVersion(\"1.4.0\") = %v, want [1 4 0]", parts)
	}
}
