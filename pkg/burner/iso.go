package burner

import (
	"fmt"
	"os"
	"strings"

	"github.com/mudler/luet-makeiso/pkg/schema"
	"github.com/mudler/luet-makeiso/pkg/utils"
	"github.com/pkg/errors"
	"github.com/twpayne/go-vfs"
)

func GenISO(s *schema.SystemSpec, source string, f vfs.FS) error {

	diskImage := s.ISOName()
	label := s.Label
	var bootloader_args string

	// Detect syslinux usage. Probably a new flag for that in schema
	// would be better but for backward compatibility we need some sort
	// of automatic detection.
	if strings.Contains(s.BootFile, "isolinux") {
		bootloader_args = fmt.Sprintf(`\
		  -boot_image isolinux bin_path="%s" \
		  -boot_image isolinux system_area="%s/%s" \
		  -boot_image isolinux partition_table=on \`, s.BootFile, source, s.IsoHybridMBR)
	} else {
		bootloader_args = fmt.Sprintf(`\
		  -boot_image grub bin_path="%s" \
		  -boot_image grub grub2_mbr="%s/%s" \
		  -boot_image grub grub2_boot_info=on`, s.BootFile, source, s.IsoHybridMBR)
	}

	if err := run(fmt.Sprintf(
		`xorriso \
		  -volid "%s" \
		  -joliet on -padding 0 \
		  -outdev "%s" \
		  -map "%s" / -chmod 0755 -- %s \
		  -boot_image any partition_offset=16 \
		  -boot_image any cat_path="%s" \
		  -boot_image any cat_hidden=on \
		  -boot_image any boot_info_table=on \
		  -boot_image any platform_id=0x00 \
		  -boot_image any emul_type=no_emulation \
		  -boot_image any load_size=2048 \
		  -append_partition 2 0xef "%s/boot/uefi.img" \
		  -boot_image any next \
		  -boot_image any efi_path=--interval:appended_partition_2:all:: \
		  -boot_image any platform_id=0xef \
		  -boot_image any emul_type=no_emulation`,
		label, diskImage, source, bootloader_args, s.BootCatalog, source)); err != nil {
		info(err)
		return err
	}

	checksum, err := utils.Checksum(diskImage)
	if err != nil {
		return errors.Wrap(err, "while calculating checksum")
	}

	return f.WriteFile(diskImage+".sha256", []byte(fmt.Sprintf("%s %s", checksum, diskImage)), os.ModePerm)
}
