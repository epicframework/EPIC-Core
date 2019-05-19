# -*- coding: utf-8 -*-
import yaml
import time
import json
from Module import Module, CommandNotRecognized
from bindings import new_ll, new_interpreter
from threading import Thread

linebuffer = []
receipt_list = []
processes = []
recover = []

#Test Code
config = {}

# Import HA/Connective bindings
ll = new_ll("../connective/connective/sharedlib/elconn.so")

class PackageEventThread(Thread):
    def __init__(self, connective):
        Thread.__init__(self)
        self.connective = connective
    def run(self):
        config = self.connective.runs('events new-package block',
            tolist=True)
        config = config[0]
        package_name = config['package_id']
        for command in config['commands']:
            id = "%s-%s" % (command, package_name)
            cmd = config['commands'][command]['cmd']
            print("*** Loading package module: " + id)
            # Add data structures for process management
            # TODO(eric): Update when parameter binding is added to HA/Connective
            self.connective.runs("heartbeats : '"+id+"' (@ heartbeat-monitor 1s)")
            try:
                process = Module(id, cmd, linebuffer, package=True, package_name=package_name)
                processes.append(process)
            except CommandNotRecognized:
                #TODO: Handle this
                print("Command not recognized")

# Main Loop
def main():
    # Gets configuration from config.yml
    # Gets configuration from config.yml
    print("Accessing configuration file")
    with open("./config.yml", 'r') as stream:
        try:
            config = yaml.load(stream)
        except yaml.YAMLError as exc:
            # TODO: Exit on failure to open config, write to log
            print(exc)

    # Get Config Details
    #
    # TODO: Confirm method for converting string to byte literal
    #       Replaces b":3111"
    #
    connectivePort     = config["config"]["port"]
    connectiveAddress  = config["config"]["address"]
    heartbeatThreshold = config["config"]["heartbeat_threshold"]

    # Init HA/Connective Server
    initMsg = ll.elconn_init(1)
    ll.elconn_display_info(initMsg)
    connective = new_interpreter(ll)
    connective.serve_remote(connectivePort)

    # TODO: Confirm configuration via config file functions correctly
    #       Replaces Code:
    #
    #       # Initialize Management Data Structures
    #       connective.runs(": heartbeats (@ directory)")
    #
    #       # Initialize event queue directory
    #       connective.runs(": events (@ directory)")
    #       connective.runs("events : new-package (@ requests)")
    #
    #       # Iniitalize IoT Data Structures
    #       connective.runs(": devices (@ directory)")
    #       connective.runs("include device ($ devices)")

    # Initialize Connective Data Structures
    print("Beginning Connective Configuration")
    for connectiveConfig in config["connective"]:
        cmd  = connectiveConfig["cmd"]
        name = connectiveConfig["name"]
        msg  = connectiveConfig["msg"]

        if cmd[0] == "run":
            print("Initializing connective with command: runs "+cmd[1])
            connective.runs(cmd[1])
    print("Connective Configuration Complete")

    # Define system according to configuration
    print("Starting subprocesses")
    for componentConfig in config['components']:
        cmd      = componentConfig['cmd']
        name     = componentConfig['name']
        id       = componentConfig['id']
        recovery = componentConfig["recovery"]
        print("Executing process "+name+" with command "+str(cmd))

        # re-write attributes in cmd to include remote address and app id
        for x in range(len(cmd[1])):
            cmd[1][x] = cmd[1][x].replace('<id>', id)
            cmd[1][x] = cmd[1][x].replace('<remote>', connectiveAddress)

        print("exe: ",cmd)

        # Add data structures for process management
        # TODO(eric): Update when parameter binding is added to HA/Connective
        connective.runs("heartbeats : '"+id+"' (@ heartbeat-monitor 1s)")

        # Start Process
        try:
            process = Module(id, cmd, recovery, linebuffer)
            processes.append(process)
        except CommandNotRecognized:
            # TODO: Handle this
            print("Command not recognized")

    doNotCollect200 = PackageEventThread(connective).start()

    # Start Main Loop
    print("Initializing main loop")
    while True:
        # TODO: Confirm System State According to Configuration
        for proc in processes:
            status = proc.GetStatus()
            if status is not None:
                recover.append(proc)
        # Monitor System Heartbeat
        print("Checking Heartbeats")
        for proc in processes:
            # TODO(any): add id field to Module, then use GetID here
            id = proc.GetName()
            result = connective.runs(
                "heartbeats '"+id+"' time-since", tolist=True)
            secondsSinceLastBeat = int(result[0])

            # Compare secondsSinceLastBeat to heartbeatThreshold
            if secondsSinceLastBeat > heartbeatThreshold:
                recover.append(proc)

        # Read Current Receipts
        # TODO: Prevent getting stuck in this code section
        print("Evaluating receipts in buffer")
        while True:
            if linebuffer:
                try:
                    receipt = json.loads(linebuffer.pop(0))
                    print("Source: {}\nEvent: {}".format(receipt["Source"], receipt["Event"]))
                    receipt_list.append(receipt)
                except json.decoder.JSONDecodeError:
                    pass
            elif not len(linebuffer):
                break

        # TODO: Validate Receipts
        print("Validating Receipts")
        while True:
            if receipt_list:
                item = receipt_list.pop(0)
                print(item)

                # TODO: possible key error ('Command' missing) would crash
                #       the manager
                receipt = connective.runs(
                    "get-receipt "+item['Command'], tolist=True)

                print("Receipt Popped: ", receipt) # Included for testing purposes only

                # TODO: if status is "started", push it to the back of the
                #       receipt list and track how long this command is taking

                # TODO: if status is neither "started" nor "complete", this
                #       may indicate a problem:
                #
                #       unrecognized:
                #       - Maybe HA/Connective just hasn't received it yet
                #              (likely if the receipt was just generated)
                #       - Maybe the component's connection to Connective has
                #         failed (likely if it's been a while)
                #
                #       inconsistent:
                #       - Internal error from HA/Connective; a failure should
                #         be reported. ("this should never happen" type error)
            elif not len(receipt_list):
                break
        # TODO: Expand methods for failure recovery based on application
        for index, proc in enumerate(recover):
            if proc.recovery == "standard":
                print("Restarting process "+proc.GetName())
                # TODO: Catch failed restart
                proc.Restart()
                # Remove subprocess from recovery list
                recover.pop(index)
            else:
                print("Recovery method not defined for "+proc.GetName())
                # TODO: Log failed recovery

        time.sleep(10) # Included for testing purposes only

if __name__ == "__main__":
    main()
