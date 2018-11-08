package tcp

import (
	"time"

	"github.com/symphonyprotocol/p2p/models"
	"github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/utils"
)

// try to sync files in ~/biu

type fileSyncDiagram struct {
	models.TCPDiagram
}

func newFileSyncDiagram(ln *node.LocalNode) fileSyncDiagram {
	diagram := models.NewTCPDiagram()
	diagram.NodeID = ln.GetID()
	diagram.Timestamp = time.Now().Unix()
	return fileSyncDiagram{
		*diagram,
	}
}

type FileSyncProvider struct {
	models.ISyncProvider
}

func NewFileSyncProvider() *FileSyncProvider {
	return &FileSyncProvider{}
}

func (f *FileSyncProvider) SendSyncRequest(network models.INetwork, ln *node.LocalNode, n *node.RemoteNode) bool {
	network.Send(n.GetRemoteIP(), n.GetRemotePort(), utils.DiagramToBytes(newFileSyncDiagram(ln)), n.GetID())
	return true
}
