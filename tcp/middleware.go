package tcp

import (
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/models"
)

var p2pLogger = log.GetLogger("P2PService")

type P2PContext struct {
	_skipped    bool
	_network    models.INetwork
	_localNode  *node.LocalNode
	_connection *TCPConnection
}

func NewP2PContext(network models.INetwork, localNode *node.LocalNode, conn *TCPConnection) *P2PContext {
	return &P2PContext{
		_skipped:    false,
		_network:    network,
		_localNode:  localNode,
		_connection: conn,
	}
}

func (ctx *P2PContext) Send(bytes []byte) {
	length, err := ctx._connection.Write(bytes)
	if err != nil {
		p2pLogger.Error("Failed to send packet (%d) to %v", length, ctx._connection.RemoteAddr().String())
	} else {
		p2pLogger.Trace("Packet (%d) sent to %v", length, ctx._connection.RemoteAddr().String())
	}
}

func (ctx *P2PContext) Next() {
	ctx._skipped = true
}
func (ctx *P2PContext) Network() models.INetwork {
	return ctx._network
}

func (ctx *P2PContext) LocalNode() *node.LocalNode {
	return ctx._localNode
}

func (ctx *P2PContext) Connection() *TCPConnection {
	return ctx._connection
}

func (ctx *P2PContext) GetSkipped() bool {
	return ctx._skipped
}

func (ctx *P2PContext) ResetSkipped() {
	ctx._skipped = false
}

type IMiddleware interface {
	Handle(*P2PContext)
}
