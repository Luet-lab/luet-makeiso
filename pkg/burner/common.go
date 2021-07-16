package burner

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/twpayne/go-vfs"
)

func CopyFile(src, dst string, f filesystem.FileSystem) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "failed opening %s", src)
	}
	defer in.Close()

	out, err := f.OpenFile(dst, os.O_CREATE|os.O_RDWR)
	if err != nil {
		return errors.Wrapf(err, "failed opening for writing %s", dst)
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return errors.Wrapf(err, "failed copying %s to %s", src, dst)
	}

	return
}

func contains(list []string, s string) bool {
	for _, elem := range list {
		if elem == s {
			return true
		}
	}
	return false
}

func runO(cmd string, opts ...func(cmd *exec.Cmd)) (string, error) {

	log.Debugf("running command `%s`", cmd)
	c := exec.Command("sh", "-c", cmd)
	c.Env = []string{fmt.Sprintf("PATH=%s", os.Getenv("PATH"))}

	for _, o := range opts {
		o(c)
	}

	out, err := c.CombinedOutput()

	return string(out), err
}

func run(cmd string, opts ...func(cmd *exec.Cmd)) error {
	log.Debugf("running command `%s`", cmd)
	c := exec.Command("sh", "-c", cmd)
	for _, o := range opts {
		o(c)
	}
	c.Env = []string{fmt.Sprintf("PATH=%s", os.Getenv("PATH")), "LUET_NOLOCK=true", "LUET_YES=true"}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Start()
	if err != nil {
		return fmt.Errorf("failed to run %s: %v", cmd, err)
	}

	return c.Wait()
}

func runEnv(cmd string, opts ...func(cmd *exec.Cmd)) error {
	log.Debugf("running command `%s`", cmd)
	c := exec.Command("sh", "-c", cmd)
	for _, o := range opts {
		o(c)
	}
	c.Env = os.Environ()
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Start()
	if err != nil {
		return fmt.Errorf("failed to run %s: %v", cmd, err)
	}

	return c.Wait()
}

func CopyDir(src string, dst string, f filesystem.FileSystem) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return errors.Wrapf(err, "failed stat of source: %s", src)
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	err = f.Mkdir(dst)
	if err != nil {
		return errors.Wrapf(err, "failed mkdir of dst: %s", dst)
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return errors.Wrapf(err, "failed read of src: %s", src)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath, f)
			if err != nil {
				return errors.Wrapf(err, "failed copy: %s", srcPath)
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath, f)
			if err != nil {
				return errors.Wrapf(err, "failed copy file: %s", srcPath)
			}
		}
	}

	return nil
}
func copyToFS(s string, f filesystem.FileSystem, fs vfs.FS) error {
	return CopyDir(s, "/", f)
}
