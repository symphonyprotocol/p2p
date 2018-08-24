package node

import (
	//"crypto/ecdsa"
	"encoding/hex"
	"net"

	symen "github.com/symphonyprotocol/p2p/encrypt"
)

type RemoteNode struct {
	Node
	Distance int
}

func (r *RemoteNode) GetID() string {
	return hex.EncodeToString(r.id)
}

func (r *RemoteNode) GetPublicKey() string {
	return symen.FromPublicKey(r.pubKey)
}

func (r *RemoteNode) GetIP() net.IP {
	return r.ip
}

func (r *RemoteNode) GetPort() int {
	return r.port
}

func (r *RemoteNode) GetIDBytes() []byte {
	return r.id
}

func (r *RemoteNode) RefreshNode(ip string, port int) {
	r.ip = net.ParseIP(ip)
	r.port = port
}

func (r *RemoteNode) SetPublicKey(keyStr string) {
	r.pubKey = symen.ToPublicKey(keyStr)
}

func NewRemoteNode(id []byte, ip net.IP, port int) *RemoteNode {
	remote := &RemoteNode{}
	remote.Node.id = id
	remote.Node.port = port
	remote.Node.ip = ip
	remote.Distance = -1
	return remote
}
