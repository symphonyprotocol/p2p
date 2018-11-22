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
				"id": "80f7f157bc113cbb0ef789ca9059657ad4a394da890e0edfd3dd55665e28a714",
				"ip":"192.168.0.110",
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
