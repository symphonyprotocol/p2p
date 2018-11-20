package p2p

import (
	"github.com/symphonyprotocol/log"

	"github.com/symphonyprotocol/p2p/models"

	"github.com/symphonyprotocol/p2p/kad"
	"github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/p2p/udp"
)

var p2pLogger = log.GetLogger("p2pServer")

type P2PServer struct {
	node        *node.LocalNode
	ktable      models.INodeProvider
	udpService  models.INetwork
	tcpService  *tcp.TLSSecuredTCPService
	syncManager *tcp.SyncManager
	middlewares []tcp.IMiddleware
	quit        chan int
}

func NewP2PServer() *P2PServer {
	node := node.NewLocalNode()
	udpService := udp.NewUDPService(node.GetID(), node.GetLocalIP(), node.GetLocalPort())
	sTcpService := tcp.NewTLSSecuredTCPService(node)
	ktable := kad.NewKTable(node, udpService)
	syncManager := tcp.NewSyncManager(ktable, sTcpService, tcp.NewFileSyncProvider())
	srv := &P2PServer{
		node:        node,
		ktable:      ktable,
		udpService:  udpService,
		tcpService:  sTcpService,
		quit:        make(chan int),
		syncManager: syncManager,
		middlewares: make([]tcp.IMiddleware, 0, 10),
	}
	return srv
}

func (s *P2PServer) Start() {
	s.node.DiscoverNAT()
	p2pLogger.Debug("%v", s.node)
	s.udpService.Start()
	s.tcpService.Start()
	s.regTCPEvents()
	s.ktable.Start()
	s.startMiddlewares()
	// s.syncManager.Start()
	defer close(s.quit)
	<-s.quit
}

func (s *P2PServer) regTCPEvents() {
	s.tcpService.RegisterCallback("default", func(p models.ICallbackParams) {
		if params, ok := p.(tcp.TCPCallbackParams); ok {
			ctx := tcp.NewP2PContext(s.tcpService, s.node, s.ktable, &params, s.middlewares)
			
			// p2pLogger.Debug("Length of middlewares is %v", len(s.middlewares))
			for _, middleware := range s.middlewares {
				middleware.Handle(ctx)
				if ctx.GetSkipped() {
					ctx.ResetSkipped()
				} else {
					break
				}
			}
		}
	})

	s.tcpService.RegisterAcceptConnectionEvent(func(conn *tcp.TCPConnection) {
		for _, middleware := range s.middlewares {
			middleware.AcceptConnection(conn)
		}
	})

	s.tcpService.RegisterDropConnectionEvent(func(conn *tcp.TCPConnection) {
		for _, middleware := range s.middlewares {
			middleware.DropConnection(conn)
		}
	})
}

func (s *P2PServer) startMiddlewares() {
	ctx := tcp.NewP2PContext(s.tcpService, s.node, s.ktable, nil, s.middlewares)
	for _, middleware := range s.middlewares {
		middleware.Start(ctx)
	}
}

// use before start
func (s *P2PServer) Use(m tcp.IMiddleware) {
	s.middlewares = append(s.middlewares, m)
}

// NodeID will be set by P2PServer
func (s *P2PServer) NewP2PContext() *tcp.P2PContext {
	return tcp.NewP2PContext(s.tcpService, s.node, s.ktable, nil, s.middlewares)
}

func (s *P2PServer) Close() {
	s.quit <- 1
}
