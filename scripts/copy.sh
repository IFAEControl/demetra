#!/bin/bash

DEST="$1"
SRC="$2"
DEVICE="$3"
MACHINE="$4"
BITSTREAM="$5"
CLEAN="$6"

SRC="$SRC/$MACHINE"

if [ ! -d "$DEST" ]; then
	echo "Creating directory $DEST"
	mkdir "$DEST" || exit 1
fi

if [ ! -d "$SRC" ]; then
	echo "ERROR: image directory $SRC could not be found. PWD=$PWD"
	exit 1
fi


if [ "$DEVICE" != "" ]; then
	mount "$DEVICE" "$DEST" || exit 1
fi


if $CLEAN; then
	#Clean old files
	/bin/rm -f "$DEST"/*
fi	

../scripts/common-copy.sh "$DEST" "$SRC" "$MACHINE" "$BITSTREAM" || exit

if [ "$DEVICE" != "" ]; then
   udisksctl unmount -b "$DEVICE"
   udisksctl power-off -b "$DEVICE"
fi
