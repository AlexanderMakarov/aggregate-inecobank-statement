package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// getAbsolutePath checks if a file exists and returns its absolute path.
func getAbsolutePath(filename string) (string, error) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %v", err)
	}

	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %v", absPath)
	} else if err != nil {
		return "", fmt.Errorf("error checking file: %v", err)
	}

	return absPath, nil
}

func getFilesByGlob(glob string) ([]string, error) {
	files, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}
	return files, nil
}
