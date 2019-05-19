#!/usr/bin/env python3

# NOTE: This example is not a module; run it as a Python script.
#
#       1. Start Manager2.py
#
#       2. Run this program with the following arguments:
#          ./example.py example http://127.0.0.1:3003
#
#       2.1. Alternatively, ask another group memeber to host HA/Manager
#            and use their host address as the last parameter.

import sys

# Import HA/Connective bindings
from bindings import new_ll, connect, new_interpreter

# CLI arguments
arg_id = sys.argv[1]
arg_re = sys.argv[2]

# Device definition from Mozilla example:
# (https://iot.mozilla.org/wot/#web-thing-rest-api)
true = True
false = False
exampleDeviceDef = {
  "name":"WoT Pi",
  "description": "A WoT-connected Raspberry Pi",
  "properties": {
    "temperature": {
      "title": "Temperature",
      "type": "number",
      "unit": "degree celsius",
      "readOnly": true,
      "description": "An ambient temperature sensor",
      "links": [{"href": "/things/pi/properties/temperature"}]
    },
    "humidity": {
      "title": "Humidity",
      "type": "number",
      "unit": "percent",
      "readOnly": true,
      "links": [{"href": "/things/pi/properties/humidity"}]
    },
    "led": {
      "title": "LED",
      "type": "boolean",
      "description": "A red LED",
      "links": [{"href": "/things/pi/properties/led"}]
    }
  },
  "actions": {
    "reboot": {
      "title": "Reboot",
      "description": "Reboot the device"
    }
  },
  "events": {
    "reboot": {
      "description": "Going down for reboot"
    }
  },
  "links": [
    {
      "rel": "properties",
      "href": "/things/pi/properties"
    },
    {
      "rel": "actions",
      "href": "/things/pi/actions"
    },
    {
      "rel": "events",
      "href": "/things/pi/events"
    },
    {
      "rel": "alternate",
      "href": "wss://mywebthingserver.com/things/pi"
    },
    {
      "rel": "alternate",
      "mediaType": "text/html",
      "href": "/things/pi"
    }
  ]
}

# TODO: uuid for running example multiple times
deviceID = "mydeviceid"
print("The device id will be: "+deviceID)

# Create local interpreter and remote interpreter
ll = new_ll("../connective/connective/sharedlib/elconn.so")
ll.elconn_init(1)
remote_connective = connect(ll, arg_re.encode())
connective = new_interpreter(ll)

# Allow messages to be send to remote interpreter by prefixing
# the command "hub"
ll.elconn_link(b"hub", connective.ii, remote_connective.ii)

# So now the following both work the same:
# remote_connective.runs(": sayhello (store 'hi')")
# connective.runs("hub : sayhello (store 'hi')")

remote_connective.runl([
    # Method path
    'devices', 'add-device',
    # Method parameters
    exampleDeviceDef, {},
    # Custom device ID
    deviceID
])
