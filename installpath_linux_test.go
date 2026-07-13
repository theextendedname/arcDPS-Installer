package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultInstallDir(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir returned an error: %v", err)
	}
	want := filepath.Join(homeDir, ".local", "share", "Steam", "steamapps", "common", "Guild Wars 2")
	if got := defaultInstallDir(); got != want {
		t.Fatalf("defaultInstallDir returned %q, want %q", got, want)
	}
}

func TestRemoveAddOnsUsesPlatformPaths(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), "Guild Wars 2")
	binDir := filepath.Join(installDir, "bin64")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}

	files := []string{
		filepath.Join(installDir, "d3d11.dll"),
		filepath.Join(binDir, "arcdps_healing_stats.dll"),
	}
	for _, file := range files {
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", file, err)
		}
	}

	removeAddOns(installDir)
	for _, file := range files {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed, got error %v", file, err)
		}
	}
}
