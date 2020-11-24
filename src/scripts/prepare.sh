#!/bin/bash

source scripts/helper_functions.sh

MACHINE="$1"
MODULE_DIR="$2"
LIBRARY_DIR="$3"
SERVER_DIR="$4"
XADC_TEST_DIR="$5"
MCP_DIR="$6"
EXTERNAL="$7"

# If it is the first time we execute it, we will need to create the build directory
if [ ! -d "poky/build" ]; then
    bash -c "cd poky; source ./oe-init-build-env > /dev/null" || exit 1
fi

if [ ! -d "poky/build" ]; then
    echo "Error when creating poky build directory"
    exit 1
fi

checkout_machine "$MACHINE"

if $EXTERNAL; then
	cp resources/local.conf /tmp/gfa-yocto-local.conf
	echo 'INHERIT += "externalsrc"' >> /tmp/gfa-yocto-local.conf
#	echo "EXTERNALSRC_pn-gfa-module = \"$MODULE_DIR\"" >> /tmp/gfa-yocto-local.conf
	echo "EXTERNALSRC_pn-gfa-library = \"$LIBRARY_DIR\"" >> /tmp/gfa-yocto-local.conf
	echo "EXTERNALSRC_pn-gfa-server = \"$SERVER_DIR\"" >> /tmp/gfa-yocto-local.conf
	echo "EXTERNALSRC_pn-gfa-xadc-test = \"$XADC_TEST_DIR\"" >> /tmp/gfa-yocto-local.conf
	echo "EXTERNALSRC_pn-gfa-mcp11aa02e48 = \"$MCP_DIR\"" >> /tmp/gfa-yocto-local.conf
	cp /tmp/gfa-yocto-local.conf poky/build/conf/local.conf
else
	cp resources/local.conf poky/build/conf/local.conf 
fi

cd poky || exit

check_layer meta-dev
check_layer meta-python
check_layer meta-oe
check_layer meta-networking
check_layer meta-gfa
check_layer meta-xilinx/meta-xilinx-bsp
check_layer meta-enclustra

