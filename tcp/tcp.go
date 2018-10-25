package tcp

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/p2p/node"

	"github.com/symphonyprotocol/p2p/models"
	"github.com/symphonyprotocol/p2p/utils"
	"time"
)

var tcpLogger = log.GetLogger("tcp")

type TCPConnection struct {
	net.Conn
	stop      chan struct{}
	isInbound bool
	nodeId    string // to be filled when confirmed.
	lastActiveTime	time.Time
}

func (t TCPConnection) GetIsInBound() bool { return t.isInbound }
func (t TCPConnection) GetNodeID() string { return t.nodeId }
func (t TCPConnection) GetLastActiveTime() time.Time { return t.lastActiveTime }

type TCPCallbackParams struct {
	models.CallbackParams
	Connection *TCPConnection
}

func (u TCPCallbackParams) GetTCPRemoteAddr() *net.TCPAddr {
	if addr, ok := u.RemoteAddr.(*net.TCPAddr); ok {
		return addr
	}

	return nil
}

func (u TCPCallbackParams) GetTCPDiagram() models.TCPDiagram {
	if diag, ok := u.Diagram.(models.TCPDiagram); ok {
		return diag
	}

	// error

	return models.TCPDiagram{}
}

func NewTCPConnection(conn net.Conn, isInbound bool) *TCPConnection {
	return &TCPConnection{
		Conn:      conn,
		isInbound: isInbound,
		stop:      make(chan struct{}),
		lastActiveTime:	time.Now(),
	}
}

type ITCPDialer interface {
	DialRemoteServer(ip net.IP, port int) (net.Conn, error)
}

type TCPDialer struct {
}

type TCPService struct {
	models.INetwork
	listener    net.Listener
	connections sync.Map // map[string] *net.TCPConn	// string(ip.To16())	net.IP(ipStr)

	tcpDialer ITCPDialer

	localNodeId string
	ip          net.IP
	port        int

	callbacks sync.Map
	newConnectionHander	func(*TCPConnection)
	connectionDroppedHandler	func(*TCPConnection)
}

func NewTCPService(localNode *node.LocalNode) *TCPService {
	service := &TCPService{
		localNodeId: localNode.GetID(),
		ip:          localNode.GetLocalIP(),
		port:        localNode.GetLocalPort(),
		tcpDialer:   &TCPDialer{},
	}

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: service.ip, Port: service.port})
	if err != nil {
		panic(err)
	}

	service.listener = listener

	return service
}

func (tcp *TCPService) getConnectionKey(ip net.IP, port int) string {
	return fmt.Sprintf("%v:%v", ip.String(), port)
}

func (tcp *TCPService) loop() {
	tcpLogger.Trace("Start listening TCP connections...")
	for {
		conn, err := tcp.listener.Accept()
		if err != nil {
			panic(err)
		}
		remoteAddr := conn.RemoteAddr()
		tcpAddr, err := net.ResolveTCPAddr(remoteAddr.Network(), remoteAddr.String())
		if err != nil {
			tcpLogger.Error("Failed to parse remote addr from the new tcp connection")
		}
		the_key := tcp.getConnectionKey(tcpAddr.IP, tcpAddr.Port)
		// 1. check if connection in map
		if _conn, ok := tcp.connections.Load(the_key); ok {
			if _, ok := _conn.(*TCPConnection); ok {
				// conn already exists.. opened by us
				conn.Close()
				continue
			}
		}
		// 2. accept this connection
		tcpLogger.Trace("Accepting incoming connection with key: %v", the_key)
		the_conn := NewTCPConnection(conn, true)
		tcp.connections.Store(the_key, the_conn)
		if tcp.newConnectionHander != nil {
			tcp.newConnectionHander(the_conn)
		}
		go tcp.handleConnection(the_conn, the_key)
	}
}

func (tcp *TCPService) handleConnection(conn *TCPConnection, key string) {
	for {
		// check if we need to close this conn
		quit := false
		select {
		case <-conn.stop:
			quit = true
		default:
			// keep reading from conn
			data := make([]byte, 1280)

			// important, client side may fail to recieve the handshake response.
			// it would read here forever.
			conn.SetReadDeadline(time.Now().Add(time.Minute * 2))
			n, err := conn.Read(data)
			if err != nil {
				tcpLogger.Error("conn: read: %s", err)
				quit = true
			} else {

				tcpLogger.Trace("conn: received: %v bytes", n)
				remoteAddr := conn.RemoteAddr()
				tcpAddr, _ := net.ResolveTCPAddr(remoteAddr.Network(), remoteAddr.String())
				rdata := data[:n]
				var diagram models.TCPDiagram
				utils.BytesToUDPDiagram(rdata, &diagram)

				// update nodeID for the connection.
				conn.nodeId = diagram.NodeID
				conn.lastActiveTime = time.Now()
				if obj, ok := tcp.callbacks.Load(diagram.DCategory); ok {
					callback := obj.(func(models.ICallbackParams))
					callback(TCPCallbackParams{
						CallbackParams: models.CallbackParams{
							RemoteAddr: tcpAddr,
							Diagram:    diagram,
							Data:       rdata,
						},
						Connection: conn,
					})
				}
			}
		}

		if quit {
			tcpLogger.Trace("TCP Connection to %v quit by signal", conn.RemoteAddr().String())
			// 2. close this connection
			conn.Close()
			if tcp.connectionDroppedHandler != nil {
				tcp.connectionDroppedHandler(conn)
			}
			// 3. remove from map
			tcp.connections.Delete(key)
			break
		}
	}
}

