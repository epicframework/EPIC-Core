from ctypes import cdll, c_int, c_ulonglong, c_char_p
import json

class Interpreter:
    def __init__(self, ll, ii):
        self.ll = ll
        self.ii = ii
    def runs(self, text, tolist=False):
        listID   = self.ll.elconn_list_from_text(text.encode())
        resultID = self.ll.elconn_call(self.ii, listID)
        # Report ID of data (default behaviour)
        if tolist == False:
            return resultID
        # Report list as native Python type
        jsonText = self.ll.elconn_list_to_json(resultID)
        return json.loads(str(jsonText, "utf-8"))
    def runl(self, inputList, tolist=False):
        print("{{{ INPUT LIST }}}")
        print(inputList)
        strList = json.dumps(inputList)
        listID   = self.ll.elconn_list_from_json(strList.encode())
        resultID = self.ll.elconn_call(self.ii, listID)
        # Report ID of data (default behaviour)
        if tolist == False:
            return resultID
        # Report list as native Python type
        jsonText = self.ll.elconn_list_to_json(resultID)
        return json.loads(str(jsonText, "utf-8"))
    def serve_remote(self, addr):
        self.ll.elconn_serve_remote(addr, self.ii)

def new_interpreter(ll):
    """Create a new local interpreter."""
    ii = ll.elconn_make_interpreter()
    return Interpreter(ll, ii)

def connect(ll, addr):
    """Create a new interpreter connected to a server."""
    ii = ll.elconn_connect_remote(addr)
    return Interpreter(ll, ii)

def new_ll(libloc):
    # === load library
    ll = cdll.LoadLibrary(libloc)

    # === set return types
    ll.elconn_get_type.restype = c_int
    ll.elconn_init.restype = c_int
    ll.elconn_list_from_json.restype = c_int
    ll.elconn_make_interpreter.restype = c_int
    ll.elconn_call.restype = c_int
    ll.elconn_connect_remote.restype = c_int
    ll.elconn_list_strfirst.restype = c_char_p
    ll.elconn_list_to_json.restype = c_char_p

    # === set argument types
    ll.elconn_list_from_json.argtypes = [c_char_p]
    ll.elconn_serve_remote.argtypes = [c_char_p, c_int]
    ll.elconn_link.argtypes = [c_char_p, c_int, c_int]

    return ll
