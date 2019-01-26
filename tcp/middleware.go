package tcp

import (
	"reflect"
	"fmt"
	"sync"
	"github.com/symphonyprotocol/p2p/utils"
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/models"
	"time"
)

var mLogger = log.GetLogger("middleware").SetLevel(log.INFO)
var TCP_CHUNK_SIZE = 500
var multipartDiagramMap sync.Map
var _multipartDiagramPartsMap sync.Map
var msgBroadcasted sync.Map

type P2PContext struct {
	_skipped    bool
	_network    models.INetwork
	_localNode  *node.LocalNode
	_nodeProvider	models.INodeProvider
	_params 	*TCPCallbackParams
	_middlewares	[]IMiddleware
}

func NewP2PContext(network models.INetwork, localNode *node.LocalNode, nodeProvider models.INodeProvider, params *TCPCallbackParams, middlewares []IMiddleware) *P2PContext {
	return &P2PContext{
		_skipped:    false,
		_network:    network,
		_localNode:  localNode,
		_nodeProvider:	nodeProvider,
		_params:	 params,
		_middlewares: middlewares,
	}
}

func (ctx *P2PContext) Send(diag models.IDiagram) {
	ctx.chunkDiagram(diag, func(bytes []byte) {
		ctx._params.Connection.WriteBytes(bytes)
	})
}

func (ctx *P2PContext) NewTCPDiagram() *models.TCPDiagram {
	tDiag := models.NewTCPDiagram()
	tDiag.NodeID = ctx.LocalNode().GetID()
	return tDiag
}

func (ctx *P2PContext) Broadcast(diag models.IDiagram) {
	ctx.BroadcastWithFilter(diag, func(p *node.RemoteNode) bool { return true })
}

func (ctx *P2PContext) BroadcastWithFilter(diag models.IDiagram, filter func(_p *node.RemoteNode) bool) {
	ctx.BroadcastToPeers(diag, ctx._nodeProvider.PeekNodes(), filter)
}

func (ctx *P2PContext) BroadcastToNearbyNodes(diag models.IDiagram, maxNodeCount int, filter func (_p *node.RemoteNode) bool) {
	ctx.BroadcastToPeers(diag, ctx._nodeProvider.GetNearbyNodes(maxNodeCount), filter)
}

func (ctx *P2PContext) BroadcastToPeers(diag models.IDiagram, peers []*node.RemoteNode, filter func(_p *node.RemoteNode) bool) {
	if _msg, _ok := msgBroadcasted.Load(diag.GetID()); _ok {
		if ts, ok := _msg.(time.Time); ok {
			if time.Since(ts) < time.Hour {
				mLogger.Trace("This msg was broadcasted in an hour and we will not broadcast it again.")
				return 
			} else {
				mLogger.Warn("Try to broadcast it again after an hour, this should not happend so much.")
				msgBroadcasted.Delete(diag.GetID())
			}
		}
	}

	for _, peer := range peers {
		if filter == nil || filter(peer) {
			mLogger.Trace("Broadcasting message %v to peer %v (%v:%v)", diag.GetID(), peer.GetID(), peer.GetRemoteIP().String(), peer.GetRemotePort())
			msgBroadcasted.Store(diag.GetID(), time.Now())
			ctx.SendToPeer(diag, peer)
		} else {
			mLogger.Trace("Node %v filtered to be excluded when broadcasting", peer.GetID())
		}
	}
}

func (ctx *P2PContext) SendToPeer(diag models.IDiagram, peer *node.RemoteNode) {
	ctx.chunkDiagram(diag, func(bytes []byte) {
		ctx._network.Send(peer.GetRemoteIP(), peer.GetRemotePort(), bytes, peer.GetID())
	})
}

func (ctx *P2PContext) chunkDiagram(diag models.IDiagram, callback func([]byte)) {
	bytes := utils.DiagramToBytes(diag)
	lenBytes := len(bytes)
	chunksCount := lenBytes / TCP_CHUNK_SIZE + 1
	dId := utils.NewUUID()
	for i := 0; i < chunksCount; i++ {
		tDiag := ctx.NewTCPDiagram()
		tDiag.ID = dId		// use same id for the diagrams
		tDiag.DCategory = diag.GetDCategory()
		tDiag.DType = diag.GetDType()
		
		end := (i+1) * TCP_CHUNK_SIZE
		if end > lenBytes {
			end = lenBytes
		}
		mDiag := &MultipartTCPDiagram{
			TCPDiagram: *tDiag,
			ChunksCount: chunksCount,
			ChunkNo: i,
			RawData: bytes[i * TCP_CHUNK_SIZE: end],
			ChunkSize: end - (i * TCP_CHUNK_SIZE),
			ChunkTotalSize: lenBytes,
		}
		bytesDiag := utils.DiagramToBytes(mDiag)
		if callback != nil {
			callback(bytesDiag)
			// if err != nil {
			// 	mLogger.Error("Failed to send packet")
			// } else {
			mLogger.Trace(
				"Packet (%d) sent with chunksCount: %v, chunkNo: %v, chunkSize: %v, chunkTotalSize: %v", 
				len(bytesDiag), 
				mDiag.ChunksCount,
				mDiag.ChunkNo,
				mDiag.ChunkSize,
				mDiag.ChunkTotalSize)
		}
	}
}

func (ctx *P2PContext) Next() {
	ctx._skipped = true
}
func (ctx *P2PContext) Network() models.INetwork {
	return ctx._network
}

func (ctx *P2PContext) LocalNode() *node.LocalNode {
	return ctx._localNode
}

func (ctx *P2PContext) NodeProvider() models.INodeProvider {
	return ctx._nodeProvider
}

