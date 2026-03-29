# ImgBB

ImgBB is an [imgbb.com](https://imgbb.com) api client.

Installation

```bash
go get github.com/tomek7667/imgbb
```

Example of usage:

```go
package main

import (
    "context"
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "time"

    imgBB "github.com/tomek7667/imgbb"
)

const (
    key = "your-imgBB-api-key"
)

func main() {
    f, err := os.Open("example.jpg")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    b, err := io.ReadAll(f)
    if err != nil {
        log.Fatal(err)
    }

    img, err := imgBB.NewImageFromFile(hashSum(b), b)
    if err != nil {
        log.Fatal(err)
    }

    imgBBClient := imgBB.NewClient(key)

    resp, err := imgBBClient.Upload(context.Background(), img)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%v\n", resp)
}

func hashSum(b []byte) string {
    sum := md5.Sum(b)

    return hex.EncodeToString(sum[:])
}
```
