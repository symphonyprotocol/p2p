package node

import (
	"crypto/ecdsa"
	"encoding/hex"
	"log"
	"net"

	"github.com/symphonyprotocol/nat"
	"github.com/symphonyprotocol/nat/upnp"
	"github.com/symphonyprotocol/p2p/config"
	symen "github.com/symphonyprotocol/p2p/encrypt"
	"github.com/symphonyprotocol/p2p/node/store"
)

type Interface interface {
	GetIDBytes() []byte
	GetID() string
	GetPublicKey() string
	GetIP() net.IP
	GetPort() int
	RefreshNode(ip string, port int)
}

type Node struct {
	id     []byte
	ip     net.IP
	port   int
	pubKey ecdsa.PublicKey
}

type LocalNode struct {
	Node
	privKey            *ecdsa.PrivateKey
	extIP              *net.IP
	extUPort, extTPort int
	isPublic           bool
}

func (n *LocalNode) GetID() string {
	return hex.EncodeToString(n.id)
}

func (n *LocalNode) GetIDBytes() []byte {
	return n.id
}

func (n *LocalNode) GetPublicKey() string {
	return symen.FromPublicKey(n.pubKey)
}

func (n *LocalNode) GetIP() net.IP {
	return n.ip
}

func (n *LocalNode) GetPort() int {
	return n.port
}

func (n *LocalNode) RefreshNode(ip string, port int) {
	n.ip = net.ParseIP(ip)
	n.port = port
}

func NewLocalNode() *LocalNode {
	var privKey *ecdsa.PrivateKey
	privKeyStr, pubKeyStr := store.GetLocalNodeKeyStr()
	if len(privKeyStr) == 0 {
		log.Println("generate key for node")
		privKey = symen.GenerateNodeKey()
		privKeyStr = symen.FromPrivateKey(privKey)
		pubKeyStr = symen.FromPublicKey(privKey.PublicKey)
		store.SaveLocalNodeKey(privKeyStr, pubKeyStr)
	} else {
		log.Println("load key for node")
		pubKey := symen.ToPublicKey(pubKeyStr)
		privKey = symen.ToPrivateKey(privKeyStr, pubKey)
	}
	localNode := &LocalNode{}
	localNode.Node.id = symen.PublicKeyToNodeId(privKey.PublicKey)
	log.Printf("setup local node: %v", localNode.GetID())
	var ipStr string
	ipStr, err := nat.GetOutbountIP()
	if err != nil {
		ips, err := nat.IntranetIP()
		if err != nil || len(ips) == 0 {
			ipStr = "127.0.0.1"
		} else {
			ipStr = ips[0]
		}
	}
	log.Printf("setup local node ip: %v", ipStr)
	ip := net.ParseIP(ipStr)
	localNode.Node.ip = ip
	localNode.Node.port = config.DEFAULT_UDP_PORT
	localNode.pubKey = privKey.PublicKey
	log.Printf("setup local node pubkey: %v", pubKeyStr)
	localNode.privKey = privKey
	return localNode
}
func (n *LocalNode) DiscoverNAT() {
	client, err := upnp.NewUPnPClient()
	if err != nil {
		log.Printf("discover upnp error:%v", err)
		return
	}
	if ok := client.Discover(); ok {
		// add mapping ports
		mappingPort := config.DEFAULT_UDP_PORT
		index := 0
		dictPorts := make(map[int]int)
		for {
			protocol, ip, extPort, _ := upnp.GetGenericPortMappingEntry(index, client)
			if extPort == 0 {
				break
			}
			if ip == n.ip.String() {
				//log.Println("find ip %v %v", protocol, extPort)
				if protocol == "UDP" {
					n.extUPort = extPort
				}
				if protocol == "TCP" {
					n.extTPort = extPort
				}
			}
			dictPorts[extPort] = 1
			index++
		}
		if n.extUPort == 0 {
			for {
				if _, ok := dictPorts[mappingPort]; ok {
					mappingPort++
				} else {
					break
				}
			}
			if ok := upnp.AddPortMapping(n.ip.String(), config.DEFAULT_UDP_PORT, mappingPort, "UDP", client); ok {
				log.Printf("add port mapping for UDP from %v to %v\n", config.DEFAULT_UDP_PORT, mappingPort)
			}
			if ok := upnp.AddPortMapping(n.ip.String(), config.DEFAULT_TCP_PORT, mappingPort, "TCP", client); ok {
				log.Printf("add port mapping for TCP from %v to %v\n", config.DEFAULT_TCP_PORT, mappingPort)
			}
		}
		// get external ip
		externalIP, err := upnp.GetExternalIPAddress(client)
		if err != nil {
			log.Printf("upnp get external ip address error:%v", err)
			n.isPublic = false
		} else {
			n.isPublic = nat.IsIntranet(externalIP)
			if n.isPublic {
				extIp := net.ParseIP(externalIP)
				n.extIP = &extIp
				n.extUPort = mappingPort
				n.extTPort = mappingPort
			}
		}
	}
}
