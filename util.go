package main

import (
	"errors"
	"os"
)

func Min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func ensureDirectoryExists(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
