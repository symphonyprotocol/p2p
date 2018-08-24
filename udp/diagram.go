package udp

var (
	DEFAULT_TIMEOUT     = 3
	UDP_DIAGRAM_VERSION = 1
	UDP_DIAGRAM_PING    = "PING"
	UDP_DIAGRAM_PONG    = "PONG"
)

type UDPDiagram struct {
	ID        string
	NodeID    string
	Timestamp int64
	DType     string
	Version   int
	Expire    int64
}

type PingDiagram struct {
	UDPDiagram
	LocalAddr string
	LocalPort int
}

type PongDiagram struct {
	UDPDiagram
	LocalAddr  string
	LocalPort  int
	RemoteAddr string
	RemotePort int
}
