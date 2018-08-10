package main

import (
    "encoding/base64"
    "fmt"
    "io/ioutil"
    "math/rand"
    "net"
    "os"
    "strconv"
    "time"
)

func main() {
    if len(os.Args) < 3 {
        help()
        os.Exit(1)
    }

    filename := os.Args[1]
    dst := os.Args[2]
    count := 1
    if len(os.Args) > 3 {
        var err error
        count, err = strconv.Atoi(os.Args[3])
        if err != nil {
            exitByErr(err)
        }
    }

    cnt, err := ioutil.ReadFile(filename)
    if err != nil {
        exitByErr(err)
    }
    conn, err := net.Dial("tcp", dst)
    if err != nil {
        exitByErr(err)
    }

    defer conn.Close()

    for index, b := range cnt {
        cnt[index] = b ^ 0xff
    }

    buf := []byte(base64.StdEncoding.EncodeToString(cnt))

    fmt.Printf("total len: %d\n", len(buf)*count)

    rand.Seed(time.Now().Unix())

    for i := 0; i < count; i++ {
        l := len(buf)
        avail := l
        offset := 0
        used := l

        for avail > 0 {
            if avail <= 1 {
                used = avail
            } else {
                used = rand.Intn(avail-1) + 1
            }
            wb := buf[offset : offset+used]
            offset += used
            avail -= used

            n, err := conn.Write(wb)
            if err != nil {
                exitByErr(err)
            }
            fmt.Printf("write %d bytes to %s\n", n, dst)

            if avail > 0 {
                time.Sleep(time.Second / 100)
            }
        }

    }
}

func help() {
    fmt.Println("usage: put filename host:port [count]")
}

func exitByErr(err error) {
    fmt.Printf("ERR: %s\n\n", err)
    help()
    os.Exit(1)
}
