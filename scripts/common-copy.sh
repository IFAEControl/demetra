#!/bin/bash

DEST="$1"
SRC="$2"

# mercury-zx5 and zc702-zynq7 supported
MACHINE="$3"

BITSTREAM="$4"

#Explained in https://github.com/Xilinx/meta-xilinx/blob/master/README.booting.md#loading-via-sd


# Generate boot.bin for enclustra
if [ "$MACHINE" == "mercury-zx5" ]; then
	cp "$SRC/u-boot.elf" resources/binaries || exit 1
	cp "$BITSTREAM" resources/binaries/fpga.bit || exit 1
	(
  	cd resources/binaries || exit
  	mkbootimage boot.bif /tmp/boot.bin
	)
	cp /tmp/boot.bin "$SRC/boot.bin" || exit 1
fi

# From microzed
#cp "$SRC/u-boot.img" "$DEST"

cp "$SRC/boot.bin" "$DEST" || exit 1
cp "$SRC/uImage" "$DEST" || exit 1

# Convert bit to bin. bit format is not compatible
python ../resources/fpga-bit-to-bin.py --flip  "$BITSTREAM" "$DEST/fpga.bin"

if [ "$MACHINE" == "mercury-zx5" ]; then
	cp "$SRC/$MACHINE.dtb" "$DEST/devicetree.dtb"
	cp "$SRC/core-image-minimal-$MACHINE.cpio.gz.u-boot" "$DEST/uramdisk" || exit 1
	
	cp resources/uEnv.txt "$DEST/uEnv.txt" || exit 1

	# Check sizes
	./scripts/check_image_files_size.py "$DEST" || exit
fi


if [ "$MACHINE" == "zc702-zynq7" ]; then
	cp "$SRC/zynq-zc702.dtb" "$DEST" || exit 1
	cp "$SRC/boot.scr" "$DEST" || exit 1
	cp "$SRC/u-boot.img" "$DEST" || exit 1

	# Couple of hacks to be able to load the FPGA
	cat "$SRC/uEnv.txt" | tr -d '\t' | sed 's/bitstream_image=boot.bin/bitstream_image=fpga.bin/' > "$DEST/uEnv.txt" || exit 1
	echo "bootcmd=run loadfpga && run distro_bootcmd" >> "$DEST/uEnv.txt" || exit 1
	#cp "$SRC/uEnv.txt" "$DEST" || exit 1
	cp "$SRC/core-image-minimal-$MACHINE.cpio.gz.u-boot" "$DEST/uramdisk.image.gz" || exit 1
fi