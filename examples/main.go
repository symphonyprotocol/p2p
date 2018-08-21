package main

import (
	"fmt"
	//"github.com/symphonyprotocol/p2p/config"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"github.com/symphonyprotocol/p2p"
	"github.com/symphonyprotocol/p2p/encrypt"
	"github.com/symphonyprotocol/p2p/kad"
	"github.com/symphonyprotocol/p2p/node"
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
	ktable := kad.NewKTable(getId())
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
	data, _ := json.Marshal(diagram)
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	gz.Write([]byte(string(data)))
	gz.Flush()
	gz.Close()
	fmt.Println(len(b.Bytes()), b.Bytes())
	fmt.Println(len(data), data)
	str := base64.StdEncoding.EncodeToString(b.Bytes())
	fmt.Println(str)
	fmt.Println(string(data))
	bd := BaseDiagram{}
	err := json.Unmarshal(data, &bd)
	fmt.Println(err)
	fmt.Println(bd)
}

func main() {
	fmt.Println("hello p2p")
	initialServer()
}
