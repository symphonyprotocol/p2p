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
	if t.localNode.GetID() == remoteNode.GetID() {
		return
	}
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

func (t *KTable) refresh(nodeID string, localIP string, localPort int, remoteIP string, remotePort int) {
	if nodeID == t.localNode.GetID() {
		return
	}

	id, _ := hex.DecodeString(nodeID)
	dist := distance(t.localNode.GetIDBytes(), id)
	log.Printf("refresh exist node：%v:%v -> %v:%v, %v\n", localIP, localPort, remoteIP, remotePort, dist)
	if bucket, ok := t.buckets[dist]; ok {
		rnode := bucket.Search(nodeID)
		if rnode != nil {
			//log.Printf("refresh exist node：%v, %v, %v\n", remoteIP, remotePort, dist)
			rnode.RefreshNode(localIP, localPort, remoteIP, remotePort)
			bucket.MoveToTail(rnode)
		} else {
			//log.Printf("refresh to add new node: %v, %v, %v\n", remoteIP, remotePort, dist)
			localAddr := net.ParseIP(localIP)
			remoteAddr := net.ParseIP(remoteIP)
			rnode = node.NewRemoteNode(id, localAddr, localPort, remoteAddr, remotePort)
			if bucket.Add(rnode) {
				return
			}
			//log.Println("refresh to ping first node")
			//todo: ping first then decide to add, ignore this action
		}
	} else {
		log.Printf("refresh to add new bucket: %v, %v, %v\n", remoteIP, remotePort, dist)
		localAddr := net.ParseIP(localIP)
		remoteAddr := net.ParseIP(remoteIP)
		rnode := node.NewRemoteNode(id, localAddr, localPort, remoteAddr, remotePort)
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

func (t *KTable) send(rnode *node.RemoteNode, data []byte) {
	ip, port := rnode.GetSendIPWithPort(t.localNode)
	t.network.Send(ip, port, data)
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
			LocalAddr: t.localNode.GetLocalIP().String(),
			LocalPort: t.localNode.GetLocalPort(),
		},
	}
	t.send(rnode, utils.DiagramToBytes(ping))
	t.addWaitReply(ping.ID, ping.Timestamp, ping.Expire, rnode)
	log.Printf("send ping to %v:%v\n", rnode.GetRemoteIP().String(), rnode.GetRemotePort())
}

func (t *KTable) pongAction(data []byte) {
	var pong PongDiagram
	utils.BytesToUDPDiagram(data, &pong)
	if t.localNode.GetRemoteIP().String() != pong.RemoteAddr || t.localNode.GetRemotePort() != pong.RemotePort {
		t.localNode.SetRemoteIPPort(pong.RemoteAddr, pong.RemotePort)
	}
}

func (t *KTable) pong(diagram models.UDPDiagram, remoteAddr *net.UDPAddr) {
	ts := time.Now().Unix()
	expire := ts + int64(models.DEFAULT_TIMEOUT)
	pong := PongDiagram{
		UDPDiagram: models.UDPDiagram{
			ID:        diagram.ID,
			NodeID:    t.localNode.GetID(),
			DCategory: KTABLE_DIAGRAM_CATEGORY,
			DType:     KTABLE_DIAGRAM_PONG,
			Version:   models.UDP_DIAGRAM_VERSION,
			Timestamp: ts,
			Expire:    expire,
			LocalAddr: t.localNode.GetLocalIP().String(),
			LocalPort: t.localNode.GetLocalPort(),
		},
		RemoteAddr: remoteAddr.IP.String(),
		RemotePort: remoteAddr.Port,
	}
	rnode := node.NewRemoteNode([]byte(diagram.NodeID), net.ParseIP(diagram.LocalAddr), diagram.LocalPort, remoteAddr.IP, remoteAddr.Port)
	t.send(rnode, utils.DiagramToBytes(pong))
	log.Printf("echo pong to %v:%v\n", rnode.GetRemoteIP().String(), rnode.GetRemotePort())
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
			LocalAddr: t.localNode.GetLocalIP().String(),
			LocalPort: t.localNode.GetLocalPort(),
		},
	}
	t.send(rnode, utils.DiagramToBytes(fn))
	t.addWaitReply(id, ts, exprie, rnode)
	log.Printf("send find node to %v:%v\n", rnode.GetRemoteIP().String(), rnode.GetRemotePort())
}

func (t *KTable) findNodeAction(msgID string, nodeID string, ip net.IP, port int) {
	nodes := t.findNodeFromBuckets(nodeID)
	nodeDiagrams := make([]NodeDiagram, 0)
	for _, n := range nodes {
		nd := NodeDiagram{
			NodeID:     n.GetID(),
			LocalAddr:  n.GetLocalIP().String(),
			LocalPort:  n.GetLocalPort(),
			RemoteIP:   n.GetRemoteIP().String(),
			RemotePort: n.GetRemotePort(),
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
			LocalAddr: t.localNode.GetLocalIP().String(),
			LocalPort: t.localNode.GetLocalPort(),
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
		t.refresh(n.NodeID, n.LocalAddr, n.LocalPort, n.RemoteIP, n.RemotePort)
	}
	log.Println("recieve find node resp")
}

func (t *KTable) callback(params models.CallbackParams) {
	if obj, ok := t.waitlist.Load(params.Diagram.ID); ok {
		wait := obj.(waitReply)
		t.waitlist.Delete(wait.MesageID)
		if wait.IsTimeout {
			return
		}
	}
	t.refresh(params.Diagram.NodeID, params.Diagram.LocalAddr, params.Diagram.LocalPort, params.RemoteAddr.IP.String(), params.RemoteAddr.Port)
	switch params.Diagram.DType {
	case KTABLE_DIAGRAM_PING:
		t.pong(params.Diagram, params.RemoteAddr)
	case KTABLE_DIAGRAM_PONG:
		log.Printf("recieve pong from %v:%v\n", params.RemoteAddr.IP.String(), params.RemoteAddr.Port)
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
		rnode := node.NewRemoteNode(id, ip, snode.Port, ip, snode.Port)
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
