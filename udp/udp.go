package udp

import (
	"fmt"
	"net"
	"sync"

	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/p2p/models"
	"github.com/symphonyprotocol/p2p/utils"
)

var logger = log.GetLogger("udp")

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

func (c *UDPService) RegisterCallback(category string, callback func(models.ICallbackParams)) {
	c.callbacks.Store(category, callback)
}

func (c *UDPService) RemoveCallback(category string) {
	c.callbacks.Delete(category)
}

func (c *UDPService) loop() {
	logger.Trace("start listenning udp...")
	data := make([]byte, 1280)
	for {
		defer func() {
			if err := recover(); err != nil {
				logger.Trace("UPDServer loop err:%v", err)
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
			callback := obj.(func(models.ICallbackParams))
			callback(models.UDPCallbackParams{
				CallbackParams: models.CallbackParams{
					RemoteAddr: remoteAddr,
					Diagram:    diagram,
					Data:       rdata,
				},
			})
		}
	}
}

func (c *UDPService) Send(ip net.IP, port int, bytes []byte, nodeId string) (int, error) {
	dstAddr := &net.UDPAddr{IP: ip, Port: port}
	//logger.Trace("send udp data to %v", dstAddr)
	length, err := c.listener.WriteToUDP(bytes, dstAddr)
	if err != nil {
		fmt.Printf("send UDP to target %v error:%v", dstAddr, err)
	}

	return length, err
}

func (c *UDPService) Start() {
	go c.loop()
}
