package main

import (
	"os"
	"path/filepath"
)

func defaultInstallDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", ".local", "share", "Steam", "steamapps", "common", "Guild Wars 2")
	}
	return filepath.Join(homeDir, ".local", "share", "Steam", "steamapps", "common", "Guild Wars 2")
}

func getInstallPath() (string, error) {
	return defaultInstallDir(), nil
}