func (tcp *TCPService) getConnection(ip net.IP, port int, nodeId string) (*TCPConnection, error) {
	the_key := tcp.getConnectionKey(ip, port)
	// 1. check if connection in map
	if _conn, ok := tcp.connections.Load(the_key); ok {
		if conn, ok := _conn.(*TCPConnection); ok {
			tcpLogger.Trace("connection %v is in the map, isInbound: %v", the_key, conn.isInbound)
			if sConn, ok := conn.Conn.(*tls.Conn); ok {
				tcpLogger.Trace("this is a secured connection, tls version: 0x%x, cipher: 0x%x", sConn.ConnectionState().Version, sConn.ConnectionState().CipherSuite)
			}
			return conn, nil
		}
	}

	// 1.1 check if the connection can only be found by nodeId, this is probably an inbound connection.
	var the_conn *TCPConnection = nil

	tcp.connections.Range(func(k interface{}, v interface{}) bool {
		if conn, ok := v.(*TCPConnection); ok {
			if conn.nodeId == nodeId {
				// got this connection
				tcpLogger.Trace("connection %v is in the map, but found by its nodeId, real address is %v, isInbound: %v", the_key, conn.RemoteAddr().String(), conn.isInbound)
				the_conn = conn
				return false
			}
		} else {
			return false
		}

		return true
	})

	if the_conn != nil {
		return the_conn, nil
	}

	// 2. create new connection
	// localIP := &net.TCPAddr{ IP: tcp.ip, Port: tcp.port }
	conn, err := tcp.tcpDialer.DialRemoteServer(ip, port)
	if err != nil {
		return nil, err
	}

	// 3. add connection to map
	the_conn = NewTCPConnection(conn, false)
	tcp.connections.Store(the_key, the_conn)

	// 4. start connection listener
	go tcp.handleConnection(the_conn, the_key)

	return the_conn, nil
}

func (tcp *TCPService) closeConnection(ip net.IP, port int) {
	key := tcp.getConnectionKey(ip, port)
	if _conn, ok := tcp.connections.Load(key); ok {
		if conn, ok := _conn.(*TCPConnection); ok {
			// 1. stop the handle Inbound Connection loop for this connection
			conn.stop <- struct{}{}
			return
		}
	}

	tcpLogger.Warn("Failed to close connection %v", key)
}

func (c *TCPService) RegisterCallback(category string, callback func(models.ICallbackParams)) {
	c.callbacks.Store(category, callback)
}

func (c *TCPService) RemoveCallback(category string) {
	c.callbacks.Delete(category)
}

func (c *TCPService) Send(ip net.IP, port int, bytes []byte, nodeId string) {
	conn, err := c.getConnection(ip, port, nodeId)
	if err != nil {
		tcpLogger.Error("Failed to send packet (%d) to %v:%v", len(bytes), ip.String(), port)
		return
	}

	// TODO: chunksize
	// TODO: encryption (can be done by tls on tcp connection?)
	length, err := conn.Write(bytes)
	if err != nil {
		tcpLogger.Error("Failed to send packet (%d) to %v:%v", length, ip.String(), port)
	} else {
		tcpLogger.Trace("Packet (%d) sent to %v:%v", length, ip.String(), port)
	}
}

func (tcp *TCPService) Start() {
	go tcp.loop()
}

func (tcp *TCPDialer) DialRemoteServer(ip net.IP, port int) (net.Conn, error) {
	remoteIP := &net.TCPAddr{IP: ip, Port: port}
	conn, err := net.DialTCP("tcp", nil, remoteIP)
	if err != nil {
		tcpLogger.Error("Failed to open tcp connection to %v:%v, error: %v", ip.String(), port, err)
		return nil, err
	}

	return conn, nil
}

func (tcp *TCPService) RegisterAcceptConnectionEvent(f func (*TCPConnection)) {
	if f != nil {
		tcp.newConnectionHander = f
	}
}

func (tcp *TCPService) RegisterDropConnectionEvent(f func (*TCPConnection)) {
	if f != nil {
		tcp.connectionDroppedHandler = f
	}
}

func (tcp *TCPService) GetTCPConnections() []*TCPConnection {
	res := make([]*TCPConnection, 0, 0)
	tcpLogger.Debug("Getting TCP Connections to public")
	if tcp != nil {
		tcp.connections.Range(func(k interface{}, v interface{}) bool {
			if conn, ok := v.(*TCPConnection); ok {
				// tcpLogger.Debug("appending tcp connections")
				res = append(res, conn)
			}
			return true
		})
	} else {
		tcpLogger.Warn("No tcp instance???")
	}

	return res
}
