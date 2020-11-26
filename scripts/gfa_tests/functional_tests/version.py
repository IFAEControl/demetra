#!/usr/bin/python3

from gfaaccesslib.gfa import GFA
import sys

try:
    gfa = GFA(sys.argv[1], 32000)
    ans = gfa.sys.remote_version()
    server = ans.get_ans('version-server')
    lib = ans.get_ans('version-lib')
    module = ans.get_ans('version-module')
    sys.exit(0)
except Exception:
    sys.exit(1)

