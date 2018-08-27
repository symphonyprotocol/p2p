package kad

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/symphonyprotocol/p2p/config"
	"github.com/symphonyprotocol/p2p/node"
)

var (
	//BUCKETS_TOTAL = 256
	BUCKETS_SIZE = 8
)

type INetwork interface {
	RegisterRefresh(f func(string, string, int))
	RegisterOffline(f func(string))
	Ping(nodeID string, ip net.IP, port int, waitChan chan bool)
}

type KTable struct {
	network   INetwork
	localNode *node.LocalNode
	buckets   map[int]*KBucket
}

func NewKTable(localNode *node.LocalNode, network INetwork) *KTable {
	buckets := make(map[int]*KBucket)
	kt := &KTable{
		network:   network,
		localNode: localNode,
		buckets:   buckets,
	}
	network.RegisterRefresh(kt.refresh)
	network.RegisterOffline(kt.offline)
	kt.loadInitNodes()
	return kt
}

func (t *KTable) loadInitNodes() {
	staticNodes := initialStaticNodes()
	for _, node := range staticNodes {
		t.add(node)
	}
}

func (t *KTable) add(remoteNode *node.RemoteNode) {
	if remoteNode.Distance == -1 {
		dist := distance(t.localNode.GetIDBytes(), remoteNode.GetIDBytes())
		remoteNode.Distance = dist
	}
	if bucket, ok := t.buckets[remoteNode.Distance]; ok {
		if bucket.Search(remoteNode.GetID()) == nil {
			bucket.Add(remoteNode)
		} else {
			bucket.MoveToTail(remoteNode)
		}
	} else {
		bucket := NewKBucket()
		bucket.Add(remoteNode)
		t.buckets[remoteNode.Distance] = bucket
	}
}

func (t *KTable) peekNodes() []*node.RemoteNode {
	remotes := make([]*node.RemoteNode, 0)
	for _, bucket := range t.buckets {
		node := bucket.Peek()
		if node != nil && node.GetID() != t.localNode.GetID() {
			remotes = append(remotes, node)
		}
	}
	return remotes
}

func (t *KTable) offline(nodeID string) {
	log.Printf("node offline %v\n", nodeID)
	id, _ := hex.DecodeString(nodeID)
	dist := distance(t.localNode.GetIDBytes(), id)
	if bucket, ok := t.buckets[dist]; ok {
		if rnode := bucket.Search(nodeID); rnode != nil {
			bucket.Remove(rnode)
		}

	}
}

func (t *KTable) refresh(nodeID string, ip string, port int) {
	log.Printf("node refresh %v\n", nodeID)
	id, _ := hex.DecodeString(nodeID)
	dist := distance(t.localNode.GetIDBytes(), id)
	if bucket, ok := t.buckets[dist]; ok {
		rnode := bucket.Search(nodeID)
		if rnode != nil {
			log.Printf("refresh exist node：%v, %v, %v\n", ip, port, dist)
			rnode.RefreshNode(ip, port)
			bucket.MoveToTail(rnode)
		} else {
			log.Printf("refresh to add new node: %v, %v, %v\n", ip, port, dist)
			ipAddr := net.ParseIP(ip)
			rnode = node.NewRemoteNode(id, ipAddr, port)
			if bucket.Add(rnode) {
				return
			}
			log.Println("refresh to ping first node")
			//todo: ping first then decide to add, ignore this action
		}
	} else {
		log.Printf("refresh to add new bucket: %v, %v, %v\n", ip, port, dist)
		ipAddr := net.ParseIP(ip)
		rnode := node.NewRemoteNode(id, ipAddr, port)
		rnode.Distance = dist
		bucket := NewKBucket()
		bucket.Add(rnode)
		t.buckets[dist] = bucket
	}
}

func (t *KTable) pingPong(rnode *node.RemoteNode) bool {
	ch := make(chan bool)
	t.network.Ping(t.localNode.GetID(), rnode.GetIP(), rnode.GetPort(), ch)
	return <-ch
}

func (t *KTable) ping(rnode *node.RemoteNode) {
	t.network.Ping(rnode.GetID(), rnode.GetIP(), rnode.GetPort(), nil)
}

func (t *KTable) Start() {
	go func() {
		for {
			time.Sleep(10 * time.Second)
			nodes := t.peekNodes()
			if len(nodes) == 0 {
				t.loadInitNodes()
			}
			for _, rnode := range nodes {
				t.ping(rnode)
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
		rnode := node.NewRemoteNode(id, ip, snode.Port)
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
