#!/bin/bash

# PRECONDITION: This tests requires gfaserver to be configured.

# When more than one client was connected the data was corrupted. 
# The solution was to create a new thread for receiving the  works that will be processsed.

# This tests creates multiples listeners to check if all the data on all the clients are the same, if not the test have failed
export GFA_IP=$1

TMP_DIR=$(mktemp -d)
# Create 5 listeners
for i in {1..3}; do
    TEMP=$(mktemp -p $TMP_DIR)
    nc -v 172.16.12.251 32001 > $TEMP&
done

python3 blocking_expose.py
sleep 5
md5sum $TMP_DIR/*
MD5_SUM=$(md5sum $TMP_DIR/* | head -n1 | cut -d ' ' -f 1)
TMP_CHECK_FILE=$(mktemp)
for i in $(md5sum $TMP_DIR/* | cut -d ' ' -f 2- | tr -d ' '); do
    echo "$MD5_SUM $i" >> $TMP_CHECK_FILE
done
md5sum --status -c $TMP_CHECK_FILE
EXIT_CODE=$?
rm -r "$TMP_DIR"
exit $EXIT_CODE


