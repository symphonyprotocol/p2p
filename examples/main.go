package main

import (
	"fmt"
	"strconv"

	//"github.com/symphonyprotocol/p2p/config"
	"github.com/symphonyprotocol/p2p"
	"github.com/symphonyprotocol/p2p/encrypt"
	"github.com/symphonyprotocol/p2p/kad"
	"github.com/symphonyprotocol/p2p/node"
	"github.com/symphonyprotocol/p2p/utils"

	//"github.com/symphonyprotocol/p2p/udp"
	//"math/big"
	"encoding/hex"
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

func distance(a, b []byte) int {
	c := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		c[i] = a[i] ^ b[i]
	}
	r := fmt.Sprintf("%v", c[0])
	x, _ := strconv.Atoi(r)
	return x
}

func testDistance() {
	node1 := "5be4506d26fe4c6e83e2fb644f0f5254679fcd112c271b1bd77b19e99f7e5482"
	node2 := "6695a9ee2972376eb9d3e7c6a4925aef6b1a4edfc5b9f496c79d11f02ca4901e"
	b1, _ := hex.DecodeString(node1)
	b2, _ := hex.DecodeString(node2)
	fmt.Print(distance(b1, b2))
}

func main() {
	fmt.Println("hello p2p")
	initialServer()
	//testJson()
	//testDistance()
}
