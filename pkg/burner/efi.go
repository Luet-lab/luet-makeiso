package burner

import (
	"fmt"
	"github.com/mudler/luet-makeiso/pkg/utils"
	"github.com/pkg/errors"
	"github.com/twpayne/go-vfs"
)

func CreateEFIImage(source, diskImage string, f vfs.FS) error {
	align := int64(4 * 1024 * 1024)
	diskImg, _ := f.RawPath(diskImage)
	diskSize, _ := utils.DirSize(source)

	diskF, err := f.Create(diskImg)
	if err != nil {
		return errors.Wrapf(err, "failed creating image %s", diskImg)
	}

	// Align disk size to the next 4MB slot
	diskSize = diskSize/align*align + align

	err = diskF.Truncate(diskSize)
	if err != nil {
		diskF.Close()
		return errors.Wrapf(err, "failed setting file size to %d bytes", diskSize)
	}
	diskF.Close()

	err = run(fmt.Sprintf("mkfs.fat %s", diskImg))
	if err != nil {
		return errors.Wrapf(err, "failed formatting %s image", diskImg)
	}

	err = run(fmt.Sprintf("mcopy -s -i %s %s/* ::", diskImg, source))
	if err != nil {
		return errors.Wrapf(err, "failed copying '%s' files to image '%s'", source, diskImg)
	}

	return nil
}
