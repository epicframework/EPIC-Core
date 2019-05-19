from ctypes import cdll, c_int, c_ulonglong, c_char_p

import time
import json

def runList(ll, interpID, inputList):
    strList = json.dumps(inputList)
    listID = ll.elconn_list_from_json(strList.encode())
    resultID = ll.elconn_call(interpID, listID)
    return resultID

def run(ll, interpID, text):
    listID   = ll.elconn_list_from_text(text.encode())
    resultID = ll.elconn_call(interpID, listID)
    return resultID

# === load library
ll = cdll.LoadLibrary("../sharedlib/elconn.so")

# === set return types
ll.elconn_get_type.restype = c_int
ll.elconn_init.restype = c_ulonglong
ll.elconn_list_from_json.restype = c_ulonglong
ll.elconn_list_from_text.restype = c_ulonglong
ll.elconn_make_interpreter.restype = c_ulonglong
ll.elconn_call.restype = c_ulonglong
ll.elconn_connect_remote.restype = c_ulonglong
ll.elconn_list_strfirst.restype = c_char_p
ll.elconn_list_to_json.restype = c_char_p

# === set argument types
ll.elconn_list_from_json.argtypes = [c_char_p]
ll.elconn_list_from_text.argtypes = [c_char_p]
ll.elconn_serve_remote.argtypes = [c_char_p, c_ulonglong]


# == Manual Test 1 == Using the interpreter
initMsg = ll.elconn_init(0)
ii = ll.elconn_make_interpreter()
ll.elconn_serve_remote(b":3003", ii)

ll.elconn_display_info(initMsg)

# Create directory for heartbeat monitors
run(ll, ii, ": heartbeats (@ directory)")
# Create a heartbeat monitor called "myclient"
run(ll, ii, "heartbeats : myclient (@ heartbeat-monitor 1s)")
# Send an initial heartbeat on behalf of the client
run(ll, ii, "heartbeats myclient beat")

for x in range (100):
    time.sleep(0.2)
    result = run(ll, ii, "heartbeats myclient time-since")
    ll.elconn_list_print(result)
