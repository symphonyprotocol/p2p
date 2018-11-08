package p2p

import (
	"math/big"
	"time"
	"math/rand"
	"github.com/symphonyprotocol/p2p/models"
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/log"
)

var syncLogger = log.GetLogger("example - syncLogger")

var BlockHeight	*big.Int

type BlockSyncMiddleware struct {
}

func (b *BlockSyncMiddleware) Start(p *tcp.P2PContext) {
	rand.Seed(time.Now().Unix())
	BlockHeight = big.NewInt(rand.Int63n(50))
	go func() {
		// randomly increase the block height
		for {
			time.Sleep(time.Duration(rand.Intn(50000)) * time.Millisecond + 50)
			r := rand.Int63n(50)
			added := big.NewInt(0)
			added.Add(BlockHeight, big.NewInt(r))
			syncLogger.Debug("My Block Height increased %v -> %v", BlockHeight, added)
			BlockHeight = added
		}
	}()

	go func() {
		for {
			time.Sleep(20 * time.Second)
			// force sync, not for real case, in real case, only restart the node will do the sync, or the node will be informed if there is news.
			tDiag := models.NewTCPDiagram()
			tDiag.NodeID = p.LocalNode().GetID()
			tDiag.DType = "/inv"
			p.Broadcast(InvDiagram{
				TCPDiagram: *tDiag,
				MyBlockHeight: BlockHeight,
			})
		}
	}()
}

type InvDiagram struct {
	models.TCPDiagram
	MyBlockHeight	*big.Int
}

type GetBlockDiagram struct {
	models.TCPDiagram
	TargetBlockHeight	*big.Int
	CurrentBlockHeight	*big.Int
}

func (b *BlockSyncMiddleware) Handle(ctx *tcp.P2PContext) {
	diag := ctx.Params().GetDiagram()
	if (diag.GetDType() == "/inv") {
		var invDiag InvDiagram
		err := ctx.GetDiagram(&invDiag)
		if err == nil {
			syncLogger.Debug("We got a good inv diag with height: %v, my current height is: %v", invDiag.MyBlockHeight, BlockHeight)

			// send back the block height
			tDiag := models.NewTCPDiagram()
			tDiag.NodeID = ctx.LocalNode().GetID()
			tDiag.DType = "/inv_res"
			diag := InvDiagram{
				TCPDiagram: *tDiag,
				MyBlockHeight: BlockHeight,
			}
			ctx.Send(diag)
		} else {
			syncLogger.Trace("BOOM, the diag we got is not inv diag ??? That's impossible !!!")
		}
	}

	if (diag.GetDType() == "/inv_res") {
		var invDiag InvDiagram
		err := ctx.GetDiagram(&invDiag)
		if err == nil {
			// good
			if invDiag.MyBlockHeight.Cmp(BlockHeight) > 0 {
				// remote is newer. ask for block
				syncLogger.Debug("Remote blocks %v are newer than us %v, will ask for blocks from %v", invDiag.MyBlockHeight, BlockHeight, ctx.Params().Connection.GetNodeID())
				tDiag := models.NewTCPDiagram()
				tDiag.NodeID = ctx.LocalNode().GetID()
				tDiag.DType = "/getblock"
				diag := GetBlockDiagram{
					TCPDiagram: *tDiag,
					TargetBlockHeight: invDiag.MyBlockHeight,
					CurrentBlockHeight: BlockHeight,
				}
				ctx.Send(diag)
			}
		}
	}

	if (diag.GetDType() == "/getblock") {
		var getBDiag GetBlockDiagram
		err := ctx.GetDiagram(&getBDiag)
		if err == nil {
			syncLogger.Debug("We got a good getblock diag with target height: %v and its current height: %v, my current height is: %v", getBDiag.TargetBlockHeight, getBDiag.CurrentBlockHeight, BlockHeight)
			
			tDiag := models.NewTCPDiagram()
			tDiag.NodeID = ctx.LocalNode().GetID()
			tDiag.DType = "/getblock_res"
			diag := GetBlockDiagram{
				TCPDiagram: *tDiag,
				TargetBlockHeight: BlockHeight,
				CurrentBlockHeight: BlockHeight,
			}
			ctx.Send(diag)
		} else {
			syncLogger.Error("BOOM, the diag we got is not getblock diag ??? That's impossible !!!")
		}
	}

	if (diag.GetDType() == "/getblock_res") {
		var getBDiag GetBlockDiagram
		err := ctx.GetDiagram(&getBDiag)
		if err == nil {
			if (getBDiag.TargetBlockHeight.Cmp(BlockHeight) > 0) {
				// newer than me. update
				syncLogger.Debug("Remote blocks are newer, will update blocks from %v to %v", BlockHeight, getBDiag.TargetBlockHeight)
				BlockHeight = getBDiag.TargetBlockHeight
			}
		}
	}
	
	ctx.Next()
}

func (b *BlockSyncMiddleware) AcceptConnection(conn *tcp.TCPConnection) {
	syncLogger.Debug("accepted new connection with %v", conn.GetNodeID())
}

func (b *BlockSyncMiddleware) DropConnection(conn *tcp.TCPConnection) {
	syncLogger.Debug("dropped connection with %v", conn.GetNodeID())
}

func (b *BlockSyncMiddleware) DashboardData() interface{} {
	return [][]string{
		[]string{
			"Current BlockHeight:", BlockHeight.String(),
		},
	}
}

func (b *BlockSyncMiddleware) DashboardType() string {
	return "table"
}

func (b *BlockSyncMiddleware) DashboardTitle() string {
	return "Middleware - Block Sync"
}

func (b *BlockSyncMiddleware) DashboardTableHasColumnTitles() bool {
	return false
}

func (b *BlockSyncMiddleware) Name() string {
	return "Block Sync"
}
