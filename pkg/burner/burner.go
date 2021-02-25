package burner

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mudler/luet-geniso/pkg/schema"
	"github.com/twpayne/go-vfs"
)

func findKernel(s string) (string, error) {

}

func Burn(s *schema.SystemSpec, fs vfs.FS) error {
	dir, err := ioutil.TempDir("", "luet-geniso")
	defer os.RemoveAll(dir)

	tempRootfs := filepath.Join(dir, "rootfs")
	tempOverlayfs := filepath.Join(dir, "overlayfs")
	tempKernel := filepath.Join(dir, "kernel")

	if err := vfs.MkdirAll(fs, tempRootfs, os.ModePerm); err != nil {
		return err
	}
	if err := vfs.MkdirAll(fs, tempOverlayfs, os.ModePerm); err != nil {
		return err
	}
	if err := vfs.MkdirAll(fs, tempKernel, os.ModePerm); err != nil {
		return err
	}

	if s.Overlay {
		if len(s.Packages.Initramfs) > 0 {
			if err := LuetInstall(tempRootfs, s.Packages.Initramfs, s.Repository.Initramfs, false, fs, s); err != nil {
				return err
			}
		}
		if err := LuetInstall(tempOverlayfs, s.Packages.Rootfs, s.Repository.Packages, s.Packages.KeepLuetDB, fs, s); err != nil {
			return err
		}
	} else {
		if err := LuetInstall(tempRootfs, s.Packages.Initramfs, s.Repository.Initramfs, s.Packages.KeepLuetDB, fs, s); err != nil {
			return err
		}
	}
	return err
}
