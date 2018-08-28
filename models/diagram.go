package models

import (
	"net"
)

var (
	DEFAULT_TIMEOUT     = 3
	UDP_DIAGRAM_VERSION = 1
)

type UDPDiagram struct {
	ID        string
	NodeID    string
	Timestamp int64
	DCategory string
	DType     string
	Version   int
	Expire    int64
	LocalAddr string
	LocalPort int
}

type CallbackParams struct {
	RemoteAddr *net.UDPAddr
	Diagram    UDPDiagram
	Data       []byte
}
