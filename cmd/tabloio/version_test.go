package main

import "testing"

func TestVersionString(t *testing.T) {
	if got, want := versionString("v1.2.3"), "v1.2.3"; got != want {
		t.Errorf("versionString(%q) = %q, want %q", want, got, want)
	}
	if got := versionString("dev"); got == "" {
		t.Errorf(`versionString("dev") = %q, want a non-empty version`, got)
	}
}
