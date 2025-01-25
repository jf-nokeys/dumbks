from socket import socket, AF_INET, SOCK_STREAM
from struct import pack


def send_bytes(b):
	with socket(AF_INET, SOCK_STREAM) as s:
		s.connect(("localhost", 7227))
		s.sendall(b)
		raw = s.recv(4096)
		if raw == b"\0":
			return None
		return raw.decode().strip()


def set(key, val, ex=0):
	key = key.encode()
	val = val.encode()
	header = pack("<cBHI", b"s", len(key), len(val), ex)
	return send_bytes(header + key + val)


def get(key):
	key = key.encode()
	return send_bytes(b"g" + key + b"\0")


def remove(key):
	key = key.encode()
	return send_bytes(b"d" + key + b"\0")
