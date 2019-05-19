# -*- coding: utf-8 -*-
import subprocess
import os
from threading import Thread

BASE_PATH = "./../server/packages"

def reader(f,buffer):
    while True:
        line = f.readline()
        if line:
            buffer.append(line)
        else:
            break

class Error(Exception):
    """Base Class for Custom Exceptions"""
    pass

class CommandNotRecognized(Error):
    """Raised when command is invalid"""
    pass

class Module:

    def __init__(self, name, cmd, recovery, linebuffer, package=False, package_name=""):
        # TODO: CLI arg variable replacement for packages
        #       (right now only happens for internal components)
        self.name = name
        self.linebuffer = linebuffer
        self.cmd = cmd
        self.package = package
        self.package_name = package_name
        self.recovery = recovery

        if not package:
            if cmd[0] == 'exec':
                p = subprocess.Popen(cmd[1], stdout=subprocess.PIPE, universal_newlines=True)
                self.process = p
            else:
                raise CommandNotRecognized
        else:
            if cmd[0] == 'docker':
                image_path = os.path.join(BASE_PATH, package_name, cmd[1])

                # Docker load
                load_command = ['./loader', image_path]
                docker_load = subprocess.Popen(load_command, stdout=subprocess.PIPE, universal_newlines=True)
                returncode = docker_load.wait()
                if returncode == 0:

                    # Docker run
                    run_command = ['docker', 'run', cmd[2]]
                    # === Currently Unsupported === #
                    # # Add CLI arguments from package manifest
                    # for arg in cmd[3]:
                    #     run_command.append(arg)
                    p = subprocess.Popen(run_command, stdout=subprocess.PIPE, universal_newlines=True)
                    self.process = p
                else:
                    raise Exception('docker load failed')
            else:
                raise CommandNotRecongized

        p = self.process
        t = Thread(target=reader,args=(p.stdout,linebuffer))
        t.daemon=True
        t.start()
        self.thread = t

    # TODO: Add logging for restart
    def Restart(self):
        cmd = self.cmd
        package = self.package
        package_name = self.package_name

        if not package:
            if cmd[0] == 'exec':
                p = subprocess.Popen(cmd[1], stdout=subprocess.PIPE, universal_newlines=True)
                self.process = p
            else:
                raise CommandNotRecognized
        else:
            if cmd[0] == 'docker':
                image_path = os.path.join(BASE_PATH, package_name, cmd[1])
                #load_command = ['docker', 'load', image_path]
                #docker_load = subprocess.Popen(load_command, stdout=subprocess.PIPE, universal_newlines=True)
                #returncode = p.wait()
                if True: # not returncode == 0:
                    print("!!! Restart not implemented for Docker yet")
            else:
                raise CommandNotRecongized

        t = Thread(target=reader,args=(p.stdout,self.linebuffer))
        t.daemon=True
        t.start()
        self.thread = t

    # TODO: Add logging for reinitialization
    def ReInit(self, cmd):
        self.cmd = cmd
        package = self.package
        package_name = self.package_name

        if not package:
            if cmd[0] == 'exec':
                p = subprocess.Popen(cmd[1], stdout=subprocess.PIPE, universal_newlines=True)
                self.process = p
            else:
                raise CommandNotRecognized
        else:
            if cmd[0] == 'docker':
                image_path = os.path.join(BASE_PATH, package_name, cmd[1])
                #load_command = ['./loader', image_path]
                #docker_load = subprocess.Popen(load_command, stdout=subprocess.PIPE, universal_newlines=True)
                #returncode = docker_load.wait()
                if True: # not returncode == 0:
                    print("!!! Reinit not implemented for Docker yet")
            else:
                raise CommandNotRecongized

        t = Thread(target=reader,args=(p.stdout,self.linebuffer))
        t.daemon=True
        t.start()
        self.thread = t

    def GetStatus(self):
        status = self.process.poll()
        if status is None:
            print("Process {} is active".format(self.name))
        else:
            print("Process for {} has exited with return code {}".format(self.name, status))
        return status

    def GetName(self):
        return self.name

    def GetCommand(self):
        return self.command

    def GetProcess(self):
        return self.process

    def SetProcess(self, process):
        self.process = process

    def GetThread(self):
        return self.thread

    def SetThread(self, thread):
        self.thread = thread

    def GetRecovery(self):
        return self.recovery

    def SetRecovery(self, recovery)
        self.recovery = recovery
