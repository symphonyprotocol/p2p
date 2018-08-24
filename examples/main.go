package main

import (
	"fmt"
	//"github.com/symphonyprotocol/p2p/config"
	"github.com/symphonyprotocol/p2p"
	"github.com/symphonyprotocol/p2p/encrypt"
	"github.com/symphonyprotocol/p2p/kad"
	"github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/utils"
	//"github.com/symphonyprotocol/p2p/udp"
	//"math/big"
)

func getId() []byte {
	privKey := encrypt.GenerateNodeKey()
	id := encrypt.PublicKeyToNodeId(privKey.PublicKey)
	return id
}

func getCurrentNode() *node.LocalNode {
	localNode := node.NewLocalNode()
	return localNode
}

func initialKtable() {
	ktable := kad.NewKTable(getCurrentNode(), nil)
	fmt.Println(ktable)
}

func initialServer() {
	srv := p2p.NewP2PServer()
	srv.Start()
}

type BaseDiagram struct {
	SequenceID int `json:"sequence_id"`
	Timestamp  int `json:"timestamp"`
}

type PingDiagram struct {
	BaseDiagram
	RemoteAddr string `json:"remote_addr"`
	RemotePort int    `json:"remote_port"`
	LocalAddr  string `json:"local_addr"`
	LocalPort  int    `json:"local_port"`
}

func testJson() {
	diagram := PingDiagram{
		BaseDiagram: BaseDiagram{
			SequenceID: 123,
			Timestamp:  456,
		},
		RemoteAddr: "192.168.0.1",
		RemotePort: 12306,
		LocalAddr:  "192.168.0.1",
		LocalPort:  12306,
	}
	data := utils.DiagramToBytes(diagram)
	fmt.Println(data)
	d := PingDiagram{}
	utils.BytesToUDPDiagram(data, &d)
	fmt.Println(d)
}

func main() {
	fmt.Println("hello p2p")
	initialServer()
	//testJson()
}
