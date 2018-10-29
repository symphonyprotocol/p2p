package models

import (
	"net"

	"github.com/symphonyprotocol/p2p/node"
)

type IDiagram interface {
	GetID()        string
	GetNodeID()    string
	GetTimestamp() int64
	GetDCategory() string
	GetDType()     string
	GetVersion()   int
}

type ICallbackParams interface {
	GetRemoteAddr() net.Addr
	GetDiagram()    IDiagram
	GetData()       []byte
}

type INetwork interface {
	RemoveCallback(category string)
	RegisterCallback(category string, callback func(ICallbackParams))
	Send(ip net.IP, port int, bytes []byte, nodeId string)
	Start()
}

type INodeProvider interface {
	PeekNodes() []*node.RemoteNode
	GetLocalNode() *node.LocalNode 
	Start()
}

type ISyncProvider interface {
	SendSyncRequest(network INetwork, localNode *node.LocalNode, remoteNode *node.RemoteNode) bool
}

type IDashboardProvider interface {
	DashboardData() interface{}	// [][]string for table, []string for list
	DashboardType() string
	DashboardTitle() string
	DashboardTableHasColumnTitles() bool
}
