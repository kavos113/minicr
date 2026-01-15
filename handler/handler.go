package handler

import (
	"os"
	"path/filepath"
)

var (
	rootPath  = "./data"
	uploadDir = filepath.Join(rootPath, "uploads")
	blobDir   = filepath.Join(rootPath, "blobs")
	tagDir    = filepath.Join(rootPath, "tag")
)

func InitDirs() error {
	dirs := []string{rootPath, uploadDir, blobDir, tagDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
