package main

import (
    "bufio"
    "encoding/binary"
    "sync"
    "time"
)


type storeVal struct {
	value   []byte
	expires time.Time
}

var db map[string]storeVal
var lock sync.RWMutex

func getValue(key string) storeVal {
    lock.RLock()
    defer lock.RUnlock()
    return db[key]
}

func delValue(key string) {
    lock.Lock()
    defer lock.Unlock()
    delete(db, key)
}

func setValue(key string, val []byte, ex time.Time) {
    lock.Lock()
    defer lock.Unlock()
    db[key] = storeVal{val, ex}
}

func InitDb() {
    db = make(map[string]storeVal)
}

func GetKey(reader *bufio.Reader) []byte {
    key := ReadKey(reader)
    rec := getValue(key)

    if len(rec.value) > 0 && (rec.expires.IsZero() || rec.expires.After(time.Now())) {
        logger.Printf("g [%s] %s\n", key, rec.value)
        return rec.value
    }

    delValue(key)
    logger.Printf("g [%s] <NULL>\n", key)
    return []byte("\x00")
}

func DelKey(reader *bufio.Reader) []byte {
    key := ReadKey(reader)
    delValue(key)
    logger.Printf("d [%s] <NULL>\n", key)
    return []byte("\x00")
}

type setArgs struct {
    keyLen uint8
    valLen uint16
    exSec uint32
}

func SetKey(reader *bufio.Reader) []byte {
    var args setArgs
    binary.Read(reader, binary.LittleEndian, &args.keyLen)
    binary.Read(reader, binary.LittleEndian, &args.valLen)
    binary.Read(reader, binary.LittleEndian, &args.exSec)
    var key []byte
    var val []byte
    key = make([]byte, args.keyLen)
    val = make([]byte, args.valLen)
    reader.Read(key)
    reader.Read(val)

    var ex time.Time
    if args.exSec > 0 {
        ex = time.Now().Add(time.Second * time.Duration(args.exSec))
    }
    setValue(string(key), val, ex)
    logger.Printf("s [%s] %s %d sec\n", string(key), val, args.exSec)
    return append(key, []byte(" added")...)
}

func TtlKey(reader *bufio.Reader) []byte {
    key := ReadKey(reader)
    rec := getValue(key)

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


