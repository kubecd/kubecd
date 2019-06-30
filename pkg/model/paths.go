package model

import "path/filepath"

func ResolvePathFromFile(path, file string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(filepath.Dir(file), path)
}

func ResolvePathFromDir(path, dir string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(dir, path)
}
