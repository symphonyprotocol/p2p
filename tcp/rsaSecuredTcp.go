package tcp

import (
	"crypto/rsa"
	"crypto/rand"
	"github.com/symphonyprotocol/p2p/node"
)

type RSASecuredTCPService struct {
	*TCPService
	rsaKey	*rsa.PrivateKey
}

func NewRSASecuredTCPService(n *node.LocalNode) *RSASecuredTCPService {
	tcpService := &TCPService{
		localNodeId: n.GetID(),
		ip:          n.GetLocalIP(),
		port:        n.GetLocalPort(),
		tcpDialer:   &SecuredTCPDialer{},
	}

	rsaService := &RSASecuredTCPService{
		TCPService: tcpService,
		rsaKey:		genRSAKeys(),
	}

	return rsaService
}

func (r *RSASecuredTCPService) Start() {

}

// the keys are for temp usage
func genRSAKeys() *rsa.PrivateKey {
	reader := rand.Reader
	key, err := rsa.GenerateKey(reader, 1 << 11)
	if err != nil {
		panic(err)
	}

	return key
}



