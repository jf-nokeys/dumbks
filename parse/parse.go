package parse 

import (
    "bufio"
    "io"
    "log"

    "github.com/jf-nokeys/dumbks/db"
)

func discardBuffer(reader *bufio.Reader, logger *log.Logger) {
    i := reader.Buffered()
    reader.Discard(i)
    if i > 0 {
       logger.Printf("discarding %d bytes\n", i)
    }
}

func ParseCmd(reader *bufio.Reader, ks *db.DumbKS, logger *log.Logger) []byte {
    cmd, err := reader.ReadByte()
    if err != nil {
        if err != io.EOF {
            logger.Println(err)
        }
        return []byte("\x00")
    }

    switch cmd {
    case 'g':
        return ks.GetKey(reader, logger)
    case 'd':
        return ks.DelKey(reader, logger)
    case 's':
        return ks.SetKey(reader, logger)
    case 't':
        return ks.TtlKey(reader, logger)
    case 'p':
        logger.Println("p pong")
        discardBuffer(reader, logger)
        return []byte("pong")
    default:
        logger.Printf("unknown cmd %c\n", cmd)
        discardBuffer(reader, logger)
        return []byte("\x00")
    }
}
