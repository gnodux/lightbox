package util

import "path/filepath"

type RootDirectory string

func (r RootDirectory) Abs(path string) string {
	path = filepath.Join(string(r), path)
	if !filepath.IsAbs(path) {
		if newPath, err := filepath.Abs(path); err == nil {
			path = newPath
		}
	}
	return path
}
