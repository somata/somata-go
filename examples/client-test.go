package main

import (
    . "github.com/spro/somata-go"
    "fmt"
    "time"
)

// Create a client that sends a ping message every second

func main() {
    c := NewClient()

    for req_n := 1; true; req_n++ {
        time.Sleep(1 * time.Millisecond)

        go func() {
            // Send a request
            resCh := c.Remote("test", "echo", fmt.Sprintf("testing #%d", req_n))
            // ...  and block for the response
            res := <-resCh
            fmt.Println(res)
        }()
    }
}

