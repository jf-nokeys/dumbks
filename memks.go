package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type storeVal struct {
	value string
	expires time.Time
}

var db map[string]storeVal



func main() {
	db = make(map[string]storeVal)
	l, err := net.Listen("tcp", "localhost:7227")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("listening on port 7227")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConn(conn)
	}
}


func handleConn(conn net.Conn) {
	reader := bufio.NewReaderSize(conn, 4096)
	defer conn.Close()

	for {
		cmd, err := reader.ReadByte()
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			return
		}

		if cmd == 'g' || cmd == 'd' {
			key, err := reader.ReadString(0)
			if err != nil {
				if err != io.EOF {
					fmt.Println(err)
				}
				return
			}
			key = key[:len(key)-1]
			if cmd == 'g' {
				rec := db[key]

				if len(rec.value) > 0 && (rec.expires.IsZero() || rec.expires.After(time.Now())) {
					conn.Write([]byte(rec.value))
				} else {
					delete(db, key)
					conn.Write([]byte("\x00"))
				}
			} else {
				delete(db, key)
				conn.Write(append([]byte(key), []byte(" deleted")...))
			}
		} else if cmd == 's' {
			var keyLen uint8
			var valLen uint16
			var exSec uint32
			binary.Read(reader, binary.LittleEndian, &keyLen)
			binary.Read(reader, binary.LittleEndian, &valLen)
			binary.Read(reader, binary.LittleEndian, &exSec)
			var key []byte
			var val []byte
			key = make([]byte, keyLen)
			val = make([]byte, valLen)
			reader.Read(key)
			reader.Read(val)

			var ex time.Time
			if exSec > 0 {
				ex = time.Now().Add(time.Second * time.Duration(exSec))
			}

			db[string(key)] = storeVal{string(val), ex}	

			conn.Write(append(key, []byte(" added")...))
		} else {
			i := reader.Buffered()
			reader.Discard(i)
			fmt.Printf("unknown cmd %c, discarding %d bytes\n", cmd, i)
		}
	}
}