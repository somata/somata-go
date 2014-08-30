package somata

import (
    "encoding/json"
    "math/rand"
    "time"
    "fmt"
)

// Helpers

func init() {
    rand.Seed(time.Now().UnixNano())
}
func RandomString() string {
    return fmt.Sprintf("%04x%04x%04x", rand.Intn(0x10000), rand.Intn(0x10000), rand.Intn(0x10000))
}

// Methods
// ------------------------------------------------------------------------------

// A method takes an array of arguments (values) and returns a value
type Method func(args []interface{}) interface{}

// Requests
// ------------------------------------------------------------------------------

// A client sends a Request message with attributes in JSON
type Request struct {
    ClientId string
    Service string
    Id string
    Method string
    Args []interface{}
}

// Parse a packet with a ClientId and JSON string into a Request
func parseRequest(msg []string) Request {
    rJson := map[string]interface{}{}
    json.Unmarshal([]byte(msg[1]), &rJson)
    req := Request {
        ClientId: msg[0],
        Id: rJson["id"].(string),
        Method: rJson["method"].(string),
        Args: rJson["args"].([]interface{}),
    }
    return req
}

// Turn a Request into a JSON string
func (r Request) toJson() string {
    rJson := map[string]interface{} {
        "kind": "method",
        "id": r.Id,
        "method": r.Method,
        "args": r.Args,
    }
    rString, _ := json.Marshal(rJson)
    return string(rString)
}
func (r Request) String() string {
    return fmt.Sprintf("[%s] --> <%s> %s(%v)", r.ClientId, r.Id, r.Method, r.Args)
}

// Responses
// ------------------------------------------------------------------------------

// Services respond to Clients with Response messages with a JSON data payload
type Response struct {
    ClientId string
    Service string
    Id string
    Data interface{}
}

// Parse a packet with a ClientId and JSON string into a Response
func parseResponse(msg []string) Response {
    rJson := map[string]interface{}{}
    json.Unmarshal([]byte(msg[0]), &rJson)
    req := Response {
        Id: rJson["id"].(string),
        Data: rJson["response"],
    }
    return req
}

// Turn a Response into a JSON string
func (r Response) toJson() string {
    rJson := map[string]interface{} {
        "kind": "response",
        "id": r.Id,
        "response": r.Data,
    }
    rString, _ := json.Marshal(rJson)
    return string(rString)
}
func (r Response) String() string {
    return fmt.Sprintf("[%s] ==> <%s> %v", r.Service, r.Id, r.Data)
}

