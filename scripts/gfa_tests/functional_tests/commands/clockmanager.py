#!/usr/bin/python3

from gfaaccesslib.gfa import GFA
import sys
import traceback
from py_fuzz_testing import fuzz_class

try:
    gfa = GFA(sys.argv[1], 32000)
    fuzz_class(gfa.clockmanager, ["remote_start_time_thread", "remote_join_time_thread"])
    sys.exit(0)
except Exception:
    traceback.print_exc()
    sys.exit(1)

