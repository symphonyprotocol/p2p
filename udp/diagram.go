package udp

import (
	"encoding/json"
	"log"
)

var (
	UDP_DIAGRAM_PING = "PING"
	UDP_DIAGRAM_PONG = "PONG"
)

type UDPDiagram struct {
	ID             string
	Timestamp      int64
	DType          string
	Version        int
	ExpireDuration int
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

func BytesToUDPDiagram(data []byte) UDPDiagram {
	diagram := UDPDiagram{}
	err := json.Unmarshal(data, &diagram)
	if err != nil {
		log.Fatal(err)
	}
	return diagram
}