func (ctx *P2PContext) Params() *TCPCallbackParams {
	return ctx._params
}

func (ctx *P2PContext) Middlewares() []IMiddleware {
	return ctx._middlewares
}

func (ctx *P2PContext) GetSkipped() bool {
	return ctx._skipped
}

func (ctx *P2PContext) ResetSkipped() {
	ctx._skipped = false
}

// Get the diagram as input 
func (ctx *P2PContext) GetDiagram(diagRef interface{}) error {
	var mDiag MultipartTCPDiagram
	if err := utils.BytesToUDPDiagram(ctx.Params().Data, &mDiag); err == nil && mDiag.GetChunksCount() > 0 {
		mLogger.Trace("I got a multipart diagram from %v", ctx.Params().GetRemoteAddr().String())
		result := ctx.ResolveMultipartDiagram(mDiag)
		if result != nil {
			mLogger.Trace("multipart diagram constructed from %v, going to convert it \n%v \nto diagram: %v", 
				ctx.Params().GetRemoteAddr().String(),
				result,
				reflect.TypeOf(diagRef).Elem(),
			)
			_err := utils.BytesToUDPDiagram(result, diagRef)
			if _, _ok := msgBroadcasted.Load(mDiag.GetID()); _ok {
				mLogger.Warn("Got a message that was broadcasted by me ! drop it")
				return fmt.Errorf("This message is broadcasted by me !!!, DROP IT")
			}
			return _err
		} else {
			mLogger.Trace("multipart diagram not ready from %v", ctx.Params().GetRemoteAddr().String())
			return fmt.Errorf("Diagram is multipart, not done yet")
		}
	} else {
		mLogger.Trace("diagram is not multipart")
		return utils.BytesToUDPDiagram(ctx.Params().Data, diagRef)
	}
}

func (ctx *P2PContext) ResolveMultipartDiagram(mDiag MultipartTCPDiagram) []byte {
	mLogger.Trace(
		"Resolving multipart diagram, chunkSize: %v, chunkNo: %v, chunkTotalSize: %v, chunkCount: %v", 
		mDiag.GetChunkSize(),
		mDiag.GetChunkNo(), 
		mDiag.GetChunkTotalSize(), 
		mDiag.GetChunksCount())
	var bytes []byte = nil
	// this is multipart diagram... need to wait
	if _bytes, ok := multipartDiagramMap.Load(mDiag.GetID()); ok {
		if __bytes, ok := _bytes.([]byte); ok {
			bytes = __bytes
		}
	} else {
		bytes = make([]byte, mDiag.GetChunkTotalSize(), mDiag.GetChunkTotalSize())
		multipartDiagramMap.Store(mDiag.GetID(), bytes)
	}

	if bytes != nil {
		start := mDiag.GetChunkNo() * TCP_CHUNK_SIZE 
		for i := start; i < start + mDiag.GetChunkSize(); i++ {
			bytes[i] = mDiag.GetRawData()[i - start]
		}

		var ids []bool
		if _ids, ok := _multipartDiagramPartsMap.Load(mDiag.GetID()); ok {
			if __ids, ok := _ids.([]bool); ok {
				ids = __ids
			}	
		} else {
			ids = make([]bool, mDiag.GetChunksCount(), mDiag.GetChunksCount())
			_multipartDiagramPartsMap.Store(mDiag.GetID(), ids)
		}

		ids[mDiag.GetChunkNo()] = true

		// check if bytes finished?
		var notFinished = false
		whatIsInIds := ""
		for _, b := range ids {
			if b == false {
				notFinished = true
				whatIsInIds += "0"
			} else {
				whatIsInIds += "1"
			}
		}

		// mLogger.Debug("What is in Ids now is: %v", whatIsInIds)

		if !notFinished {
			
			// broadcast the diagram to middlewares
			multipartDiagramMap.Delete(mDiag.GetID())
			_multipartDiagramPartsMap.Delete(mDiag.GetID())
			return bytes
		}
	}

	return nil
}

type IMiddleware interface {
	models.IDashboardProvider
	Handle(*P2PContext)
	Start(*P2PContext)
	AcceptConnection(*TCPConnection)
	DropConnection(*TCPConnection)
	Name() string
}

type BaseMiddleware struct {
	reqHandlers map[string]func(*P2PContext)
	startCtx	*P2PContext
	workChan	chan *P2PContext
}

func NewBaseMiddleware() *BaseMiddleware {
	return &BaseMiddleware{
		reqHandlers: make(map[string]func(*P2PContext)),
		workChan: make(chan *P2PContext),
	}
}

func (b *BaseMiddleware) Handle(ctx *P2PContext) {
	diag := ctx.Params().GetDiagram()
	dType := diag.GetDType()
	if handler, ok := b.reqHandlers[dType]; ok {
		handler(ctx)
	}

	ctx.Next()
}

func (b *BaseMiddleware) HandleRequest(reqPath string, handler func (*P2PContext)) {
	b.reqHandlers[reqPath] = handler
}

func (b *BaseMiddleware) Start(ctx *P2PContext) { 
	b.startCtx = ctx 
	// go b.HandleWorkLoop()
}
func (b *BaseMiddleware) AcceptConnection(*TCPConnection) {}
func (b *BaseMiddleware) DropConnection(*TCPConnection) {}

func (b *BaseMiddleware) Name() string {
	return "BaseMiddleware"
}

func (b *BaseMiddleware) DashboardData() interface{} { return nil }
func (b *BaseMiddleware) DashboardType() string { return "" }
func (b *BaseMiddleware) DashboardTitle() string { return "" }
func (b *BaseMiddleware) DashboardTableHasColumnTitles() bool { return false }
