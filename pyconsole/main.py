#!/usr/bin/env python3
import datetime
from time import sleep
import numpy as np
import sys
import json

# Import HA/Connective bindings
from bindings import new_ll, connect, new_interpreter

arg_id = sys.argv[1]
arg_re = sys.argv[2]

print(arg_re)

ll = new_ll("../connective/connective/sharedlib/elconn.so")
ll.elconn_init(1)
remote_connective = connect(ll, arg_re.encode())
connective = new_interpreter(ll)

ll.elconn_link(b"hub", connective.ii, remote_connective.ii)

def main():
    while True:
        print(">", end='')
        cmd = input()
        if cmd == 'quit':
            return
        result = connective.runs(cmd, tolist=True)
        print(result)
    
if __name__ == "__main__": 
    main()

