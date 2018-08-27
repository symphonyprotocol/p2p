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
				"id": "6695a9ee2972376eb9d3e7c6a4925aef6b1a4edfc5b9f496c79d11f02ca4901e",
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
