package tcp

import (
	"github.com/symphonyprotocol/p2p/interfaces"
	"github.com/symphonyprotocol/p2p/utils"
	"github.com/symphonyprotocol/p2p/models"
	"github.com/symphonyprotocol/p2p/node"
)

// try to sync files in ~/biu

type fileSyncDiagram struct {
	models.TCPDiagram
}

func newFileSyncDiagram() fileSyncDiagram {
	return fileSyncDiagram{
		models.TCPDiagram{
			DCategory: "shit",
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

func (f *FileSyncProvider) SendSyncRequest(network interfaces.INetwork, n *node.RemoteNode) bool {
	network.Send(n.GetRemoteIP(), n.GetRemotePort(), utils.DiagramToBytes(newFileSyncDiagram()), n.GetID())
	return true
}
