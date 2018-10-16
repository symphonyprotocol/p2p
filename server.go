package p2p

import (
	"fmt"

	"github.com/symphonyprotocol/p2p/models"

	"github.com/symphonyprotocol/p2p/kad"
	"github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/p2p/udp"
)

type P2PServer struct {
	node        *node.LocalNode
	ktable      models.INodeProvider
	udpService  models.INetwork
	tcpService  models.INetwork
	syncManager *tcp.SyncManager
	middlewares []tcp.IMiddleware
	quit        chan int
}

func NewP2PServer() *P2PServer {
	node := node.NewLocalNode()
	udpService := udp.NewUDPService(node.GetID(), node.GetLocalIP(), node.GetLocalPort())
	sTcpService := tcp.NewSecuredTCPService(node)
	ktable := kad.NewKTable(node, udpService)
	syncManager := tcp.NewSyncManager(ktable, sTcpService, tcp.NewFileSyncProvider())
	srv := &P2PServer{
		node:        node,
		ktable:      ktable,
		udpService:  udpService,
		tcpService:  sTcpService,
		quit:        make(chan int),
		syncManager: syncManager,
		middlewares: make([]tcp.IMiddleware, 10),
	}
	return srv
}

func (s *P2PServer) Start() {
	s.node.DiscoverNAT()
	fmt.Println(s.node)
	s.udpService.Start()
	s.tcpService.Start()
	s.tcpService.RegisterCallback("default", func(p models.ICallbackParams) {
		if params, ok := p.(tcp.TCPCallbackParams); ok {
			ctx := tcp.NewP2PContext(s.tcpService, s.node, params.Connection)
			for _, middleware := range s.middlewares {
				middleware.Handle(ctx)
				if ctx.GetSkipped() {
					ctx.ResetSkipped()
					break
				}
			}
		}
	})
	s.ktable.Start()
	s.syncManager.Start()
	defer close(s.quit)
	<-s.quit
}

func (s *P2PServer) Use(m tcp.IMiddleware) {
	s.middlewares = append(s.middlewares, m)
}

func (s *P2PServer) Broadcast() {
	
}

func (s *P2PServer) Close() {
	s.quit <- 1
}
