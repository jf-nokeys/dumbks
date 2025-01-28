package main

import (
	"bufio"
	"encoding/binary"
    "flag"
	"fmt"
	"io"
    "log"
	"net"
    "os"
    "sync"
	"time"
)

type storeVal struct {
	value   []byte
	expires time.Time
}

var db map[string]storeVal
var lock sync.RWMutex

func main() {
    port := flag.String("p", "7227", "server port")
    verbose := flag.Bool("v", false, "log to stdout")
    flag.Parse()

    logger := log.New(io.Discard, "", 0)

    if *verbose == true {
        logger.SetOutput(os.Stdout)
    }

	db = make(map[string]storeVal)
	l, err := net.Listen("tcp", "localhost:" + *port)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Fprintln(os.Stderr, "listening on port", *port)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConn(conn, logger)
	}
}

func handleConn(conn net.Conn, logger *log.Logger) {
	reader := bufio.NewReaderSize(conn, 4096)
	defer conn.Close()

	for {
		cmd, err := reader.ReadByte()
		if err != nil {
			if err != io.EOF {
				logger.Println(err)
			}
			return
		}

		if cmd == 'g' || cmd == 'd' {
            lock.RLock()
			key, err := reader.ReadString(0)
			if err != nil {
				if err != io.EOF {
					logger.Println(err)
				}
				return
			}
			key = key[:len(key)-1]
			if cmd == 'g' {
				rec := db[key]

				if len(rec.value) > 0 && (rec.expires.IsZero() || rec.expires.After(time.Now())) {
                    logger.Printf("g [%s] %s\n", key, rec.value)
					conn.Write(rec.value)
				} else {
					delete(db, key)
                    logger.Printf("g [%s] <NULL>\n", key)
					conn.Write([]byte("\x00"))
				}
			} else {
				delete(db, key)
                logger.Printf("d [%s]\n", key)
				conn.Write(append([]byte(key), []byte(" deleted")...))
			}
            lock.RUnlock()
		} else if cmd == 's' {
            lock.Lock()
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

			db[string(key)] = storeVal{val, ex}
            lock.Unlock()
            logger.Printf("s [%s] %s %d sec\n", string(key), val, exSec)

			conn.Write(append(key, []byte(" added")...))
		} else {
			i := reader.Buffered()
			reader.Discard(i)
			logger.Printf("unknown cmd %c, discarding %d bytes\n", cmd, i)
			conn.Write([]byte("\x00"))
		}
	}
}
