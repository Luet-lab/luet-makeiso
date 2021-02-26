package burner

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"
	"github.com/kyokomi/emoji/v2"
	"github.com/mudler/luet-makeiso/pkg/schema"
	"github.com/mudler/luet-makeiso/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/twpayne/go-vfs"
)

func prepareWorkDir(fs vfs.FS, dirs ...string) error {
	for _, d := range dirs {
		if err := vfs.MkdirAll(fs, d, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func info(a ...interface{}) {
	log.Info(emoji.Sprint(a...))
}

func Burn(s *schema.SystemSpec, fs vfs.FS) error {

	dir, err := ioutil.TempDir("", "luet-geniso")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	if s.Arch == "" {
		s.Arch = "x86_64"
	}

	tempRootfs := filepath.Join(dir, "rootfs")
	tempOverlayfs := filepath.Join(dir, "overlayfs")
	tempUEFI := filepath.Join(dir, "tempUEFI")
	tempISO := filepath.Join(dir, "tempISO")

	defer fs.RemoveAll(tempRootfs)
	defer fs.RemoveAll(tempOverlayfs)
	defer fs.RemoveAll(tempUEFI)
	defer fs.RemoveAll(tempISO)

	info(":mag: Preparing folders")
	if err := prepareWorkDir(fs, tempRootfs, tempOverlayfs, tempUEFI, tempISO); err != nil {
		return err
	}

	sp := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	sp.Start()
	defer sp.Stop()

	info(":superhero: Installing EFI packages")
	if err := LuetInstall(tempUEFI, s.Packages.UEFI, s.Repository.Packages, false, fs, s); err != nil {
		return err
	}

	info(":steaming_bowl: Installing Overlay packages")
	if err := LuetInstall(tempOverlayfs, s.Packages.Rootfs, s.Repository.Packages, s.Packages.KeepLuetDB, fs, s); err != nil {
		return err
	}

	info(":superhero:Copying EFI kernels")
	if err := vfs.MkdirAll(fs, filepath.Join(tempUEFI, "minimal", s.Arch), os.ModePerm); err != nil {
		return err
	}

	kernelFile := filepath.Join(tempOverlayfs, "boot", s.Initramfs.KernelFile)
	initrdFile := filepath.Join(tempOverlayfs, "boot", s.Initramfs.RootfsFile)

	if err := utils.CopyFile(kernelFile, filepath.Join(tempUEFI, "minimal", s.Arch, "kernel.xz"), fs); err != nil {
		return err
	}

	if err := utils.CopyFile(initrdFile, filepath.Join(tempUEFI, "minimal", s.Arch, "rootfs.xz"), fs); err != nil {
		return err
	}

	if err := vfs.MkdirAll(fs, filepath.Join(tempISO, "boot"), os.ModePerm); err != nil {
		return err
	}

	info(":superhero:Creating EFI image")
	if err := CreateEFIImage(tempUEFI, filepath.Join(tempISO, "boot", "uefi.img"), fs); err != nil {
		return err
	}

	info(":thinking:Populating ISO folder")
	if err := LuetInstall(tempISO, s.Packages.IsoImage, s.Repository.Packages, false, fs, s); err != nil {
		return err
	}

	info(":superhero:Copying BIOS kernels")
	if err := utils.CopyFile(kernelFile, filepath.Join(tempISO, "boot", "kernel.xz"), fs); err != nil {
		return err
	}

	if err := utils.CopyFile(initrdFile, filepath.Join(tempISO, "boot", "rootfs.xz"), fs); err != nil {
		return err
	}

	info(":tv:Create squashfs")
	if err := CreateSquashfs(filepath.Join(tempISO, "rootfs.squashfs"), "squashfs", tempOverlayfs, fs); err != nil {
		return err
	}

	info(fmt.Sprintf(":tropical_drink:Generate ISO %s", s.ISOName()))
	if _, err := fs.Stat(s.ISOName()); err == nil {
		// Remove iso if already present
		fs.RemoveAll(s.ISOName())
	}

	return GenISO(s.ISOName(), s.Label, tempISO, fs)
}
