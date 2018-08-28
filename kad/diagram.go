package kad

import (
	"github.com/symphonyprotocol/p2p/models"
)

var (
	KTABLE_DIAGRAM_CATEGORY     = "KTABKE"
	KTABLE_DIAGRAM_PING         = "PING"
	KTABLE_DIAGRAM_PONG         = "PONG"
	KTABLE_DIAGRAM_FINDNODE     = "FINDNODE"
	KTABLE_DIAGRAM_FINDNODERESP = "FINDNODERESP"
)

type PingDiagram struct {
	models.UDPDiagram
}

type PongDiagram struct {
	models.UDPDiagram
	RemoteAddr string
	RemotePort int
}

type FindNodeDiagram struct {
	models.UDPDiagram
}

type FindNodeRespDiagram struct {
	models.UDPDiagram
	Nodes []NodeDiagram
}

type NodeDiagram struct {
	NodeID    string
	IP        string
	Port      int
	LocalAddr string
	LocalPort int
}
