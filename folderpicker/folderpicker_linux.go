package folderpicker

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func pickFolder(msg string) (string, error) {
	if _, err := exec.LookPath("zenity"); err == nil {
		return runPicker("zenity", "--file-selection", "--directory", "--title="+msg)
	}

	if _, err := exec.LookPath("kdialog"); err == nil {
		homeDir, _ := os.UserHomeDir()
		return runPicker("kdialog", "--getexistingdirectory", homeDir, "--title", msg)
	}

	return "", errors.New("no supported GUI folder picker found; install zenity or kdialog")
}

func runPicker(name string, args ...string) (string, error) {
	output, err := exec.Command(name, args...).Output()
	if err == nil {
		return string(output), nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return "", nil
	}
	return "", fmt.Errorf("%s folder picker failed: %w", name, err)
}
