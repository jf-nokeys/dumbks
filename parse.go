package main 

import (
    "bufio"
    "io"
)

func ReadKey(reader *bufio.Reader) string {
    key, err := reader.ReadString(0)
    if err != nil {
        if err != io.EOF {
            logger.Println(err)
        }
        return ""
    }
    return key[:len(key)-1]
}

func ParseCmd(reader *bufio.Reader) []byte {
    cmd, err := reader.ReadByte()
    if err != nil {
        if err != io.EOF {
            logger.Println(err)
        }
        return []byte("\x00")
    }

    switch cmd {
    case 'g':
        return GetKey(reader)
    case 'd':
        return DelKey(reader)
    case 's':
        return SetKey(reader)
    case 't':
        return TtlKey(reader)
    case 'p':
        logger.Println("p pong")
        return []byte("pong")
    default:
        i := reader.Buffered()
        reader.Discard(i)
        logger.Printf("unknown cmd %c, discarding %d bytes\n", cmd, i)
        return []byte("\x00")
    }
}
