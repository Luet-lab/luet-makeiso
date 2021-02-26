# Luet makeiso

Golang extension to build ISOs with Luet

## Requirements

- mksquashfs

This tool generates ISO reading specfiles in the following syntax:

```yaml
packages:
  # Packages to be installed in the rootfs
  rootfs:
  - utils/busybox 
  # Packages to be installed in the uefi image
  uefi:
  - live/systemd-boot
  - system/mocaccino-live-boot
  # Packages to be installed in the isoimage
  isoimage:
  - live/syslinux
  - system/mocaccino-live-boot

# This configuration isn't necessarly required. You can also just specify the repository to be used in the luet configuration file
repository:
  packages:
  - repository/mocaccino-micro
  - repository/mocaccino-musl-universe
  
# Specify initramfs/kernel and avoid generation on-the-fly
# files must be present on /boot folder in the rootfs
initramfs:
  kernel_file: "bzImage"
  rootfs_file: "rootfs.cpio.xz"

# Image prefix. If Image date is disabled is used as the full title.
image_prefix: "MocaccinoOS-Micro-0."
image_date: true

# Luet config to use.
# It has to contain the repositories required to install the packages defined above.
luet:
  config: conf/luet-micro.yaml
```

## Configuration reference


### `packages.rootfs`

A list of luet packages to install in the rootfs. The rootfs will be squashed to a `rootfs.squashfs` file

### `packages.uefi`

A list of luet packages to be present in the efi ISO sector.

### `packages.isoimage`

A list of luet packages to be present in the ISO image.

### `repository.packages`

A list of package repository (e.g. `repository/mocaccino-extra`) to be installed before `luet install` commands Requirements

### `initramfs.kernel_file`

The kernel file under `/boot/` that is your running  kernel. e.g. `vmlinuz` or `bzImage`

### `initramfs.rootfs_file`

The initrd file under `/boot/` that has all the utils for the initramfs

### `image_prefix`

ISO image prefix to use

### `image_date`

Boolean indicating if the output image name has to contain the date

### `image_name`

A string representing the ISO final image name

### `arch`

A string representing the arch. Defaults to `x86_64`.

### `luet.config`

Path to the luet config to use to install the packages from

## Build

Run `go build` or `make build` from the checkout.

## Install

Download the binary from [the releases](https://github.com/mudler/luet-makeiso/releases) if you haven't compiled locally.

Otherwise it's available in the `luet-official` repo, you can install it with:

```bash
$> luet install extensions/makeiso
```

You don't need anything special than running the binary with the specfile as argument:

```bash

$> luet-makeiso myspec.yaml

```

Note: It respects `TMPDIR` for setting up temporary folders