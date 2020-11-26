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
setup_build_dir || exit 1

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
