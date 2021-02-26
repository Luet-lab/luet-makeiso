package burner

import (
	"fmt"
	"os"

	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
	"github.com/mudler/luet-makeiso/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/twpayne/go-vfs"
)

func GenISO(diskImage, label, source string, f vfs.FS) error {
	diskImg, err := f.RawPath(diskImage)

	if diskImg == "" {
		log.Fatal("must have a valid path for diskImg")
	}

	diskSize, err := utils.DirSize(source)
	mydisk, err := diskfs.Create(diskImg, diskSize, diskfs.Raw)
	if err != nil {
		return errors.Wrapf(err, "while creating disk")
	}

	// the following line is required for an ISO, which may have logical block sizes
	// only of 2048, 4096, 8192
	mydisk.LogicalBlocksize = 2048
	fspec := disk.FilesystemSpec{Partition: 0, FSType: filesystem.TypeISO9660, VolumeLabel: label}
	fs, err := mydisk.CreateFilesystem(fspec)
	if err != nil {
		return errors.Wrapf(err, "while creating fs")
	}

	if err := copyToFS(source, fs, f); err != nil {
		return errors.Wrapf(err, "while copying files")
	}

	iso, ok := fs.(*iso9660.FileSystem)
	if !ok {
		return errors.New("not an iso filesystem")
	}

	options := iso9660.FinalizeOptions{
		VolumeIdentifier: label,
		ElTorito: &iso9660.ElTorito{
			BootCatalog: "boot/syslinux/boot.cat",
			Entries: []*iso9660.ElToritoEntry{
				{
					Platform:  iso9660.BIOS,
					Emulation: iso9660.NoEmulation,
					BootFile:  "boot/syslinux/isolinux.bin",
					BootTable: true,
					LoadSize:  4,
				},
				{
					Platform:  iso9660.EFI,
					Emulation: iso9660.NoEmulation,
					BootFile:  "boot/uefi.img",
				},
			},
		},
	}
	if err := iso.Finalize(options); err != nil {
		return err
	}

	checksum, err := utils.Checksum(diskImage)
	if err != nil {
		return errors.Wrap(err, "while calculating checksum")
	}

	return f.WriteFile(diskImage+".sha256", []byte(fmt.Sprintf("%s %s", checksum, diskImage)), os.ModePerm)
}
