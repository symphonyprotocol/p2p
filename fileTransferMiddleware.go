package p2p

import (
	"github.com/symphonyprotocol/p2p/models"
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/log"
	"math/rand"
	"time"
	"math/big"
	"crypto/sha256"
	"strings"
	"fmt"
)


var fSyncLogger = log.GetLogger("example - fileSyncLogger")

type FileDiagram struct {
	models.TCPDiagram
	Bytes []byte
	FileHash	string
}

type FileTransferMiddleware struct {
	bytesSent	*big.Int
	bytesReceived *big.Int
	succeeded *big.Int
	failed *big.Int
}

func NewFileTransferMiddleware() *FileTransferMiddleware {
	return &FileTransferMiddleware{
		bytesSent: big.NewInt(0),
		bytesReceived: big.NewInt(0),
		succeeded: big.NewInt(0),
		failed: big.NewInt(0),
	}
}

func (d *FileTransferMiddleware) Handle(ctx *tcp.P2PContext) {
	diag := ctx.Params().GetDiagram()
	if (diag.GetDType() == "/file_sync") {
		var fileDiag FileDiagram
		err := ctx.GetDiagram(&fileDiag)
		if err == nil {
			fSyncLogger.Info("Good, file sync diag received, will check the sha256 sum")
			d.bytesReceived.Add(d.bytesReceived, big.NewInt(int64(len(fileDiag.Bytes))))
			h := sha256.New()
			h.Write(fileDiag.Bytes)
			hash := fmt.Sprintf("%x", h.Sum(nil))
			fSyncLogger.Debug("comparing hash we calculated: %v with the hash we got: %v", hash, fileDiag.FileHash)
			if strings.Compare(hash, fileDiag.FileHash) == 0 {
				fSyncLogger.Info("Good, hashes are the same")
				d.succeeded.Add(d.succeeded, big.NewInt(1))
			} else {
				fSyncLogger.Error("Boom, file transfer got damaged in the middle.")
				d.failed.Add(d.failed, big.NewInt(1))
			}
		} else {
			fSyncLogger.Error("Boom, what we got is not a 10M file sync diagram?")
		}
	}
	ctx.Next()
}
func (d *FileTransferMiddleware) Start(ctx *tcp.P2PContext) {
	
	rand.Seed(time.Now().Unix())
	// BlockHeight = big.NewInt(rand.Int63n(50))
	// go func() {
	// 	// randomly increase the block height
	// 	for {
	// 		time.Sleep(time.Duration(rand.Intn(50000)) * time.Millisecond + 50)
	// 		r := rand.Int63n(50)
	// 		added := big.NewInt(0)
	// 		added.Add(BlockHeight, big.NewInt(r))
	// 		syncLogger.Debug("My Block Height increased %v -> %v", BlockHeight, added)
	// 		BlockHeight = added
	// 	}
	// }()

	go func() {
		for {
			time.Sleep(300 * time.Second)
			// force sync, not for real case, in real case, only restart the node will do the sync, or the node will be informed if there is news.
			tDiag := models.NewTCPDiagram()
			tDiag.DType = "/file_sync"
			tDiag.NodeID = ctx.LocalNode().GetID()
			bytes := d.randomSizeBytes()
			h := sha256.New()
			h.Write(bytes)
			hash := h.Sum(nil)
			d.bytesSent.Add(d.bytesSent, big.NewInt(int64(len(bytes))))
			ctx.Broadcast(FileDiagram{
				TCPDiagram: *tDiag,
				Bytes: bytes,
				FileHash: fmt.Sprintf("%x", hash),
			})
		}
	}()
}

func (d *FileTransferMiddleware) AcceptConnection(*tcp.TCPConnection) {

}
func (d *FileTransferMiddleware) DropConnection(*tcp.TCPConnection) {

}

func (b *FileTransferMiddleware) DashboardData() interface{} {
	return [][]string{
		[]string{
			"Bytes Sent:", b.bytesSent.String(),
		},
		[]string{
			"Bytes Recieved:", b.bytesReceived.String(),
		},
		[]string{
			"Succeeded Transfer:", b.succeeded.String(),
		},
		[]string{
			"Failed Transfer:", b.failed.String(),
		},
	}
}

func (b *FileTransferMiddleware) DashboardType() string {
	return "table"
}

func (b *FileTransferMiddleware) DashboardTitle() string {
	return "Middleware - File Transfer (Big []byte)"
}

func (b *FileTransferMiddleware) DashboardTableHasColumnTitles() bool {
	return false
}

func (b *FileTransferMiddleware) Name() string {
	return "Dashboard"
}

func (b *FileTransferMiddleware) randomSizeBytes() []byte {
	size := 5000000 + rand.Intn(500000)
	res := make([]byte, size, size)
	for n, _ := range res {
		res[n] = byte(rand.Intn(255))
	}
	return res
}

