#!/usr/bin/python3

import sys
from gfaaccesslib.gfa import GFA
try:
    gfa = GFA(sys.argv[1], 32000)

    s = ""
    gfa.tests.echo("")
    """
    for j in range(1500):
        s += chr(j)
    for i in range(100):
        ans = gfa.tests.echo(s)
        if not s == ans._json_dict["arguments"].get("message", None) or ans.get_ans("warm") != "kitty":
            sys.exit(1)
    """

    gfa.close()
    sys.exit(0)
except Exception as ex:
    import traceback
    traceback.print_exc()
    sys.exit(1)
