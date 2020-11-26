#!/bin/bash

# PRECONDITION: This test requires gfa to be configured

# Check fd leak when executing commands. The socket instance was not deleted so the underlying fd was not closed.
# As the bug triggers on commands only, execute a lot of commands and check if it is still running
# A better aproach is to get how many file descriptors are opened when a command is executed.

(
for i in {1..200}; do
     python3 echotest_2.py $GFA_IP || exit
done
)





