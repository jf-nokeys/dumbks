package db

import (
    "bufio"
    "encoding/binary"
    "io"
    "log"
    "sync"
    "time"
)


type storeVal struct {
	value   []byte
	expires time.Time
}

type DumbKS struct {
    db map[string]storeVal
    sync.RWMutex
}

func (ks *DumbKS) getValue(key string) storeVal {
    ks.RLock()
    defer ks.RUnlock()
    return ks.db[key]
}

func (ks *DumbKS) delValue(key string) {
    ks.Lock()
    defer ks.Unlock()
    delete(ks.db, key)
}

func (ks *DumbKS) setValue(key string, val []byte, ex time.Time) {
    ks.Lock()
    defer ks.Unlock()
    ks.db[key] = storeVal{val, ex}
}

func InitDb() *DumbKS {
    return &DumbKS{
        db: make(map[string]storeVal),
    }
}

func readKey(reader *bufio.Reader, logger *log.Logger) string {
    key, err := reader.ReadString(0)
    if err != nil {
        if err != io.EOF {
            logger.Println(err)
        }
        return ""
    }
    return key[:len(key)-1]
}


func (ks *DumbKS) GetKey(reader *bufio.Reader, logger *log.Logger) []byte {
    key := readKey(reader, logger)
    rec := ks.getValue(key)

    if len(rec.value) > 0 && (rec.expires.IsZero() || rec.expires.After(time.Now())) {
        logger.Printf("g [%s] %s\n", key, rec.value)
        return rec.value
    }

    ks.delValue(key)
    logger.Printf("g [%s] <NULL>\n", key)
    return []byte("\x00")
}

func (ks *DumbKS) DelKey(reader *bufio.Reader, logger *log.Logger) []byte {
    key := readKey(reader, logger)
    ks.delValue(key)
    logger.Printf("d [%s] <NULL>\n", key)
    return []byte("\x00")
}

type setArgs struct {
    keyLen uint8
    valLen uint16
    exSec uint32
}

func (ks *DumbKS) SetKey(reader *bufio.Reader, logger *log.Logger) []byte {
    var args setArgs
    binary.Read(reader, binary.LittleEndian, &args.keyLen)
    binary.Read(reader, binary.LittleEndian, &args.valLen)
    binary.Read(reader, binary.LittleEndian, &args.exSec)
    key := make([]byte, args.keyLen)
    val := make([]byte, args.valLen)
    reader.Read(key)
    reader.Read(val)

    var ex time.Time
    if args.exSec > 0 {
        ex = time.Now().Add(time.Second * time.Duration(args.exSec))
    }
    ks.setValue(string(key), val, ex)
    logger.Printf("s [%s] %s %d sec\n", string(key), val, args.exSec)
    return append(key, []byte(" added")...)
}

func (ks *DumbKS) TtlKey(reader *bufio.Reader, logger *log.Logger) []byte {
    key := readKey(reader, logger)
    rec := ks.getValue(key)

    if len(rec.value) > 0 {
        if rec.expires.IsZero() {
            return []byte("-1")
        }
        if rec.expires.After(time.Now()) {
            var ttl uint32
            var buf []byte

            dur := time.Until(rec.expires)
            ttl = uint32(dur.Seconds())
            buf = make([]byte, 4)
            _, _ = binary.Encode(buf, binary.LittleEndian, &ttl)
            logger.Printf("-- debug %v\n", dur)

            logger.Printf("t [%s] %d\n", key, ttl)
            return buf
        }
    }
    return []byte("\x00")
}


