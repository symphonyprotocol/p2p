package p2p

import (
	"fmt"

	"github.com/symphonyprotocol/p2p/kad"
	"github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/udp"
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/p2p/interfaces"
)

type P2PServer struct {
	node       *node.LocalNode
	ktable     *kad.KTable
	udpService *udp.UDPService
	tcpService interfaces.INetwork
	syncManager	*tcp.SyncManager
	quit       chan int
}

func NewP2PServer() *P2PServer {
	node := node.NewLocalNode()
	udpService := udp.NewUDPService(node.GetID(), node.GetLocalIP(), node.GetLocalPort())
	sTcpService := tcp.NewTCPService(node)
	ktable := kad.NewKTable(node, udpService)
	syncManager := tcp.NewSyncManager(ktable, sTcpService, tcp.NewFileSyncProvider())
	srv := &P2PServer{
		node:       node,
		ktable:     ktable,
		udpService: udpService,
		tcpService: sTcpService,
		quit:       make(chan int),
		syncManager: syncManager,
	}
	return srv
}

func (s *P2PServer) Start() {
	s.node.DiscoverNAT()
	fmt.Println(s.node)
	s.udpService.Start()
	s.tcpService.Start()
	s.ktable.Start()
	s.syncManager.Start()
	defer close(s.quit)
	<-s.quit
}

func (s *P2PServer) Close() {
	s.quit <- 1
}
