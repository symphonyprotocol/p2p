package models

import (
	"net"
	"github.com/symphonyprotocol/p2p/utils"
)

var (
	DEFAULT_TIMEOUT     = 3
	UDP_DIAGRAM_VERSION = 1
)

type NetworkDiagram struct {
	ID        string
	NodeID    string
	Timestamp int64
	DCategory string
	DType     string
	Version   int
}

func (d NetworkDiagram) GetID() string { return d.ID }
func (d NetworkDiagram) GetNodeID() string { return d.NodeID }
func (d NetworkDiagram) GetTimestamp() int64 { return d.Timestamp }
func (d NetworkDiagram) GetDCategory() string { return d.DCategory }
func (d NetworkDiagram) GetDType() string { return d.DType }
func (d NetworkDiagram) GetVersion() int { return d.Version }

type UDPDiagram struct {
	NetworkDiagram
	Expire    int64
	LocalAddr string
	LocalPort int
}

type CallbackParams struct {
	RemoteAddr net.Addr
	Diagram    IDiagram
	Data       []byte
}

type UDPCallbackParams struct {
	CallbackParams
}

func (c CallbackParams) GetRemoteAddr() net.Addr { return c.RemoteAddr }
func (c CallbackParams) GetDiagram() IDiagram { return c.Diagram }
func (c CallbackParams) GetData() []byte { return c.Data }

func (u UDPCallbackParams) GetUDPRemoteAddr() *net.UDPAddr {
	if addr, ok := u.RemoteAddr.(*net.UDPAddr); ok {
		return addr
	}

	return nil
}

func (u UDPCallbackParams) GetUDPDiagram() UDPDiagram {
	if diag, ok := u.Diagram.(UDPDiagram); ok {
		return diag
	}

	// error

	return UDPDiagram{}
}

type TCPDiagram struct {
	NetworkDiagram
}

func NewTCPDiagram() TCPDiagram {
	return TCPDiagram{
		NetworkDiagram{
			ID: utils.NewUUID(),
			DCategory: "default",
		},
	}
}
