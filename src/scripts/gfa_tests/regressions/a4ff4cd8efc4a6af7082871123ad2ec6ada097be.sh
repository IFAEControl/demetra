#!/bin/bash

# Dispatcher was not checking the size of the packetsize variable, which resulted in a crash or unresponsive application.

# This tests sends an invalid buffer with a big size.

python3 -c 'print("A"*3200)' | nc -v $1 32000
sleep 0.2

# If the server is still up and accepting connections, the test have passed
nc -vz $1 32000

