package main

import (
    . "github.com/spro/somata-go"
    "fmt"
    "os"
    "os/signal"
    "time"
)

// Create a new service called "tester" with a few example methods
// ------------------------------------------------------------------------------

// Define the method map
var methods map[string]Method = map[string]Method {

    "echo": func(args []interface{}) interface{} {
        time.Sleep(100 * time.Millisecond)
        return args[0]
    },

    "ping": func(args []interface{}) interface{} {
        return "pong"
    },

    "one": func(args []interface{}) interface{} {
        return 1
    },

}

// Create the service and wait for a quit signal
func main() {
    testService := NewService("test", methods)

    start := time.Now()
    sigint := make(chan os.Signal, 1)
    signal.Notify(sigint, os.Interrupt)

    for {
        select {

        case <-testService.Quit:
            fmt.Println("Quit")
            return

        case <-sigint:
            fmt.Println("Quit")
            since := time.Since(start)
            fmt.Println(since)
            fmt.Println(testService.NResponses)
            fmt.Println(float64(testService.NResponses) / since.Seconds())
            return

        }
    }
}

