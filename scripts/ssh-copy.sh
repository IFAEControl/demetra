#!/bin/bash

SRC="$1"
MACHINE="$2"
BITSTREAM="$3"
PASSWORD="$4"
SSH="$5"
NO_QSPI="$6"

SRC="$SRC/$MACHINE"
DEST=$(mktemp -d)

if ! which sshpass; then
   echo "I need sshpass to be installed so I can connect to the ssh server with the password that you told me."
   exit 1
fi

if [ ! -d "$SRC" ]; then
	echo "ERROR: image directory could not be found."
	exit 1
fi

../scripts/common-copy.sh "$DEST" "$SRC" "$MACHINE" "$BITSTREAM" || exit

OLD_DIR="$PWD"
cd "$DEST" || exit
tar -cSf "$DEST/yocto.tar" boot.bin uramdisk devicetree.dtb uImage fpga.bin uEnv.txt
cd "$OLD_DIR" || exit

sshpass -p"$PASSWORD" scp -r "$DEST/yocto.tar" scripts/remote_update/ root@"$SSH":/tmp/
if ! $NO_QSPI; then
   echo "sh /tmp/remote_update/mmc_copy.sh || exit 1" | sshpass -p"$PASSWORD" ssh -t root@"$SSH"
else
   echo "sh /tmp/remote_update/mmc_copy.sh || exit 1" | sshpass -p"$PASSWORD" ssh -t root@"$SSH" || exit
fi
if ! $NO_QSPI; then
    echo "sh /tmp/remote_update/qspi_copy.sh || exit 1" | sshpass -p"$PASSWORD" ssh -t root@"$SSH" || exit
fi
echo "sh /tmp/remote_update/reboot.sh || exit 1" | sshpass -p"$PASSWORD" ssh -t root@"$SSH" || exit


rm -r "$DEST"
