package tcp

import (
	"time"
	"github.com/symphonyprotocol/p2p/interfaces"
	"github.com/symphonyprotocol/p2p/utils"
	"github.com/symphonyprotocol/p2p/models"
	"github.com/symphonyprotocol/p2p/node"
)

// try to sync files in ~/biu

type fileSyncDiagram struct {
	models.TCPDiagram
}

func newFileSyncDiagram(ln *node.LocalNode) fileSyncDiagram {
	return fileSyncDiagram{
		models.TCPDiagram{
			DCategory: "shit",
			NodeID: ln.GetID(),
			Timestamp: time.Now().Unix(),
		},
	}
}

type FileSyncProvider struct {
	interfaces.ISyncProvider
}

func NewFileSyncProvider() *FileSyncProvider {
	return &FileSyncProvider {
	}
}

func (f *FileSyncProvider) SendSyncRequest(network interfaces.INetwork, ln *node.LocalNode, n *node.RemoteNode) bool {
	network.Send(n.GetRemoteIP(), n.GetRemotePort(), utils.DiagramToBytes(newFileSyncDiagram(ln)), n.GetID())
	return true
}
