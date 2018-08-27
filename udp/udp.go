package udp

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/symphonyprotocol/p2p/utils"
)

type ITable interface {
	Refresh(nodeID string, ip string, port int)
}

type recieveData struct {
	RemoteAddr *net.UDPAddr
	Data       []byte
}

type waitReply struct {
	MessageID   string
	NodeID      string
	IP          net.IP
	Port        int
	RecieveData recieveData
	ExprieTs    int64
	WaitHandler waitHandler
	RemoteAddr  *net.UDPAddr
	IsTimeout   bool
	Retry       int
	WaitChan    chan bool
}

type waitHandler func(wait waitReply)

type UDPService struct {
	listener *net.UDPConn
	port     int
	ip       net.IP
	waitList sync.Map
	refresh  func(string, string, int)
	offline  func(string)
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

func (c *UDPService) waitHandler(wait waitReply) {
	if wait.WaitChan != nil {
		wait.WaitChan <- !wait.IsTimeout
	}
	if wait.IsTimeout {
		if c.offline != nil {
			c.offline(wait.NodeID)
		}
	} else {
		if c.refresh != nil {
			c.refresh(wait.NodeID, wait.IP.String(), wait.Port)
		}
	}
	count := 0
	c.waitList.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	log.Printf("waitlist length: %v", count)
}

func (c *UDPService) addWaitReply(msgID string, nodeID string, ip net.IP, port int, expire int64, waitChan chan bool) {
	wait := waitReply{
		MessageID:   msgID,
		NodeID:      nodeID,
		IP:          ip,
		Port:        port,
		ExprieTs:    expire,
		WaitHandler: waitHandler(c.waitHandler),
		WaitChan:    waitChan,
	}
	c.waitList.Store(msgID, wait)
}

func (c *UDPService) RegisterRefresh(f func(string, string, int)) {
	c.refresh = f
}

func (c *UDPService) RegisterOffline(f func(string)) {
	c.offline = f
}

func (c *UDPService) Ping(nodeID string, ip net.IP, port int, waitChan chan bool) {
	id := utils.NewUUID()
	ts := time.Now().Unix()
	expire := ts + int64(DEFAULT_TIMEOUT)
	ping := PingDiagram{
		UDPDiagram: UDPDiagram{
			ID:        id,
			NodeID:    nodeID,
			Timestamp: ts,
			DType:     UDP_DIAGRAM_PING,
			Version:   UDP_DIAGRAM_VERSION,
			Expire:    expire,
		},
		LocalAddr: c.ip.String(),
		LocalPort: c.port,
	}
	log.Printf("ping %v:%v with msg id:%v\n", ip, port, id)
	c.addWaitReply(ping.ID, nodeID, ip, port, expire, waitChan)
	c.send(ip, port, utils.DiagramToBytes(ping))
}

func (c *UDPService) Pong(msgID string, nodeID string, remoteAddr *net.UDPAddr) {
	ts := time.Now().Unix()
	expire := ts + int64(DEFAULT_TIMEOUT)
	pong := PongDiagram{
		UDPDiagram: UDPDiagram{
			ID:        msgID,
			NodeID:    nodeID,
			Timestamp: ts,
			DType:     UDP_DIAGRAM_PONG,
			Version:   UDP_DIAGRAM_VERSION,
			Expire:    expire,
		},
		LocalAddr:  c.ip.String(),
		LocalPort:  c.port,
		RemoteAddr: remoteAddr.IP.String(),
		RemotePort: remoteAddr.Port,
	}
	log.Printf("pong %v:%v with msg id:%v\n", remoteAddr.IP.String(), remoteAddr.Port, msgID)
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
		rdata := data[:n]
		diagram := UDPDiagram{}
		utils.BytesToUDPDiagram(rdata, &diagram)
		if iwait, ok := c.waitList.Load(diagram.ID); ok {
			log.Println("recieve callback")
			wait := iwait.(waitReply)
			c.waitList.Delete(wait.MessageID)
			wait.WaitHandler(wait)
			wait.IsTimeout = false
		} else {
			log.Println("recieve new request")
			if diagram.DType == UDP_DIAGRAM_PING {
				ping := PingDiagram{}
				utils.BytesToUDPDiagram(rdata, &ping)
				c.Pong(ping.ID, ping.NodeID, remoteAddr)
			}
			if c.refresh != nil {
				c.refresh(diagram.NodeID, remoteAddr.IP.String(), remoteAddr.Port)
			}
		}
	}
}

func (c *UDPService) loopTimeout() {
	log.Println("start loop timeout...")
	for {
		var msgID string
		var ts int64
		c.waitList.Range(func(key, value interface{}) bool {
			tmp := value.(waitReply)
			if ts < tmp.ExprieTs {
				ts = tmp.ExprieTs
				msgID = tmp.MessageID
			}
			return true
		})
		if ts == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		delta := ts - time.Now().Unix()
		if delta > 0 {
			//log.Printf("wait %v for %v sec", msgID, delta)
			timer := time.NewTimer(time.Second * time.Duration(delta))
			<-timer.C
		}

		if iwait, ok := c.waitList.Load(msgID); ok {
			log.Println("find timeout:" + msgID)
			wait := iwait.(waitReply)
			c.waitList.Delete(msgID)
			wait.IsTimeout = true
			wait.WaitHandler(wait)
		}
	}
}

func (c *UDPService) send(ip net.IP, port int, bytes []byte) {
	dstAddr := &net.UDPAddr{IP: ip, Port: port}
	//log.Printf("send udp data to %v\n", dstAddr)
	_, err := c.listener.WriteToUDP(bytes, dstAddr)
	if err != nil {
		fmt.Printf("send UDP to target %v error:%v", dstAddr, err)
	}
}

func (c *UDPService) Start() {
	go c.loop()
	go c.loopTimeout()
}
