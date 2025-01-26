from socket import socket, AF_INET, SOCK_STREAM
from struct import pack


class Conn:
    def __init__(self, port=7227):
        self.port = port

    def _send_bytes(self, b):
        with socket(AF_INET, SOCK_STREAM) as sock:
            sock.connect(("localhost", self.port))
            sock.sendall(b)
            raw = sock.recv(4096)

            if raw == b"\0":
                return None
            return raw.decode().strip()

    def set(self, key, val, ex=0):
        key = key.encode()
        val = val.encode()
        header = pack("<cBHI", b"s", len(key), len(val), ex)
        return self._send_bytes(header + key + val)

    def get(self, key):
        key = key.encode()
        return self._send_bytes(b"g" + key + b"\0")

    def remove(self, key):
        key = key.encode()
        return self._send_bytes(b"d" + key + b"\0")
