package main

import (
	"os"
	"path"
)

func CreateRunDirectory() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir, err := os.MkdirTemp(wd, "metalkast-bootstrap-*")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path.Join(dir, ".gitignore"), []byte("*\n"), 0644); err != nil {
		return "", err
	}

	return dir, nil
}
