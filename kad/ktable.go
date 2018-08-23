package kad

import (
	"encoding/hex"
	"fmt"
	"github.com/symphonyprotocol/p2p/config"
	"github.com/symphonyprotocol/p2p/node"
	"net"
	"strconv"
	"sync"
	"time"
)

var (
	//BUCKETS_TOTAL = 256
	BUCKETS_SIZE = 8
)

type INetwork interface {
	Ping(nodeID string, ip net.IP, port int)
}

type KTable struct {
	network INetwork
	localID []byte
	buckets map[int]*KBucket
	mux     sync.Mutex
}

func NewKTable(localID []byte, network INetwork) *KTable {
	buckets := make(map[int]*KBucket)
	kt := &KTable{
		network: network,
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

func (t *KTable) Refresh(nodeID string, ip string, port int) {
	id := []byte(nodeID)
	dist := distance(t.localID, id)
	if bucket, ok := t.buckets[dist]; ok {
		rnode := bucket.Search(nodeID)
		if rnode != nil {
			bucket.MoveToTail(nodeID)
		} else {
			ipAddr := net.ParseIP(ip)
			rnode = node.NewRemoteNode(id, ipAddr, port, port)
			if bucket.Add(rnode) {
				return
			}
			//todo: ping first then decide to add
		}
	} else {
		ipAddr := net.ParseIP(ip)
		rnode := node.NewRemoteNode(id, ipAddr, port, port)
		rnode.Distance = dist
		bucket := NewKBucket()
		bucket.Add(rnode)
		t.buckets[dist] = bucket
	}
}

func (t *KTable) PingPong(rnode *node.RemoteNode) {
	t.network.Ping(string(t.localID), rnode.GetIP(), rnode.GetUdpPort())
}

func (t *KTable) Start() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			for _, rnode := range t.PeekNodes() {
				t.PingPong(rnode)
			}
		}
	}()
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
