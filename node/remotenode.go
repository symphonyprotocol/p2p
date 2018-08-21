package node

import (
	//"crypto/ecdsa"
	"encoding/hex"
	symen "github.com/symphonyprotocol/p2p/encrypt"
	"net"
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

func (r *RemoteNode) GetUdpPort() int {
	return r.uport
}

func (r *RemoteNode) GetTcpPort() int {
	return r.tport
}

func (r *RemoteNode) GetIDBytes() []byte {
	return r.id
}

func (r *RemoteNode) SetPublicKey(keyStr string) {
	r.pubKey = symen.ToPublicKey(keyStr)
}

func NewRemoteNode(id []byte, ip net.IP, uport int, tport int) *RemoteNode {
	remote := &RemoteNode{}
	remote.Node.id = id
	remote.Node.uport = uport
	remote.Node.tport = tport
	remote.Node.ip = ip
	remote.Distance = -1
	return remote
}
