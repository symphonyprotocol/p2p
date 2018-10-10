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

type ILocalNode interface {
	Interface
	GetPrivateKey() *ecdsa.PrivateKey
}

type Node struct {
	id         []byte
	localIP    net.IP
	localPort  int
	remoteIP   net.IP
	remotePort int
	pubKey     ecdsa.PublicKey
	network    string
}

func (n *Node) GetID() string {
	return hex.EncodeToString(n.id)
}

func (n *Node) GetIDBytes() []byte {
	return n.id
}

func (n *Node) GetPublicKey() string {
	return symen.FromPublicKey(n.pubKey)
}

func (n *Node) GetLocalIP() net.IP {
	return n.localIP
}

func (n *Node) GetLocalPort() int {
	return n.localPort
}

func (n *Node) GetRemoteIP() net.IP {
	return n.remoteIP
}

func (n *Node) GetRemotePort() int {
	return n.remotePort
}

func (n *Node) GetNetwork() string {
	return n.GetNetwork()
}

type LocalNode struct {
	Node
	privKey  *ecdsa.PrivateKey
	isPublic bool
}

func (n *LocalNode) SetRemoteIPPort(ip string, port int) {
	n.remoteIP = net.ParseIP(ip)
	n.remotePort = port
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
	localNode.Node.network = config.DEFAULT_NET_WORK
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

	ip := net.ParseIP(ipStr)
	localNode.Node.localIP = ip
	localNode.Node.localPort = config.DEFAULT_UDP_PORT
	log.Printf("setup local node ip: %v:%v\n", localNode.localIP, localNode.localPort)
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
		mappingPort := 0
		index := 0
		dictPorts := make(map[int]int)
		for {
			_, ip, extPort, _ := upnp.GetGenericPortMappingEntry(index, client)
			if extPort == 0 {
				break
			}
			if ip == n.localIP.String() {
				mappingPort = extPort
			}
			dictPorts[extPort] = 1
			index++
		}
		if mappingPort == 0 {
			mappingPort = config.DEFAULT_UDP_PORT
			for {
				if _, ok := dictPorts[mappingPort]; ok {
					mappingPort++
				} else {
					break
				}
			}
			if ok := upnp.AddPortMapping(n.localIP.String(), config.DEFAULT_UDP_PORT, mappingPort, "UDP", client); ok {
				log.Printf("add port mapping for UDP from %v to %v\n", config.DEFAULT_UDP_PORT, mappingPort)
			}
			if ok := upnp.AddPortMapping(n.localIP.String(), config.DEFAULT_TCP_PORT, mappingPort, "TCP", client); ok {
				log.Printf("add port mapping for TCP from %v to %v\n", config.DEFAULT_TCP_PORT, mappingPort)
			}
		}
		// get external ip
		externalIP, err := upnp.GetExternalIPAddress(client)
		if err != nil {
			log.Printf("upnp get external ip address error:%v", err)
			n.isPublic = false
		} else {
			n.isPublic = !nat.IsIntranet(externalIP)
			if n.isPublic {
				extIP := net.ParseIP(externalIP)
				n.remoteIP = extIP
				n.remotePort = mappingPort
			}
		}
	}
}

func (ln *LocalNode) GetPrivateKey() *ecdsa.PrivateKey {
	return ln.privKey
}
