# P2P
Decentralized P2P networking framework

## How to run example
### 1. Create a configuration file `~/.symchaincfg` in json format with the static nodes:
```json
{
    "nodes":
    [
        {
            "id": "c4ef0694fee0cdf78eab30c83b325293047e0b27511b92e8e206b199b24f13ea",
            "ip":"101.200.156.243",
            "port": 32768
        }
    ]
}
```

> You can find the `id` in the console log when launching your application for now...

### 2. Start go file `examples/main.go` and it will connect to the static node and discover other nodes.

## How to use it in your application
```go
import (
    "github.com/symphonyprotocol/p2p"
    "github.com/symphonyprotocol/p2p/tcp"
    "fmt"
)

type SampleMiddleware struct {}

func (s *SimpleMiddleware) Handle(ctx *tcp.P2PContext) {
    fmt.Println("Handling message")
}
func (s *SimpleMiddleware) Start(ctx *tcp.P2PContext) {
    fmt.Println("Middleware started")
}
func (s *SimpleMiddleware) AcceptConnection(conn *tcp.TCPConnection) {
    fmt.Println("New connection got")
}
func (s *SimpleMiddleware) DropConnection(conn *tcp.TCPConnection) {
    fmt.Println("Connection dropped")
}

func (s *SimpleMiddleware) Name() string { return "Sample Middleware" }
func (s *SimpleMiddleware) DashboardData() interface{} { return [][]string{ } } 
func (s *SimpleMiddleware) DashboardType() string { return "table" }
func (s *SimpleMiddleware) DashboardTitle() string { return "Sample" }
func (s *SimpleMiddleware) DashboardTableHasColumnTitles() bool { return false }

func main() {
    server := p2p.NewP2PServer()
    server.Use(&SimpleMiddleware{})
}
```
