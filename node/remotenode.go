package node

import (
	//"crypto/ecdsa"

	"net"

	symen "github.com/symphonyprotocol/p2p/encrypt"
)

type RemoteNode struct {
	Node
	Distance int
}

func (r *RemoteNode) RefreshNode(localIP string, localPort int, remoteIP string, remotePort int) {
	r.localIP = net.ParseIP(localIP)
	r.localPort = localPort
	r.remoteIP = net.ParseIP(remoteIP)
	r.remotePort = remotePort
}

func (r *RemoteNode) SetPublicKey(keyStr string) {
	r.pubKey = symen.ToPublicKey(keyStr)
}

func (r *RemoteNode) GetSendIPWithPort(local *LocalNode) (net.IP, int) {
	//remote and local are behind same NAT
	if r.remoteIP.String() == local.remoteIP.String() {
		return r.localIP, r.localPort
	}
	return r.remoteIP, r.remotePort
}

func NewRemoteNode(id []byte, localIP net.IP, localPort int, remoteIP net.IP, remotePort int) *RemoteNode {
	remote := &RemoteNode{}
	remote.Node.id = id
	remote.Node.localIP = localIP
	remote.Node.localPort = localPort
	remote.Node.remoteIP = remoteIP
	remote.Node.remotePort = remotePort
	remote.Distance = -1
	return remote
}
