package kad

import (
	"sync"

	"github.com/symphonyprotocol/p2p/node"
)

type KBucket struct {
	mux   sync.RWMutex
	nodes []*node.RemoteNode
}

func NewKBucket() *KBucket {
	return &KBucket{
		nodes: make([]*node.RemoteNode, 0),
	}
}

func (b *KBucket) GetAll() []*node.RemoteNode {
	return b.nodes
}

func (b *KBucket) Add(remoteNode *node.RemoteNode) bool {
	b.mux.Lock()
	defer b.mux.Unlock()
	if len(b.nodes) < BUCKETS_SIZE {
		b.nodes = append(b.nodes, remoteNode)
		return true
	}
	return false
}

func (b *KBucket) Peek() *node.RemoteNode {
	b.mux.RLock()
	defer b.mux.RUnlock()
	if len(b.nodes) == 0 {
		return nil
	}
	return b.nodes[0]
}

func (b *KBucket) Search(nodeID string) *node.RemoteNode {
	b.mux.RLock()
	defer b.mux.RUnlock()
	for _, rnode := range b.nodes {
		if rnode.GetID() == nodeID {
			return rnode
		}
	}
	return nil
}

func (b *KBucket) Size() int {
	return len(b.nodes)
}

func (b *KBucket) Remove(remoteNode *node.RemoteNode) {
	b.mux.Lock()
	defer b.mux.Unlock()
	if len(b.nodes) == 0 {
		return
	} else if len(b.nodes) == 1 {
		b.nodes = make([]*node.RemoteNode, 0)
	} else {
		var idx = -1
		for id, rnode := range b.nodes {
			if rnode == remoteNode {
				idx = id
				break
			}
		}
		if idx == -1 {
			return
		}
		if idx+1 == len(b.nodes) {
			b.nodes = b.nodes[:idx]
		} else {
			b.nodes = append(b.nodes[:idx], b.nodes[idx+1:]...)
		}
	}
}

func (b *KBucket) MoveToTail(remoteNode *node.RemoteNode) {
	b.mux.Lock()
	b.mux.Unlock()
	if len(b.nodes) <= 1 {
		return
	}
	rIndex := -1
	for i := 0; i < len(b.nodes); i++ {
		if b.nodes[i] == remoteNode {
			rIndex = i
			break
		}
	}
	if len(b.nodes) == (rIndex + 1) {
		return
	}
	b.nodes = append(b.nodes[rIndex+1:], remoteNode)
}
