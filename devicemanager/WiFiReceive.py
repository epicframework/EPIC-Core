import socket, time, hashlib, random, sys
from time import sleep
from threading import Thread

HOST = "192.168.1.40"
PORT = 5454

server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
server.bind((HOST, PORT))
server.listen(1)

class SensorMonitor(Thread):

	def __init__(self):
		Thread.__init__(self)
		self.data = None
		self.last_data_hash = None
		conn, addr, type = self.socketConnect()
		self.conn = conn
		self.type = type
		self.addr = addr

	def socketConnect(self):
		conn, addr = server.accept()
		type = conn.recv(1024)
		type = type.strip().decode('utf-8')
		print(" ")
		print(type, "Connected")
		print(" ")
		return conn, addr, type

	def socketDisconnect(self):
		self.conn.close()
		print(self.type, "Disconnected")
		print(" ")

	def update(self):
		newData = self.conn.recv(1024)
		hash = hashlib.sha1(newData)
		if hash != self.last_data_hash:
			self.data = newData
			newData = newData.strip().decode('utf-8')
			return newData

	def run(self):
		while True:
			newData = self.update()
			print(newData)
			if bool(newData) == False:
				self.socketDisconnect()
				break

while True:
	device = SensorMonitor()
	try:
		device.start()
	except KeyboardInterrupt:
		server.close()
