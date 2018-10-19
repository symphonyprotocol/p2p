package config

import (
	"encoding/json"
)

var (
	staticNodeList = []byte(`
	{
		"nodes":
		[
			{
				"id": "04b4ff3f54f3c575631cd30763d9bd4c6785658bbd7fa3b0f241a52ecad271a4",
				"ip":"192.168.1.102",
				"port": 32768
			}
		]
	}`)

// var (
// 	staticNodeList = []byte(`
// 	{
// 		"nodes":
// 		[
// 			{
// 				"id": "e9fa8677cdff28ccc9a0f27d74b032e62deba74c5adc05b394a90182e596726d",
// 				"ip":"10.106.53.150",
// 				"port": 32768
// 			}
// 		]
// 	}`)

	DEFAULT_UDP_PORT = 32768
	DEFAULT_TCP_PORT = 32768
	DEFAULT_NET_WORK = "MINOR"
)

type StaticNodes struct {
	Nodes []StaticNode `json:"nodes"`
}

type StaticNode struct {
	ID        string `json:"id"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	PublicKey string `json:"publickey"`
}

func LoadStaticNodes() StaticNodes {
	var nodes StaticNodes
	err := json.Unmarshal(staticNodeList, &nodes)
	if err != nil {
		panic(err)
	}
	return nodes
}
