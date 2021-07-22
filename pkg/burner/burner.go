package burner

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyokomi/emoji/v2"
	"github.com/mudler/luet-makeiso/pkg/schema"
	"github.com/mudler/luet-makeiso/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/twpayne/go-vfs"
)

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

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

func ensureDirs(path string) {
	modes := map[os.FileMode][]string{
		0755: {"/dev", "/run"},
		0555: {"/sys", "/proc"},
		1777: {"/tmp"},
	}
	for mode, paths := range modes {
		for _, p := range paths {
			cpath := filepath.Join(path, p)
			if _, err := os.Stat(cpath); err != nil && os.IsNotExist(err) {
				info(fmt.Sprintf("dir '%s' missing from rootfs. creating with %+v", cpath, mode))
				os.MkdirAll(cpath, mode)
			}
		}
	}
}

func prepareRootfs(s *schema.SystemSpec, fs vfs.FS, tempOverlayfs string) error {
	if s.RootfsImage != "" {
		image := s.RootfsImage
		if !strings.Contains(image, ":") {
			image = image + ":latest"
		}
		// Check if we have it locally in docker first, uncompress the image using luet
		if contains(DockerImages(), image) {
			info(fmt.Sprintf("Image '%s' found locally, using it", image))
			if err := DockerExtract(image, tempOverlayfs); err != nil {
				return err
			}
		} else {
			info(":steaming_bowl: Downloading container image")
			if err := LuetImageUnpack(s.RootfsImage, tempOverlayfs); err != nil {
				return err
			}
		}
	} else if len(s.Packages.Rootfs) > 0 {
		info(":steaming_bowl: Installing luet packages")
		if err := LuetInstall(tempOverlayfs, s.Packages.Rootfs, s.Repository.Packages, s.Packages.KeepLuetDB, fs, s); err != nil {
			return err
		}
	}

	if s.Overlay.Rootfs != "" {
		info(":steaming_bowl: Adding files to rootfs from overlay")
		if err := utils.CopyContent(s.Overlay.Rootfs, tempOverlayfs); err != nil {
			return err
		}
	}

	if s.EnsureCommonDirs {
		ensureDirs(tempOverlayfs)
	}
	return nil
}

func prepareUEFI(s *schema.SystemSpec, fs vfs.FS, tempISO, tempUEFI, kernelFile, initrdFile string) error {

	if s.UEFIImage == "" {
		// Generate efi image
		info(":superhero: Installing EFI packages")
		if err := LuetInstall(tempUEFI, s.Packages.UEFI, s.Repository.Packages, false, fs, s); err != nil {
			return err
		}

		info(":superhero:Copying EFI kernels")
		if err := vfs.MkdirAll(fs, filepath.Join(tempUEFI, "minimal", s.Arch), os.ModePerm); err != nil {
			return err
		}

		if err := utils.CopyFile(kernelFile, filepath.Join(tempUEFI, "minimal", s.Arch, "kernel.xz"), fs); err != nil {
			return err
		}

		if err := utils.CopyFile(initrdFile, filepath.Join(tempUEFI, "minimal", s.Arch, "rootfs.xz"), fs); err != nil {
			return err
		}

		if err := vfs.MkdirAll(fs, filepath.Join(tempISO, "boot"), os.ModePerm); err != nil {
			return err
		}

		if s.Overlay.UEFI != "" {
			info(":steaming_bowl: Adding files to EFI from overlay")
			if err := utils.CopyContent(s.Overlay.UEFI, tempUEFI); err != nil {
				return err
			}
		}

		info(":superhero:Creating EFI image")
		if err := CreateEFIImage(tempUEFI, filepath.Join(tempISO, "boot", "uefi.img"), fs); err != nil {
			return err
		}
	} else {
		info("copying EFI image from", s.UEFIImage)
		str, err := fs.RawPath(filepath.Join(tempISO, "boot", "uefi.img"))
		if err != nil {
			return err
		}

		os.MkdirAll(filepath.Dir(str), os.ModePerm)

		if _, err := os.Stat(str); os.IsNotExist(err) {
			file, err := os.Create(str)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
		}

		if _, err := copy(s.UEFIImage, str); err != nil {
			return err
		}
	}
	return nil
}

func prepareISO(s *schema.SystemSpec, fs vfs.FS, tempISO, tempOverlayfs, kernelFile, initrdFile string) error {
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

	if s.Overlay.IsoImage != "" {
		info(":steaming_bowl: Adding files to ISO from overlay")
		if err := utils.CopyContent(s.Overlay.IsoImage, tempISO); err != nil {
			return err
		}
	}
	return nil
}

func Burn(s *schema.SystemSpec, fs vfs.FS) error {

	if s.RootfsImage == "" && len(s.Packages.Rootfs) == 0 && len(s.Overlay.Rootfs) == 0 {
		return errors.New("No container image, packages or overlay specified in the yaml file")
	}

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

	info(":steaming_bowl: Installing Overlay packages")
	if err := prepareRootfs(s, fs, tempOverlayfs); err != nil {
		return err
	}

	kernelFile := filepath.Join(tempOverlayfs, "boot", s.Initramfs.KernelFile)
	initrdFile := filepath.Join(tempOverlayfs, "boot", s.Initramfs.RootfsFile)

	if err := prepareUEFI(s, fs, tempISO, tempUEFI, kernelFile, initrdFile); err != nil {
		return err
	}

	if err := prepareISO(s, fs, tempISO, tempOverlayfs, kernelFile, initrdFile); err != nil {
		return err
	}

	info(fmt.Sprintf(":tropical_drink:Generate ISO %s", s.ISOName()))
	if _, err := fs.Stat(s.ISOName()); err == nil {
		// Remove iso if already present
		fs.RemoveAll(s.ISOName())
	}

	return GenISO(s.ISOName(), s.Label, tempISO, fs)
}
