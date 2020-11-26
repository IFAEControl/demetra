#!/bin/bash

export GFA_IP=$1

# Wait for gfaserver to start
echo -n "Waiting for gfaserver"
while ! nc -z -w 1 $GFA_IP 32000; do
    sleep 0.75
    echo -ne "." 
done
echo -ne "\n"
echo "--- STARTING TESTS ---"


FAIL=0

find ./functional_tests/commands -type f \( -name "*.py" -or -name "*.sh" \) -exec bash -c 'echo -ne "Executing {}: "; {} $GFA_IP &> /dev/null && echo "OK" || (echo "FAILED"; FAIL=1)' \;

python3 configure_gfa.py  &>/dev/null
find ./functional_tests/ -maxdepth 1 -type f \( -name "*.py" -or -name "*.sh" \) -exec bash -c 'echo -ne "Executing {}: "; {} $GFA_IP &> /dev/null && echo "OK" || (echo "FAILED"; FAIL=1)' \;
find ./regressions -type f \( -name "*.py" -or -name "*.sh" \) -exec bash -c 'echo -ne "Executing {}: "; {} $GFA_IP &> /dev/null && echo "OK" || (echo "FAILED"; FAIL=1)' \;

#for i in functional_tests/* regressions/* ; do
#    echo -ne "Executing $i: "
#    ./$i $GFA_IP &>/dev/null && echo "OK" || (echo "FAILED"; FAIL=1)
#   ./$i $IP && echo "OK" || echo "FAILED"
#done
exit $FAIL
