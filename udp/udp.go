package udp

import (
	"fmt"
	"github.com/symphonyprotocol/p2p/utils"
	"log"
	"net"
	"sync"
	"time"
)

type RecieveData struct {
	RemoteAddr *net.UDPAddr
	Data       []byte
}

type WaitReply struct {
	MessageID   string
	NodeID      string
	IP          net.IP
	Port        int
	RecieveData RecieveData
	ExprieTs    int64
	WaitHandler WaitHandler
	RemoteAddr  *net.UDPAddr
	IsTimeout   bool
	Retry       int
}

type WaitHandler func(wait WaitReply)

type UDPService struct {
	listener *net.UDPConn
	port     int
	ip       net.IP
	waitList sync.Map
}

func NewUDPService(ip net.IP, port int) *UDPService {
	client := &UDPService{
		port: port,
		ip:   ip,
	}
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: ip, Port: port})
	if err != nil {
		panic(err)
	}
	client.listener = listener
	return client
}

func (c *UDPService) WaitHandler(wait WaitReply) {
	if wait.IsTimeout {
		log.Printf("recieve %v with timeout", wait.MessageID)
	} else {
		log.Printf("recieve %v with data", wait.MessageID)
	}
	count := 0
	c.waitList.Range(func(key, value interface{}) bool {
		count += 1
		return true
	})
	log.Printf("waitlist length: %v", count)
}

func (c *UDPService) addWaitReply(msgId string, nodeID string, expire int64) {
	wait := WaitReply{
		MessageID:   msgId,
		NodeID:      nodeID,
		IP:          c.ip,
		Port:        c.port,
		ExprieTs:    expire,
		WaitHandler: WaitHandler(c.WaitHandler),
	}
	c.waitList.Store(msgId, wait)
}

func (c *UDPService) Ping(nodeID string, ip net.IP, port int) {
	id := utils.NewUUID()
	ts := time.Now().Unix()
	expire := ts + int64(DEFAULT_TIMEOUT)
	ping := PingDiagram{
		UDPDiagram: UDPDiagram{
			ID:        id,
			Timestamp: ts,
			DType:     UDP_DIAGRAM_PING,
			Version:   UDP_DIAGRAM_VERSION,
			Expire:    expire,
		},
		NodeID:    nodeID,
		LocalAddr: c.ip.String(),
		LocalPort: c.port,
	}
	c.addWaitReply(ping.ID, nodeID, expire)
	c.send(ip, port, utils.DiagramToBytes(ping))
}

func (c *UDPService) Pong(msgID string, nodeID string, remoteAddr *net.UDPAddr) {
	ts := time.Now().Unix()
	expire := ts + int64(DEFAULT_TIMEOUT)
	pong := PongDiagram{
		UDPDiagram: UDPDiagram{
			ID:        msgID,
			Timestamp: ts,
			DType:     UDP_DIAGRAM_PONG,
			Version:   UDP_DIAGRAM_VERSION,
			Expire:    expire,
		},
		NodeID:     nodeID,
		LocalAddr:  c.ip.String(),
		LocalPort:  c.port,
		RemoteAddr: remoteAddr.IP.String(),
		RemotePort: remoteAddr.Port,
	}
	c.send(remoteAddr.IP, remoteAddr.Port, utils.DiagramToBytes(pong))
}

func (c *UDPService) loop() {
	log.Println("start listenning udp...")
	data := make([]byte, 1280)
	for {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("UPDServer loop err:%V\n", err)
			}
		}()
		n, remoteAddr, err := c.listener.ReadFromUDP(data)
		if err != nil {
			fmt.Printf("error during read: %v", err)
		}
		//log.Println(remoteAddr, data[:n])
		rdata := data[:n]
		diagram := UDPDiagram{}
		utils.BytesToUDPDiagram(rdata, &diagram)
		if iwait, ok := c.waitList.Load(diagram.ID); ok {
			wait := iwait.(WaitReply)
			c.waitList.Delete(wait.MessageID)
			wait.WaitHandler(wait)
			wait.IsTimeout = false
		} else {
			now := time.Now().Unix()
			if now-diagram.Timestamp > 0 {
				continue
			}
			if diagram.DType == UDP_DIAGRAM_PING {
				ping := PingDiagram{}
				utils.BytesToUDPDiagram(rdata, &ping)
				c.Pong(ping.ID, ping.NodeID, remoteAddr)
			}
		}
	}
}

func (c *UDPService) loopTimeout() {
	log.Println("start loop timeout...")
	for {
		var msgId string
		var ts int64
		c.waitList.Range(func(key, value interface{}) bool {
			tmp := value.(WaitReply)
			if ts < tmp.ExprieTs {
				ts = tmp.ExprieTs
				msgId = tmp.MessageID
			}
			return true
		})
		if ts == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		delta := ts - time.Now().Unix()
		if delta > 0 {
			log.Printf("wait %v for %v sec", msgId, delta)
			timer := time.NewTimer(time.Second * time.Duration(delta))
			<-timer.C
		}

		if iwait, ok := c.waitList.Load(msgId); ok {
			log.Println("find timeout:" + msgId)
			wait := iwait.(WaitReply)
			c.waitList.Delete(msgId)
			wait.IsTimeout = true
			wait.WaitHandler(wait)
		}
	}
}

func (c *UDPService) send(ip net.IP, port int, bytes []byte) {
	dstAddr := &net.UDPAddr{IP: ip, Port: port}
	log.Printf("send udp data to %v\n", dstAddr)
	_, err := c.listener.WriteToUDP(bytes, dstAddr)
	if err != nil {
		fmt.Printf("send UDP to target %v error:%v", dstAddr, err)
	}
}

func (c *UDPService) Start() {
	go c.loop()
	go c.loopTimeout()
}
