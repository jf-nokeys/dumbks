package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/jf-nokeys/dumbks/db"
	"github.com/jf-nokeys/dumbks/parse"
)

var logger *log.Logger

func main() {
	port := flag.String("p", "7227", "server port")
	verbose := flag.Bool("v", false, "log to stdout")
	flag.Parse()

	logger = log.New(io.Discard, "", 0)

	if *verbose == true {
		logger.SetOutput(os.Stdout)
	}

	l, err := net.Listen("tcp", "localhost:"+*port)
	if err != nil {
		fmt.Println(err)
		return
	}

	logger.Println("verbose logging")
	fmt.Fprintln(os.Stderr, "listening on port", *port)

	ks := db.InitDb()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConn(conn, ks, logger)
	}
}

func handleConn(conn net.Conn, ks *db.DumbKS, logger *log.Logger) {
	reader := bufio.NewReaderSize(conn, 4096)
	defer conn.Close()

	resp := parse.ParseCmd(reader, ks, logger)
	conn.Write(resp)
}
