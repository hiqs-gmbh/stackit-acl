package selfupdate

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"v0.4.0", "0.4.0"},
		{"0.4.0", "0.4.0"},
		{"v0.4.0+dirty", "0.4.0+dirty"},
		{"dev", "dev"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := Normalize(tt.in); got != tt.want {
			t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTag(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"0.4.0", "v0.4.0"},
		{"v0.4.0", "v0.4.0"},
		{"", "v"},
	}
	for _, tt := range tests {
		if got := Tag(tt.in); got != tt.want {
			t.Errorf("Tag(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
