package p2p

import (
	"github.com/symphonyprotocol/p2p/utils"
	"time"
	"math/rand"
	"github.com/symphonyprotocol/p2p/models"
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/log"
)

var syncLogger = log.GetLogger("example - syncLogger")


type BlockSyncMiddleware struct {
	blockHeight	int
}

func (b *BlockSyncMiddleware) Start(p *tcp.P2PContext) {
	rand.Seed(time.Now().Unix())
	b.blockHeight = rand.Intn(50)
	go func() {
		// randomly increase the block height
		for {
			time.Sleep(time.Duration(rand.Intn(50000)) * time.Millisecond + 50)
			r := rand.Intn(50)
			syncLogger.Debug("My Block Height increased by %v from %v", r, b.blockHeight)
			b.blockHeight = b.blockHeight + r
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
				TCPDiagram: tDiag,
				MyBlockHeight: b.blockHeight,
			})
		}
	}()
}

type InvDiagram struct {
	models.TCPDiagram
	MyBlockHeight	int
}

type GetBlockDiagram struct {
	models.TCPDiagram
	TargetBlockHeight	int
	CurrentBlockHeight	int
}

func (b *BlockSyncMiddleware) Handle(ctx *tcp.P2PContext) {
	diag := ctx.Params().GetDiagram()
	if (diag.GetDType() == "/inv") {
		var invDiag InvDiagram
		utils.BytesToUDPDiagram(ctx.Params().Data, &invDiag)
		if invDiag.ID != "" {
			syncLogger.Debug("We got a good inv diag with height: %v, my current height is: %v", invDiag.MyBlockHeight, b.blockHeight)

			// send back the block height
			tDiag := models.NewTCPDiagram()
			tDiag.NodeID = ctx.LocalNode().GetID()
			tDiag.DType = "/inv_res"
			diag := InvDiagram{
				TCPDiagram: tDiag,
				MyBlockHeight: b.blockHeight,
			}
			ctx.Send(diag)
		} else {
			syncLogger.Error("BOOM, the diag we got is not inv diag ??? That's impossible !!!")
		}
	}

	if (diag.GetDType() == "/inv_res") {
		var invDiag InvDiagram
		utils.BytesToUDPDiagram(ctx.Params().Data, &invDiag)
		if invDiag.ID != "" {
			// good
			if invDiag.MyBlockHeight > b.blockHeight {
				// remote is newer. ask for block
				syncLogger.Debug("Remote blocks %v are newer than us %v, will ask for blocks from %v", invDiag.MyBlockHeight, b.blockHeight, ctx.Params().Connection.GetNodeID())
				tDiag := models.NewTCPDiagram()
				tDiag.NodeID = ctx.LocalNode().GetID()
				tDiag.DType = "/getblock"
				diag := GetBlockDiagram{
					TCPDiagram: tDiag,
					TargetBlockHeight: invDiag.MyBlockHeight,
					CurrentBlockHeight: b.blockHeight,
				}
				ctx.Send(diag)
			}
		}
	}

	if (diag.GetDType() == "/getblock") {
		var getBDiag GetBlockDiagram
		utils.BytesToUDPDiagram(ctx.Params().Data, &getBDiag)
		if getBDiag.ID != "" {
			syncLogger.Debug("We got a good getblock diag with target height: %v and its current height: %v, my current height is: %v", getBDiag.TargetBlockHeight, getBDiag.CurrentBlockHeight, b.blockHeight)
			
			tDiag := models.NewTCPDiagram()
			tDiag.NodeID = ctx.LocalNode().GetID()
			tDiag.DType = "/getblock_res"
			diag := GetBlockDiagram{
				TCPDiagram: tDiag,
				TargetBlockHeight: b.blockHeight,
				CurrentBlockHeight: b.blockHeight,
			}
			ctx.Send(diag)
		} else {
			syncLogger.Error("BOOM, the diag we got is not getblock diag ??? That's impossible !!!")
		}
	}

	if (diag.GetDType() == "/getblock_res") {
		var getBDiag GetBlockDiagram
		utils.BytesToUDPDiagram(ctx.Params().Data, &getBDiag)
		if getBDiag.ID != "" {
			if (getBDiag.TargetBlockHeight > b.blockHeight) {
				// newer than me. update
				syncLogger.Debug("Remote blocks are newer, will update blocks from %v to %v", b.blockHeight, getBDiag.TargetBlockHeight)
				b.blockHeight = getBDiag.TargetBlockHeight
			}
		}
	}
}

func (b *BlockSyncMiddleware) AcceptConnection(conn *tcp.TCPConnection) {
	syncLogger.Debug("accepted new connection with %v", conn.GetNodeID())
}

func (b *BlockSyncMiddleware) DropConnection(conn *tcp.TCPConnection) {
	syncLogger.Debug("dropped connection with %v", conn.GetNodeID())
}
