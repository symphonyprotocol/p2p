package tcp

import (
	"github.com/symphonyprotocol/p2p/node"
	"crypto/tls"
	"fmt"
	"github.com/symphonyprotocol/log"
	"net"
	"crypto/ecdsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
	"crypto/rand"
	"encoding/pem"
)

var sTcpLogger = log.GetLogger("SecuredTcp")

type SecuredTCPService struct {
	*TCPService
}

type SecuredTCPDialer struct {
	TlsConfig	*tls.Config
}

func NewSecuredTCPService(n *node.LocalNode) *SecuredTCPService {
	tcpService := &TCPService{
		localNodeId:	n.GetID(), 
		ip: 			n.GetLocalIP(), 
		port:			n.GetLocalPort(),
		tcpDialer:		&SecuredTCPDialer{ },
	}

	service := &SecuredTCPService{ TCPService: tcpService }
	privKey := n.GetPrivateKey()
	cer, err := tls.X509KeyPair([]byte(service.genCert(privKey)), []byte(service.genPriv(privKey)))
	if err != nil {
		sTcpLogger.Fatal("%v", err)
	}
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cer}}

	listener, err := tls.Listen("tcp", fmt.Sprintf("%v:%v", n.GetLocalIP().String(), n.GetLocalPort()), tlsCfg)
	if err != nil {
		panic(err)
	}

	service.listener = listener

	return service
}

func (tcp *SecuredTCPService) genCert(pri *ecdsa.PrivateKey) string {
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"SymphonyProtocol"},
		},
		NotBefore: time.Now().Add(-time.Hour * 24 * 365),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365),
	}

	certDer, err := x509.CreateCertificate(
		rand.Reader, &template, &template, pri.Public(), pri,
	)
	
	if err != nil {
		sTcpLogger.Fatal("Failed to create certificate: %s", err)
	}
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDer,
	}

	certEncoded := pem.EncodeToMemory(certBlock)
	return string(certEncoded)
}

func (tcp *SecuredTCPService) genPriv(pri *ecdsa.PrivateKey) string {
    x509Encoded, _ := x509.MarshalECPrivateKey(pri)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})
	return string(pemEncoded)
}

func (tcp *SecuredTCPDialer) DialRemoteServer(ip net.IP, port int) (net.Conn, error) {
	conn, err := tls.DialWithDialer(&net.Dialer{ KeepAlive: time.Minute }, "tcp", fmt.Sprintf("%v:%v", ip.String(), port), &tls.Config{ InsecureSkipVerify: true })
	if err != nil {
		sTcpLogger.Error("Failed to open secured tcp connection to %v:%v, error: %v", ip.String(), port, err)
		return nil, err
	}

	return conn, nil
}



