package folderpicker
//https://github.com/oliverpool/go-folderpicker/blob/master/folderpicker.go
import (
	"errors"
	"path/filepath"
	"strings"
)


// Prompt let the user pick a folder and returns a clean result
func Prompt(msg string) (folder string, err error) {
	folder, err = pickFolder(msg)
	folder = cleanFolder(folder)
	if err == nil && folder == "" {
		//user cancilation of dialog
		err = errors.New("No folder selected")
	}
	return
}

func cleanFolder(s string) string {
	s = strings.TrimSpace(s)
	s = filepath.Clean(s)
	if s == "." || s == `\` {
		return ""
	}
	return s
}

type Prompter interface {
	Prompt(msg string) (string, error)
}
