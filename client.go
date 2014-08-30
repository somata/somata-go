package somata

import (
    zmq "github.com/pebbe/zmq4"
    "fmt"
    "time"
)

// Client and connection definitions
// ------------------------------------------------------------------------------

// A Client has one Connection per Service; each Connection has a socket and
// channels for requests and responses

type Client struct {
    Connections map[string]Connection
    Requests chan Request
    PendingResponses map[string]chan Response
}

// Creating a client
// ------------------------------------------------------------------------------

func NewClient() *Client {
    c := &Client {
        Connections: make(map[string]Connection),
        Requests: make(chan Request),
        PendingResponses: make(map[string]chan Response),
    }
    go c.RunDispatch()
    return c
}

// Sending Requests
// ------------------------------------------------------------------------------

func (c *Client) Remote(service string, method string, args ...interface{}) chan Response {
    req := Request {
        Id: RandomString(),
        Service: service,
        Method: method,
        Args: args,
    }
    return c.SendRequest(service, req)
}

func (c *Client) SendRequest(service string, req Request) chan Response {
    c.Requests <- req
    //fmt.Println("Sending request", req)
    onResponse := make(chan Response)
    c.PendingResponses[req.Id] = onResponse
    return onResponse
}

func (c *Client) HandleResponse(res Response) {
    // Send response to the pending channel
    pendingResponse := c.PendingResponses[res.Id]
    if pendingResponse != nil {
        pendingResponse <- res
    }
}

// Socket Dispatch
// ------------------------------------------------------------------------------

// ?? Each of the Client's Connections has a Socket that must send and receive on
// a single thread. The Client thus has a Requests channel which pulls Requests 
// through to send with the Socket and a pendingResponses map of channels to send
// read Responses through, all managed from this single socket dispatch goroutine.

func (c *Client) RunDispatch() {
    for {
        time.Sleep(pollInterval)
        c.SendRequests()
        c.RecvResponses()
    }
}

func (c *Client) SendRequests() {
    for {
        select {
        case req := <-c.Requests:
            // Send outgoing Requests
            conn, exists := c.Connections[req.Service]
            if !exists {
                conn = c.CreateConnection(req.Service)
            }
            conn.Socket.SendMessage(req.toJson())
        default:
            return
        }
    }
}

func (c *Client) RecvResponses() {

    // Iterate through connections to read from sockets
    for service, conn := range c.Connections {
        conn.RecvResponses(service, c)
    }
}

func (conn *Connection) RecvResponses(s string, c *Client) {
    // Read socket
    msg, err := conn.Socket.RecvMessage(zmq.DONTWAIT)
    if err == nil {

        // Got a Response
        res := parseResponse(msg)
        res.Service = s

        go c.HandleResponse(res)
        conn.RecvResponses(s, c)
    }
}

// Creating a connection
// ------------------------------------------------------------------------------

type Connection struct {
    Service string
    Socket *zmq.Socket
}

func NewConnection() Connection {
    sock, _ := zmq.NewSocket(zmq.DEALER)
    sock.SetIdentity(RandomString())
    conn := Connection {
        Socket: sock,
    }
    return conn
}

func (c *Client) CreateConnection(id string) Connection {
    conn := NewConnection()
    conn.Socket.Connect("tcp://localhost:4444")
    c.Connections[id] = conn
    fmt.Printf("Connected to %s\n", id)
    return conn
}

