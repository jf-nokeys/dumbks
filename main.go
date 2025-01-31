package main

import (
    "bufio"
    "flag"
    "fmt"
    "io"
    "log"
    "net"
    "os"
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

    l, err := net.Listen("tcp", "localhost:" + *port)
    if err != nil {
        fmt.Println(err)
        return
    }

    logger.Println("verbose logging")
    fmt.Fprintln(os.Stderr, "listening on port", *port)

    InitDb()


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
        resp := ParseCmd(reader)
        conn.Write(resp)
    }
}

