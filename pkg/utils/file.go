package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
	"github.com/twpayne/go-vfs"
)

func CopyContent(src, dst string) error {
	opt := copy.Options{
		OnSymlink: func(string) copy.SymlinkAction { return copy.Shallow },
		Sync:      true,
	}
	return copy.Copy(src, dst, opt)
}

func CopyFile(src, dst string, fs vfs.FS) error {
	_, err := fs.Stat(src)
	if err != nil {
		return err
	}

	source, err := fs.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := fs.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func Checksum(source string) (string, error) {
	f, err := os.Open(source)
	if err != nil {
		return "", nil
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", nil
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
