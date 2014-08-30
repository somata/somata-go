package somata

import (
    zmq "github.com/pebbe/zmq4"
    "fmt"
    "time"
)

var pollInterval = 100 * time.Millisecond

// Service definition
// ------------------------------------------------------------------------------

type Service struct {
    Name string
    Methods map[string]Method

    Binding *zmq.Socket

    Requests chan Request
    Responses chan Response
    NRequests int
    NResponses int

    Log chan interface{}
    Logging bool
    Quit chan bool
}

// Socket dispatch
// ------------------------------------------------------------------------------

// The dispatch loop keeps socket operations in one place, controlled by
// incoming and outgoing channels `Requests` and `Responses`, respectively.
// Each iteration the socket is first Recv'd on for new requests and sent to
// the Requests channel, then the Responses channel is pulled from to send
// outgoing messages over the socket.

func (s *Service) RunDispatch() {
    for {
        //fmt.Println("BindingDispatch ...")
        time.Sleep(pollInterval)
        s.RecvRequests()
        s.SendResponses()
        //fmt.Println("... BindingDispatch")
    }
}

// Recv tries getting new messages with a non-blocking socket read
// Receives a message, parsing and sending it to HandleRequest
func (s *Service) RecvRequests() {
    msg, err := s.Binding.RecvMessage(zmq.DONTWAIT)
    if err == nil {
        s.Requests <- parseRequest(msg)
        s.NRequests += 1
        // We call another Recv after a successful Recv in case there are
        // buffered messages waiting on the socket.
        s.RecvRequests()
    }
}

// Respond tries sending outgoing messages if they exist

func (s *Service) SendResponses() {
    for {
        ////fmt.Println("Checking BindingResponses")
        select {
        case res := <-s.Responses:
            s.NResponses += 1
            //fmt.Println("Read from BindingResponses")
            rs := res.toJson()
            //fmt.Println("Sending: ", rs)
            s.Binding.SendMessage(res.ClientId, rs)
        default:
            return
        }
    }
}

// Handling Requests
// ------------------------------------------------------------------------------

// The Handler loop waits for incoming messages on the Requests channel and
// does some work on them, sending results to the Responses channel.

// Listen on the Requests channel and send them to HandleRequest
func (s *Service) RunHandler() {
    for {
        select {
        case req := <-s.Requests:
            if s.Logging { s.Log <- req }
            go s.HandleRequest(req)
        }
    }
}

// Parse a Request for method and arguments, interpret and respond on Responses
func (s *Service) HandleRequest(req Request) {

    method := s.Methods[req.Method]

    //fmt.Println(req)

    if method != nil {
        r := Response {
            Id: req.Id,
            ClientId: req.ClientId,
            Data: method(req.Args),
        }
        s.Responses <- r
    } else {
        fmt.Printf("[ERROR] no such method %s\n", req.Method)
    }
}

// Creating a Service
// ------------------------------------------------------------------------------

// Create the binding
func (s *Service) CreateBinding() {
    s.Binding, _ = zmq.NewSocket(zmq.ROUTER)
    s.Binding.Bind("tcp://0.0.0.0:4444")
}

// Create a new service and start its listening methods
func NewService(name string, methods map[string]Method) *Service {
    s := &Service {
        Name: name,
        Methods: methods,
        Requests: make(chan Request),
        Responses: make(chan Response),
        Log: make(chan interface{}),
        Quit: make(chan bool),
    }
    s.CreateBinding()
    go s.RunDispatch()
    go s.RunHandler()
    return s
}

