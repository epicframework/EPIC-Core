from ctypes import cdll, c_int, c_ulonglong, c_char_p

import time
import json

def runList(ll, interpID, inputList):
    strList = json.dumps(inputList)
    listID = ll.elconn_list_from_json(strList.encode())
    resultID = ll.elconn_call(interpID, listID)
    return resultID

# === load library
ll = cdll.LoadLibrary("../sharedlib/elconn.so")

# === set return types
ll.elconn_get_type.restype = c_int
ll.elconn_init.restype = c_ulonglong
ll.elconn_list_from_json.restype = c_ulonglong
ll.elconn_make_interpreter.restype = c_ulonglong
ll.elconn_call.restype = c_ulonglong
ll.elconn_connect_remote.restype = c_ulonglong
ll.elconn_list_strfirst.restype = c_char_p
ll.elconn_list_to_json.restype = c_char_p

# === set argument types
ll.elconn_list_from_json.argtypes = [c_char_p]
ll.elconn_serve_remote.argtypes = [c_char_p, c_ulonglong]


# == Manual Test 1 == Using the interpreter
initMsg = ll.elconn_init(0)
ll.elconn_display_info(initMsg)

testList = json.dumps(["format", "Hello, %s!", "World"])
listID = ll.elconn_list_from_json(testList.encode())
ll.elconn_list_print(listID)

interpID = ll.elconn_make_interpreter()
resultID = ll.elconn_call(interpID, listID)
ll.elconn_list_print(resultID)


i_manager = ll.elconn_connect_remote(b"http://localhost:3003")

cmd_heartbeat = ll.elconn_list_from_json(json.dumps([
    "heartbeats", "myclient", "beat"]).encode())

ll.elconn_call(i_manager, cmd_heartbeat)
time.sleep(0.5)
ll.elconn_call(i_manager, cmd_heartbeat)
time.sleep(0.5)
ll.elconn_call(i_manager, cmd_heartbeat)
time.sleep(0.5)
ll.elconn_call(i_manager, cmd_heartbeat)
time.sleep(0.5)
ll.elconn_call(i_manager, cmd_heartbeat)
time.sleep(0.5)
ll.elconn_call(i_manager, cmd_heartbeat)
time.sleep(2)
ll.elconn_call(i_manager, cmd_heartbeat)
time.sleep(5)
ll.elconn_call(i_manager, cmd_heartbeat)
