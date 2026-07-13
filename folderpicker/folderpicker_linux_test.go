package folderpicker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPromptUsesZenity(t *testing.T) {
	binDir := t.TempDir()
	zenity := filepath.Join(binDir, "zenity")
	if err := os.WriteFile(zenity, []byte("#!/bin/sh\nprintf '/tmp/Guild Wars 2\\n'\n"), 0755); err != nil {
		t.Fatalf("failed to create fake zenity: %v", err)
	}
	t.Setenv("PATH", binDir)

	folder, err := Prompt("Select folder")
	if err != nil {
		t.Fatalf("Prompt returned an error: %v", err)
	}
	if folder != "/tmp/Guild Wars 2" {
		t.Fatalf("Prompt returned %q", folder)
	}
}

func TestPromptHandlesPickerCancellation(t *testing.T) {
	binDir := t.TempDir()
	zenity := filepath.Join(binDir, "zenity")
	if err := os.WriteFile(zenity, []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatalf("failed to create fake zenity: %v", err)
	}
	t.Setenv("PATH", binDir)

	if _, err := Prompt("Select folder"); err == nil || err.Error() != "No folder selected" {
		t.Fatalf("Prompt returned error %v, want No folder selected", err)
	}
}
