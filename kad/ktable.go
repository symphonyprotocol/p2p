package kad

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/symphonyprotocol/p2p/models"

	"github.com/symphonyprotocol/p2p/config"
	"github.com/symphonyprotocol/p2p/interfaces"
	"github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/utils"
)

var (
	//BUCKETS_TOTAL = 256
	BUCKETS_SIZE = 8
)

type RecieveData struct {
	RemoteAddr *net.UDPAddr
	Bytes      []byte
}

type waitReply struct {
	MesageID   string
	SendTs     int64
	ExpireTs   int64
	RemoteNode *node.RemoteNode
	IsTimeout  bool
}

type KTable struct {
	network   interfaces.INetwork
	localNode *node.LocalNode
	buckets   map[int]*KBucket
	waitlist  sync.Map
}

func NewKTable(localNode *node.LocalNode, network interfaces.INetwork) *KTable {
	buckets := make(map[int]*KBucket)
	kt := &KTable{
		network:   network,
		localNode: localNode,
		buckets:   buckets,
	}
	kt.loadInitNodes()
	network.RegisterCallback(KTABLE_DIAGRAM_CATEGORY, kt.callback)
	return kt
}

func (t *KTable) loadInitNodes() {
	staticNodes := initialStaticNodes()
	for _, node := range staticNodes {
		if node.GetID() == t.localNode.GetID() {
			continue
		}
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
		if node != nil {
			if node.GetID() == t.localNode.GetID() {
				t.refresh(node.GetID(), node.GetIP().String(), node.GetPort(), t.localNode.GetIP().String(), t.localNode.GetPort())
			} else {
				remotes = append(remotes, node)
			}
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

func (t *KTable) refresh(nodeID string, ip string, port int, localIP string, localPort int) {
	if nodeID == t.localNode.GetID() {
		return
	}
	if t.localNode.GetExtIP().String() == ip {
		log.Printf("node in same subnet map %v:%v -> %v:%v", ip, port, localIP, localPort)
		ip = localIP
		port = localPort
	}
	id, _ := hex.DecodeString(nodeID)
	dist := distance(t.localNode.GetIDBytes(), id)
	if bucket, ok := t.buckets[dist]; ok {
		rnode := bucket.Search(nodeID)
		if rnode != nil {
			log.Printf("refresh exist node：%v, %v, %v\n", ip, port, dist)
			rnode.RefreshNode(ip, port, localIP, localPort)
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

func (t *KTable) addWaitReply(msgID string, sendTs int64, expireTs int64, rnode *node.RemoteNode) {
	wait := waitReply{
		MesageID:   msgID,
		SendTs:     sendTs,
		ExpireTs:   expireTs,
		RemoteNode: rnode,
		IsTimeout:  false,
	}
	t.waitlist.Store(msgID, wait)
}

func (t *KTable) ping(rnode *node.RemoteNode) {
	//t.network.Ping(rnode.GetID(), rnode.GetIP(), rnode.GetPort(), nil)
	id := utils.NewUUID()
	ts := time.Now().Unix()
	exprie := ts + int64(models.DEFAULT_TIMEOUT)
	ping := PingDiagram{
		UDPDiagram: models.UDPDiagram{
			ID:        id,
			NodeID:    t.localNode.GetID(),
			Timestamp: ts,
			DCategory: KTABLE_DIAGRAM_CATEGORY,
			DType:     KTABLE_DIAGRAM_PING,
			Version:   models.UDP_DIAGRAM_VERSION,
			Expire:    exprie,
			LocalAddr: t.localNode.GetIP().String(),
			LocalPort: t.localNode.GetPort(),
		},
	}
	t.network.Send(rnode.GetIP(), rnode.GetPort(), utils.DiagramToBytes(ping))
	t.addWaitReply(ping.ID, ping.Timestamp, ping.Expire, rnode)
	log.Printf("send ping to %v:%v\n", rnode.GetIP().String(), rnode.GetPort())
}

func (t *KTable) pongAction(data []byte) {
	var pong PongDiagram
	utils.BytesToUDPDiagram(data, &pong)
	if t.localNode.GetExtIP().String() != pong.RemoteAddr || t.localNode.GetExtPort() != pong.RemotePort {
		t.localNode.SetExtIP(pong.RemoteAddr, pong.RemotePort)
	}
}

func (t *KTable) pong(msgID string, ip net.IP, port int) {
	ts := time.Now().Unix()
	expire := ts + int64(models.DEFAULT_TIMEOUT)
	pong := PongDiagram{
		UDPDiagram: models.UDPDiagram{
			ID:        msgID,
			NodeID:    t.localNode.GetID(),
			DCategory: KTABLE_DIAGRAM_CATEGORY,
			DType:     KTABLE_DIAGRAM_PONG,
			Version:   models.UDP_DIAGRAM_VERSION,
			Timestamp: ts,
			Expire:    expire,
			LocalAddr: t.localNode.GetIP().String(),
			LocalPort: t.localNode.GetPort(),
		},
		RemoteAddr: ip.String(),
		RemotePort: port,
	}
	t.network.Send(ip, port, utils.DiagramToBytes(pong))
	log.Printf("echo pong to %v:%v\n", ip.String(), port)
}

func (t *KTable) findNode(rnode *node.RemoteNode) {
	id := utils.NewUUID()
	ts := time.Now().Unix()
	exprie := ts + int64(models.DEFAULT_TIMEOUT)
	fn := FindNodeDiagram{
		UDPDiagram: models.UDPDiagram{
			ID:        id,
			NodeID:    t.localNode.GetID(),
			Timestamp: ts,
			DCategory: KTABLE_DIAGRAM_CATEGORY,
			DType:     KTABLE_DIAGRAM_FINDNODE,
			Version:   models.UDP_DIAGRAM_VERSION,
			Expire:    exprie,
			LocalAddr: t.localNode.GetIP().String(),
			LocalPort: t.localNode.GetPort(),
		},
	}
	t.addWaitReply(id, ts, exprie, rnode)
	t.network.Send(rnode.GetIP(), rnode.GetPort(), utils.DiagramToBytes(fn))
	log.Printf("send find node to %v:%v\n", rnode.GetIP().String(), rnode.GetPort())
}

func (t *KTable) findNodeAction(msgID string, nodeID string, ip net.IP, port int) {
	nodes := t.findNodeFromBuckets(nodeID)
	nodeDiagrams := make([]NodeDiagram, 0)
	for _, n := range nodes {
		nd := NodeDiagram{
			NodeID:    n.GetID(),
			IP:        n.GetIP().String(),
			Port:      n.GetPort(),
			LocalAddr: n.GetLocalIP().String(),
			LocalPort: n.GetLocalPort(),
		}
		nodeDiagrams = append(nodeDiagrams, nd)
	}
	ts := time.Now().Unix()
	exprie := ts + int64(models.DEFAULT_TIMEOUT)
	resp := FindNodeRespDiagram{
		UDPDiagram: models.UDPDiagram{
			ID:        msgID,
			NodeID:    t.localNode.GetID(),
			Timestamp: ts,
			DCategory: KTABLE_DIAGRAM_CATEGORY,
			DType:     KTABLE_DIAGRAM_FINDNODERESP,
			Version:   models.UDP_DIAGRAM_VERSION,
			Expire:    exprie,
			LocalAddr: t.localNode.GetIP().String(),
			LocalPort: t.localNode.GetPort(),
		},
		Nodes: nodeDiagrams,
	}
	t.network.Send(ip, port, utils.DiagramToBytes(resp))
	log.Printf("echo find node resp to %v:%v\n", ip.String(), port)
}

func (t *KTable) findNodeFromBuckets(nodeID string) []*node.RemoteNode {
	nodes := make([]*node.RemoteNode, 0)
	id, _ := hex.DecodeString(nodeID)
	dist := distance(t.localNode.GetIDBytes(), id)
	var i, j int
	i = dist
	j = dist + 1
	for i >= 0 || j < 256 {
		if bucket, ok := t.buckets[i]; i >= 0 && ok {
			inodes := bucket.GetAll()
			for _, ind := range inodes {
				if len(nodes) < BUCKETS_SIZE {
					nodes = append(nodes, ind)
				} else {
					break
				}
			}
		}
		i--
		if bucket, ok := t.buckets[j]; j < 256 && ok {
			jnodes := bucket.GetAll()
			for _, jnd := range jnodes {
				if len(nodes) < BUCKETS_SIZE {
					nodes = append(nodes, jnd)
				} else {
					break
				}
			}
		}
		j++
		if len(nodes) >= BUCKETS_SIZE {
			break
		}
	}
	return nodes
}

func (t *KTable) findNodeResp(data []byte) {
	var resp FindNodeRespDiagram
	utils.BytesToUDPDiagram(data, &resp)
	for _, n := range resp.Nodes {
		t.refresh(n.NodeID, n.IP, n.Port, n.LocalAddr, n.LocalPort)
	}
}

func (t *KTable) callback(params models.CallbackParams) {
	if obj, ok := t.waitlist.Load(params.Diagram.ID); ok {
		wait := obj.(waitReply)
		t.waitlist.Delete(wait.MesageID)
		if wait.IsTimeout {
			return
		}
	}
	t.refresh(params.Diagram.NodeID, params.RemoteAddr.IP.String(), params.RemoteAddr.Port, params.Diagram.LocalAddr, params.Diagram.LocalPort)
	switch params.Diagram.DType {
	case KTABLE_DIAGRAM_PING:
		t.pong(params.Diagram.ID, params.RemoteAddr.IP, params.RemoteAddr.Port)
	case KTABLE_DIAGRAM_PONG:
		t.pongAction(params.Data)
	case KTABLE_DIAGRAM_FINDNODE:
		t.findNodeAction(params.Diagram.ID, params.Diagram.NodeID, params.RemoteAddr.IP, params.RemoteAddr.Port)
	case KTABLE_DIAGRAM_FINDNODERESP:
		t.findNodeResp(params.Data)
	default:
	}
}

func (t *KTable) timeoutCallback(wait waitReply) {
	t.offline(wait.RemoteNode.GetID())
}

func (t *KTable) Start() {
	go t.loopPing()
	go t.loopTimeout()
	go t.loopFindNode()
}

func (t *KTable) loopTimeout() {
	for {
		var expire int64
		var messageID string
		t.waitlist.Range(func(key, value interface{}) bool {
			wait := value.(waitReply)
			if expire < wait.ExpireTs {
				expire = wait.ExpireTs
				messageID = key.(string)
			}
			return true
		})
		if expire == 0 {
			time.Sleep(100 * time.Microsecond)
			continue
		}
		delta := expire - time.Now().Unix()
		if delta > 0 {
			timer := time.NewTimer(time.Duration(delta) * time.Second)
			<-timer.C
		}
		if obj, ok := t.waitlist.Load(messageID); ok {
			t.waitlist.Delete(messageID)
			wait := obj.(waitReply)
			wait.IsTimeout = true
			t.timeoutCallback(wait)
		}
	}
}

func (t *KTable) loopPing() {
	for {
		nodes := t.peekNodes()
		if len(nodes) == 0 {
			t.loadInitNodes()
		}
		for _, rnode := range nodes {
			t.ping(rnode)
		}
		time.Sleep(10 * time.Second)
	}
}

func (t *KTable) loopFindNode() {
	time.Sleep(12 * time.Second)
	for {
		nodes := t.peekNodes()
		for _, rnode := range nodes {
			t.findNode(rnode)
		}
		time.Sleep(10 * time.Second)
	}
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
