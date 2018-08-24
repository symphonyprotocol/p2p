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
				"id": "047de44480ff166ee6c8a325f8d2a1bbb12d056159cd4f871c92f8464efe9f564baad09334aba0f83f7b805400ed85af7b54e59c3cce111dfbce5e36928e73c9f3",
				"ip":"47.88.227.223",
				"port": 32768
			}
		]
	}`)

	DEFAULT_UDP_PORT = 32768
	DEFAULT_TCP_PORT = 32768
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
