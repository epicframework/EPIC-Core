# -*- coding: utf-8 -*-
import datetime
from time import sleep
import sys
import json

# Import HA/Connective bindings
from bindings import new_ll, connect

arg_id = sys.argv[1]
arg_re = sys.argv[2]

print(arg_re)

ll = new_ll("../connective/connective/sharedlib/elconn.so")
ll.elconn_init(0)
connective = connect(ll, arg_re.encode())

def main():
    for i in range(1,5):
        #does thing
        # timestamp = datetime.datetime.now()
        # receipt = json.dumps({
        #         "Source": "Test",
        #         "Event": "Literally nothing",
        #         "Command": "It was a comment",
        #         "Time Stamp": str(timestamp),
        #         "Result": "Literally nothing, this is just a test"
        #            })
        # sys.stdout.write(receipt+"\n")
        # sys.stdout.flush()
        connective.runs("heartbeats {arg_id} beat")
        sleep(5)
        
    return
    
if __name__ == "__main__": 
    main()
