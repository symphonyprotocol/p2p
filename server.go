package p2p

import (
	"fmt"

	"github.com/symphonyprotocol/p2p/kad"
	"github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/udp"
)

type P2PServer struct {
	node       *node.LocalNode
	ktable     *kad.KTable
	udpService *udp.UDPService
	quit       chan int
}

func NewP2PServer() *P2PServer {
	node := node.NewLocalNode()
	udpService := udp.NewUDPService(node.GetIP(), node.GetPort())
	ktable := kad.NewKTable(node, udpService)
	srv := &P2PServer{
		node:       node,
		ktable:     ktable,
		udpService: udpService,
		quit:       make(chan int),
	}
	return srv
}

func (s *P2PServer) Start() {
	s.node.DiscoverNAT()
	fmt.Println(s.node)
	s.udpService.Start()
	s.ktable.Start()
	defer close(s.quit)
	<-s.quit
}

func (s *P2PServer) Close() {
	s.quit <- 1
}
