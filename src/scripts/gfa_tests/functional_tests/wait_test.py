#!/usr/bin/python3

import sys
from gfaaccesslib.gfa import GFA
import time

try:
    time.sleep(5)
    gfa = GFA(sys.argv[1], 32000)

    ans = gfa.tests.wait(3000)
    if int(ans.elapsed_ms/10) != 300 or ans.get_ans("millisecs") != 3000:
        sys.exit(1)

    gfa.close()
    sys.exit(0)
except Exception as ex:
    print(ex)
    sys.exit(1)
