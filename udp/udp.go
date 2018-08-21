package udp

import (
	"encoding/json"
	"fmt"
	"github.com/symphonyprotocol/p2p/kad"
	"github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/utils"
	"log"
	"net"
	"sync"
	"time"
)

var (
	SEND_RETRY = 3
)

type UDPOption struct {
	LocalNode *node.LocalNode
	KTable    *kad.KTable
}

type UDPService struct {
	option     *UDPOption
	listener   *net.UDPConn
	port       int
	ip         net.IP
	mux        sync.Mutex
	lastMsgMap map[string]int64
}

func NewUDPService(ip net.IP, port int, option *UDPOption) *UDPService {
	client := &UDPService{
		port: port,
		ip:   ip,
	}
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: ip, Port: port})
	if err != nil {
		panic(err)
	}
	client.listener = listener
	client.option = option
	client.lastMsgMap = make(map[string]int64)
	return client
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
		rdata := utils.UnzipDiagramBytes(data[:n])
		diagram := BytesToUDPDiagram(rdata)
		if ts, ok := c.lastMsgMap[diagram.DType]; ok {
			if ts < diagram.Timestamp {
				c.lastMsgMap[diagram.DType] = diagram.Timestamp
			} else {
				continue
			}
		} else {
			c.lastMsgMap[diagram.DType] = diagram.Timestamp
		}
		if int(time.Now().Unix()-diagram.Timestamp) > diagram.ExpireDuration {
			continue
		}
		switch diagram.DType {
		case UDP_DIAGRAM_PING:
			go c.handlePing(rdata, remoteAddr)
		case UDP_DIAGRAM_PONG:
			go c.handlerPong(rdata)
		default:
		}
	}
}

func (c *UDPService) Dial(ip net.IP, port int, bytes []byte) {
	c.mux.Lock()
	dstAddr := &net.UDPAddr{IP: ip, Port: port}
	//log.Printf("send udp data to %v\n", dstAddr)
	send := 0
	for send < SEND_RETRY {
		_, err := c.listener.WriteToUDP(bytes, dstAddr)
		if err != nil {
			fmt.Printf("send UDP to target %v error:%v", dstAddr, err)
		}
		send++
	}
	c.mux.Unlock()
}

func (c *UDPService) Start() {
	go c.loop()
	go c.pingNodeLoop()
}

func (c *UDPService) handlerPong(data []byte) {
	pong := PongDiagram{}
	err := json.Unmarshal(data, &pong)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(pong)
}

func (c *UDPService) handlePing(data []byte, remoteAddr *net.UDPAddr) {
	ping := PingDiagram{}
	err := json.Unmarshal(data, &ping)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(ping)
	diagram := PongDiagram{
		UDPDiagram: UDPDiagram{
			Timestamp:      time.Now().Unix(),
			DType:          UDP_DIAGRAM_PONG,
			Version:        1,
			ExpireDuration: 5,
		},
		NodeID:     c.option.LocalNode.GetID(),
		LocalAddr:  c.option.LocalNode.GetIP().String(),
		LocalPort:  c.option.LocalNode.GetUdpPort(),
		RemoteAddr: remoteAddr.IP.String(),
		RemotePort: remoteAddr.Port,
	}
	bytes := utils.ZipDiagramToBytes(diagram)
	c.Dial(remoteAddr.IP, remoteAddr.Port, bytes)
}

func (c *UDPService) pingNodeLoop() {
	for {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("ping node err:%v", err)
			}
		}()
		time.Sleep(5 * time.Second)
		nodes := c.option.KTable.PeekNodes()
		fmt.Println(nodes)
		for _, n := range nodes {
			diagram := PingDiagram{
				UDPDiagram: UDPDiagram{
					Timestamp:      time.Now().Unix(),
					DType:          UDP_DIAGRAM_PING,
					Version:        1,
					ExpireDuration: 5,
				},
				NodeID:    c.option.LocalNode.GetID(),
				LocalAddr: c.option.LocalNode.GetIP().String(),
				LocalPort: c.option.LocalNode.GetUdpPort(),
			}
			bytes := utils.ZipDiagramToBytes(diagram)
			c.Dial(n.GetIP(), n.GetUdpPort(), bytes)
		}
	}
}

func (c *UDPService) disvoerNodes() {
}
