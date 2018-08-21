package kad

import (
	"encoding/hex"
	"fmt"
	"github.com/symphonyprotocol/p2p/config"
	//symen "github.com/symphonyprotocol/p2p/encrypt"
	"github.com/symphonyprotocol/p2p/node"
	"net"
	"strconv"
	"sync"
)

var (
	//BUCKETS_TOTAL = 256
	BUCKETS_SIZE = 8
)

type KTable struct {
	localID []byte
	buckets map[int]*KBucket
	mux     sync.Mutex
}

func NewKTable(localID []byte) *KTable {
	buckets := make(map[int]*KBucket)
	kt := &KTable{
		localID: localID,
		buckets: buckets,
	}

	staticNodes := initialStaticNodes()

	for _, node := range staticNodes {
		kt.Add(node)
	}
	return kt
}

func (t *KTable) Add(remoteNode *node.RemoteNode) {
	t.mux.Lock()
	if remoteNode.Distance == -1 {
		dist := distance(t.localID, remoteNode.GetIDBytes())
		remoteNode.Distance = dist
	}
	if bucket, ok := t.buckets[remoteNode.Distance]; ok {
		if bucket.Search(remoteNode.GetID()) == nil {
			bucket.Add(remoteNode)
		} else {
			bucket.MoveToTail(remoteNode.GetID())
		}
	} else {
		bucket := NewKBucket()
		bucket.Add(remoteNode)
		t.buckets[remoteNode.Distance] = bucket
	}
	t.mux.Unlock()
}

func (t *KTable) PeekNodes() []*node.RemoteNode {
	remotes := make([]*node.RemoteNode, 0)
	for _, bucket := range t.buckets {
		node := bucket.Peek()
		if node != nil {
			remotes = append(remotes, node)
		}
	}
	return remotes
}

func initialStaticNodes() []*node.RemoteNode {
	remoteNodes := make([]*node.RemoteNode, 0)
	nodes := config.LoadStaticNodes()
	for _, snode := range nodes.Nodes {
		id, _ := hex.DecodeString(snode.ID)
		ip := net.ParseIP(snode.IP)
		rnode := node.NewRemoteNode(id, ip, snode.UPort, snode.TPort)
		remoteNodes = append(remoteNodes, rnode)
	}
	return remoteNodes
}

func distance(a, b []byte) int {
	c := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		c[i] = a[i] ^ b[i]
	}
	r := fmt.Sprintf("%v", c[0])
	x, _ := strconv.Atoi(r)
	return x
}
