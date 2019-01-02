package config

import (
	"encoding/json"
	"os/user"
	"io/ioutil"
)

var (
	staticNodeList = []byte(`
	{
		"nodes":
		[
			{
				"id": "c4ef0694fee0cdf78eab30c83b325293047e0b27511b92e8e206b199b24f13ea",
				"ip":"101.200.156.243",
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
	
	CURRENT_USER, _ = user.Current()
	LEVEL_DB_FILE = CURRENT_USER.HomeDir + "/.symchaindb"
	CONFIG_FILE = CURRENT_USER.HomeDir + "/.symchaincfg"
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
	nodeList, err := ioutil.ReadFile(CONFIG_FILE)
	if err != nil {
		err = json.Unmarshal(staticNodeList, &nodes)
	} else {
		err = json.Unmarshal(nodeList, &nodes)
	}
	if err != nil {
		panic(err)
	}
	return nodes
}
