package main

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

func defaultInstallDir() string {
	return `C:\Program Files\Guild Wars 2`
}

func getInstallPath() (string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\ArenaNet\Guild Wars 2`, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("error opening registry key: %w", err)
	}
	defer key.Close()

	path, _, err := key.GetStringValue("Path")
	if err != nil {
		return "", fmt.Errorf("error reading registry value: %w", err)
	}

	return filepath.Dir(path), nil
}
