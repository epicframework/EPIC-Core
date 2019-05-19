import socket, json, sys, datetime, logging
from threading import Thread

# Import HA/Connective bindings
from bindings import new_ll, connect, new_interpreter

# Setup Logging
logging.basicConfig(filename='devicemanager.log', filemode='w', format='%(name)s - %(levelname)s - %(message)s')

#Listen for broadcasts when found return name of device found
#Maintain map of address for devices (name tp ip associated) - dict that
#When connective requests, connect to device from name connective gives associa$
#Once connected, new dict made of device name to socket-thread (FAB) established
#Send JSON given from connective to device name provided referenced in socket d$

deviceAddressDict = {}
connectedDeviceThreads = {}

def listenForBroadcast(broadcastClient):
        global deviceAddressDict
        data, addr = broadcastClient.recvfrom(1024)
        jsonData = json.loads(str(data, "utf-8"))
        for key in jsonData:
                ip = jsonData[key]
                if jsonData[key] not in deviceAddressDict:
                        deviceAddressDict[ip] = key
                        print("SAVED", deviceAddressDict)
                        return key, ip, jsonData

def createReceipt(source, event, command, result):
        timestamp = datetime.datetime.now()
        receipt = json.dumps({
                "Source": source,
                "Event": event,
                "Command": command,
                "Time Stamp": timestamp,
                "Result": result
        })
        logging.info(receipt)
        sys.stdout.write(receipt+"\n")
        sys.stdout.flush()

class SendCommandThread(Thread):

        def __init__(self, device, command, dataType, listener):
                Thread.__init__(self)
                self.device = device
                self.command = command
                self.dataType = dataType
                self.listener = listener
                self.conn = self.connectToDevice(device)

        def run(self):
                global connectedDeviceThreads
                device = self.device
                command = self.command
                client = connectedDeviceThreads[device]
                while True:
                        try:
                                client.send(str.encode(command))
                                response = client.recv(1024)
                                stringResponse = str(response, "utf-8")
                                jsonResponse = None
                                try:
                                    jsonResponse = json.loads(stringResponse)
                                except Exception as exc:
                                    logging.error("blank reading O.o")
                                    logging.error(exc)
                                if True:
                                    self.listener.update_device_single_value(self.device,
                                        jsonResponse["output"])
                                    logging.info("\r\n %s: %s" % (jsonResponse["src"], jsonResponse["output"]))
                        except Exception as exc:
                                logging.error("\r\n %s disconnected" % self.device)
                                logging.error(exc)
                                self.disconnectFromDevice()
                                exit()

        def connectToDevice(self, haConnectiveRequest):
                global connectedDeviceThreads
                client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                HOST = haConnectiveRequest
                PORT = 5454
                connectedDeviceThreads[haConnectiveRequest] = client
                try:
                        client.connect((HOST, PORT))
                        bytesSent = client.send(str.encode(json.dumps({"src": "hub", "code": "10", "msg": "request specification"})))
                        mozdef = json.loads(str(client.recv(1024), 'utf-8'))
                        # Figure out what do
                        whatDo = self.dataType
                        if whatDo == 'read hum':
                                # Bind??? device to Connective
                                self.listener.add_device(self.device, mozdef, "humidity")
                        elif whatDo == 'read temp':
                                # Bind??? device to Connective
                                self.listener.add_device(self.device, mozdef, "temperature")
                        elif whatDo == 'read occ':
                                # Bind??? device to Connective
                                self.listener.add_device(self.device, mozdef, "occupancy")
                except Exception as exc:
                        client.close()
                        logging.error(exc)
                        exit()
                logging.info("Connected: %s" % (connectedDeviceThreads))
                return client

        def disconnectFromDevice(self):
                # Removing from dictionary
                del deviceAddressDict[self.device]
                del connectedDeviceThreads[self.device]
                self.listener.del_device(self.device)
                # Close connection
                self.conn.close()

class DeviceEventListener:
    def __init__(self, connective):
        self.source = "broadcast_client"
        self.connective = connective
        self.deviceIPs = {}
        self.ipToKeyName = {}
    def add_device(self, deviceIP, deviceDef, keyName):
        # Add device via connective
        result = self.connective.runl(
                'hub devices add-device'.split(' ') +
                [deviceDef, {}, self.source + '.' + deviceIP],
                tolist=True)
        deviceUUID = result[0]
        self.deviceIPs[deviceIP] = deviceUUID
        self.ipToKeyName[deviceIP] = keyName
    def update_device_single_value(self, deviceIP, value):
        keyName = self.ipToKeyName[deviceIP]
        deviceUUID = self.deviceIPs[deviceIP]
        self.connective.runl(
                'hub devices internal_registry'.split(' ') +
                [deviceUUID, 'update-queue', 'enque',
                    {keyName: value}], tolist=True)
    def del_device(self, deviceIP):
        # Remove device via connective
        deviceUUID = self.deviceIPs[deviceIP]
        del self.ipToKeyName[deviceIP]
        del self.deviceIPs[deviceIP]
        # TODO: Call connective to remove device


def launch_client(deviceEventListener):
        broadcastClient = socket.socket(socket.AF_INET, socket.SOCK_DGRAM) # UDP
        broadcastClient.setsockopt(socket.SOL_SOCKET, socket.SO_BROADCAST, 1)
        broadcastClient.bind(("", 37020))

        check = None

        while True:

                deviceName, device, broadcastJson = listenForBroadcast(broadcastClient)

                # TODO: make connectToDevice handle JSON instead of string
                if deviceName == "Humidity Sensor":
                        command = "read hum"
                elif deviceName == "Temperature Sensor":
                        command = "read temp"
                elif deviceName == "Occupancy Sensor":
                        command = "read occ"
                else:
                        logging.info(deviceName)
                        logging.info("Not a compatible device for this framework")
                        exit()
                dataType = command
                command = json.dumps({"src": "Hub", "code": "20", "msg": command})
                commandThread = SendCommandThread(device, command, dataType,
                        deviceEventListener)
                commandThread.start()

def main():
        # CLI arguments
        arg_id = sys.argv[1]
        arg_re = sys.argv[2]

        # Create local interpreter and remote interpreter
        ll = new_ll("../connective/connective/sharedlib/elconn.so")
        ll.elconn_init(1)
        remote_connective = connect(ll, arg_re.encode())
        connective = new_interpreter(ll)

        # Allow messages to be send to remote interpreter by prefixing
        # the command "hub"
        ll.elconn_link(b"hub", connective.ii, remote_connective.ii)

        deviceEventListener = DeviceEventListener(connective)
        launch_client(deviceEventListener)

if __name__ == '__main__':
        main()
