package tcp

import (
	"time"

	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/p2p/models"
)

var (
	smLogger = log.GetLogger("SyncManager")
)

// SyncManager
// what to sync? block? file? config? rawdata?
// how to sync? from closer nodes? from every nodes? from those who has the content?
// sync details? blockid? filehash?
type SyncManager struct {
	// get the available nodes.
	nodeProvider models.INodeProvider
	// the network service
	network models.INetwork
	// what to sync
	syncProvider models.ISyncProvider
}

func NewSyncManager(nodeProvider models.INodeProvider, network models.INetwork, syncProvider models.ISyncProvider) *SyncManager {
	return &SyncManager{
		nodeProvider, network, syncProvider,
	}
}

func (s *SyncManager) test() {

}

func (s *SyncManager) syncLoop() {
	for {
		// loop through the nodes.
		nodes := s.nodeProvider.PeekNodes()
		smLogger.Trace("Peek peers got: %v", len(nodes))
		for _, n := range nodes {
			smLogger.Trace("Trying to sync from peer: %v", n.GetID())
			isSuccess := s.syncProvider.SendSyncRequest(s.network, s.nodeProvider.GetLocalNode(), n)
			if isSuccess {
				// break
			}
		}
		time.Sleep(5000 * time.Millisecond)
	}
}

func (s *SyncManager) Start() {
	go s.syncLoop()
}
