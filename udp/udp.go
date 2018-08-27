package udp

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/symphonyprotocol/p2p/models"
	"github.com/symphonyprotocol/p2p/utils"
)

type UDPService struct {
	listener    *net.UDPConn
	localNodeID string
	port        int
	ip          net.IP
	callbacks   sync.Map
}

func NewUDPService(localNodeID string, ip net.IP, port int) *UDPService {
	client := &UDPService{
		localNodeID: localNodeID,
		port:        port,
		ip:          ip,
	}
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: ip, Port: port})
	if err != nil {
		panic(err)
	}
	client.listener = listener
	return client
}

func (c *UDPService) RegisterCallback(category string, callback func(models.CallbackParams)) {
	c.callbacks.Store(category, callback)
}

func (c *UDPService) RemoveCallback(category string) {
	c.callbacks.Delete(category)
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
		var diagram models.UDPDiagram
		utils.BytesToUDPDiagram(rdata, &diagram)
		if obj, ok := c.callbacks.Load(diagram.DCategory); ok {
			callback := obj.(func(models.CallbackParams))
			callback(models.CallbackParams{
				RemoteAddr: remoteAddr,
				Diagram:    diagram,
				Data:       rdata,
			})
		}
	}
}

func (c *UDPService) Send(ip net.IP, port int, bytes []byte) {
	dstAddr := &net.UDPAddr{IP: ip, Port: port}
	//log.Printf("send udp data to %v\n", dstAddr)
	_, err := c.listener.WriteToUDP(bytes, dstAddr)
	if err != nil {
		fmt.Printf("send UDP to target %v error:%v", dstAddr, err)
	}
}

func (c *UDPService) Start() {
	go c.loop()
}
