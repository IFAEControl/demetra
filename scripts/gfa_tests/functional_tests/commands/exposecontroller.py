#!/usr/bin/python3

from gfaaccesslib.gfa import GFA
import sys
import traceback
from py_fuzz_testing import fuzz_class

try:
    gfa = GFA(sys.argv[1], 32000)
    fuzz_class(gfa.exposecontroller, ["remote_start_stack_exec"])
    sys.exit(0)
except Exception:
    traceback.print_exc()
    sys.exit(1)

