package interfaces

import (
	"net"

	"github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/models"
)

type INetwork interface {
	RemoveCallback(category string)
	RegisterCallback(category string, callback func(models.CallbackParams))
	Send(ip net.IP, port int, bytes []byte, nodeId string)
	Start()
}

type INodeProvider interface {
	PeekNodes() []*node.RemoteNode
	GetLocalNode() *node.LocalNode 
}

type ISyncProvider interface {
	SendSyncRequest(network INetwork, remoteNode *node.RemoteNode) bool
}
