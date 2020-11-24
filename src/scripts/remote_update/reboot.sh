#!/bin/sh
#modprobe -r gfa_e2vccd230
killall gfaserverd gfaserver &> /dev/null
reboot || exit 1
