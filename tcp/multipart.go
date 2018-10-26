package tcp

import (
	"github.com/symphonyprotocol/p2p/models"
)

type MultipartTCPDiagram struct {
	models.TCPDiagram
	ChunkSize	int			// size of rawData
	ChunkNo		int			// index of all the chunks
	ChunksCount	int			// size of chunks
	RawData		[]byte		// part of a TCP Diagram or just rawData
	ChunkTotalSize	int		// size of all
}

func (m *MultipartTCPDiagram) GetChunkSize() int { return m.ChunkSize }
func (m *MultipartTCPDiagram) GetChunkNo() int { return m.ChunkNo }
func (m *MultipartTCPDiagram) GetChunksCount() int { return m.ChunksCount }
func (m *MultipartTCPDiagram) GetRawData() []byte { return m.RawData }
func (m *MultipartTCPDiagram) GetChunkTotalSize() int { return m.ChunkTotalSize }
