package burner

import (
	"fmt"

	"github.com/mudler/luet-geniso/pkg/utils"
	"github.com/pkg/errors"
	"github.com/twpayne/go-vfs"

	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/filesystem/fat32"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
	"github.com/diskfs/go-diskfs/filesystem/squashfs"
)

func CreateFilesystem(d *disk.Disk, spec disk.FilesystemSpec) (filesystem.FileSystem, error) {
	// find out where the partition starts and ends, or if it is the entire disk
	var (
		size, start int64
	)
	switch {
	case spec.Partition == 0:
		size = d.Size
		start = 0
	case d.Table == nil:
		return nil, fmt.Errorf("cannot create filesystem on a partition without a partition table")
	default:
		partitions := d.Table.GetPartitions()
		// API indexes from 1, but slice from 0
		partition := spec.Partition - 1
		if spec.Partition > len(partitions) {
			return nil, fmt.Errorf("cannot create filesystem on partition %d greater than maximum partition %d", spec.Partition, len(partitions))
		}
		size = partitions[partition].GetSize()
		start = partitions[partition].GetStart()
	}

	switch spec.FSType {
	case filesystem.TypeFat32:
		return fat32.Create(d.File, size, start, d.LogicalBlocksize, spec.VolumeLabel)
	case filesystem.TypeISO9660:
		return iso9660.Create(d.File, size, start, d.LogicalBlocksize, spec.WorkDir)
	case filesystem.TypeSquashfs:
		return squashfs.Create(d.File, size, start, d.LogicalBlocksize)
	default:
		return nil, errors.New("Unknown filesystem type requested")
	}
}

// XXX: This doesn't work still
func nativeSquashfs(diskImage, label, source string, f vfs.FS) error {

	diskImg, err := f.RawPath(diskImage)

	if diskImg == "" {
		return errors.New("must have a valid path for diskImg")
	}
	size, err := utils.DirSize(source)
	var diskSize int64 = size // 10 MB
	mydisk, err := diskfs.Create(diskImg, diskSize, diskfs.Raw)
	if err != nil {
		return errors.Wrapf(err, "while creating disk")
	}

	mydisk.LogicalBlocksize = 4096
	fspec := disk.FilesystemSpec{Partition: 0, FSType: filesystem.TypeSquashfs, VolumeLabel: label}
	fs, err := CreateFilesystem(mydisk, fspec)
	if err != nil {
		return errors.Wrapf(err, "while creating squashfs size: %d", size)
	}

	if err := copyToFS(source, fs, f); err != nil {
		return errors.Wrapf(err, "while copying files")
	}

	sqs, ok := fs.(*squashfs.FileSystem)
	if !ok {
		return errors.Wrapf(err, "not a squashfs")
	}

	return sqs.Finalize(squashfs.FinalizeOptions{})
}

func CreateSquashfs(diskImage, label, source string, f vfs.FS) error {
	_, err := run(fmt.Sprintf("mksquashfs %s %s -b 1024k -comp xz -Xbcj x86", source, diskImage))
	//log.Info(l)
	return err
}
