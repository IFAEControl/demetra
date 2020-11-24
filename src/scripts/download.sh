#!/bin/bash

source scripts/helper_functions.sh

RELEASE="$1"
export RELEASE


# Clone poky repo
clone git://git.yoctoproject.org/poky

if [ ! -d "poky" ]; then
    echo "Poky directory doesn't exist"
    exit 1
fi

cd poky || exit

clone git@gitlab.pic.es:DESI-GFA/yocto/meta-dev.git
clone git@gitlab.pic.es:DESI-GFA/yocto/meta-gfa.git

# Clone meta-xilinx
clone git://git.yoctoproject.org/meta-xilinx
#apply_patch meta-xilinx xilinx-workarounds-sumo.patch || exit 1

clone git@gitlab.pic.es:DESI-GFA/meta-enclustra.git

# Clone and set-up misc repos (used by chrony and tcpdump)
clone git://git.openembedded.org/meta-openembedded
if [ ! -d "../meta-oe" ]; then
	ln -s meta-openembedded/meta-oe . || exit
fi
if [ ! -d "../meta-python" ]; then
	ln -s meta-openembedded/meta-python . || exit
fi
if [ ! -d "../meta-networking" ]; then 
	ln -s meta-openembedded/meta-networking . || exit
fi
