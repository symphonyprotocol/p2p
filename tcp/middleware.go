package tcp

import (
	"github.com/symphonyprotocol/p2p/utils"
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/models"
)

var p2pLogger = log.GetLogger("P2PService")

type P2PContext struct {
	_skipped    bool
	_network    models.INetwork
	_localNode  *node.LocalNode
	_nodeProvider	models.INodeProvider
	_params 	*TCPCallbackParams
}

func NewP2PContext(network models.INetwork, localNode *node.LocalNode, nodeProvider models.INodeProvider, params *TCPCallbackParams) *P2PContext {
	return &P2PContext{
		_skipped:    false,
		_network:    network,
		_localNode:  localNode,
		_nodeProvider:	nodeProvider,
		_params:	 params,
	}
}

func (ctx *P2PContext) Send(diag models.IDiagram) {
	length, err := ctx._params.Connection.Write(utils.DiagramToBytes(diag))
	if err != nil {
		p2pLogger.Error("Failed to send packet (%d) to %v", length, ctx._params.Connection.RemoteAddr().String())
	} else {
		p2pLogger.Trace("Packet (%d) sent to %v", length, ctx._params.Connection.RemoteAddr().String())
	}
}

func (ctx *P2PContext) Broadcast(diag models.IDiagram) {
	peers := ctx._nodeProvider.PeekNodes()
	for _, peer := range peers {
		p2pLogger.Trace("Broadcasting message %v to peer %v (%v:%v)", diag.GetID(), peer.GetID(), peer.GetRemoteIP().String(), peer.GetRemotePort())
		ctx._network.Send(peer.GetRemoteIP(), peer.GetRemotePort(), utils.DiagramToBytes(diag), peer.GetID())
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

func (ctx *P2PContext) NodeProvider() models.INodeProvider {
	return ctx._nodeProvider
}

func (ctx *P2PContext) Params() *TCPCallbackParams {
	return ctx._params
}

func (ctx *P2PContext) GetSkipped() bool {
	return ctx._skipped
}

func (ctx *P2PContext) ResetSkipped() {
	ctx._skipped = false
}

// Get the diagram as input 
func (ctx *P2PContext) GetDiagram(diagRef interface{}) {
	utils.BytesToUDPDiagram(ctx.Params().Data, diagRef)
}

type IMiddleware interface {
	Handle(*P2PContext)
	Start(*P2PContext)
	AcceptConnection(*TCPConnection)
	DropConnection(*TCPConnection)
}
