package burner

import (
	diskfs "github.com/diskfs/go-diskfs"
	diskpkg "github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/partition/gpt"
	"github.com/mudler/luet-makeiso/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/twpayne/go-vfs"
)

func CreateEFIImage(source, diskImage string, f vfs.FS) error {
	diskImg, err := f.RawPath(diskImage)
	espSize, err := utils.DirSize(source)
	var (
		diskSize         int64 = espSize + 4*1024*1024 // 104 MB
		blkSize          int64 = 512
		partitionStart   int64 = 2048
		partitionSectors int64 = espSize / blkSize
		partitionEnd     int64 = partitionSectors - partitionStart + 1
	)

	// create a disk image
	disk, err := diskfs.Create(diskImg, diskSize, diskfs.Raw)
	if err != nil {
		log.Panic(err)
	}
	// create a partition table
	table := &gpt.Table{
		Partitions: []*gpt.Partition{
			{Start: uint64(partitionStart), End: uint64(partitionEnd), Type: gpt.EFISystemPartition, Name: "EFI System"},
		},
	}
	// apply the partition table
	err = disk.Partition(table)

	spec := diskpkg.FilesystemSpec{Partition: 0, FSType: filesystem.TypeFat32}
	fs, err := disk.CreateFilesystem(spec)

	if err := copyToFS(source, fs, f); err != nil {
		return errors.Wrapf(err, "while copying EFI files")
	}

	return nil
}
