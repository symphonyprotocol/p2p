package udp

import ()

var (
	DEFAULT_TIMEOUT     = 3
	UDP_DIAGRAM_VERSION = 1
	UDP_DIAGRAM_PING    = "PING"
	UDP_DIAGRAM_PONG    = "PONG"
)

type UDPDiagram struct {
	ID        string
	Timestamp int64
	DType     string
	Version   int
	Expire    int64
}

type PingDiagram struct {
	UDPDiagram
	NodeID    string
	LocalAddr string
	LocalPort int
}

type PongDiagram struct {
	UDPDiagram
	NodeID     string
	LocalAddr  string
	LocalPort  int
	RemoteAddr string
	RemotePort int
}
