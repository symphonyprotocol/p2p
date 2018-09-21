package tcp

import (
	"fmt"
	"sync"
	"net"
	"github.com/symphonyprotocol/log"

	"github.com/symphonyprotocol/p2p/models"
	"github.com/symphonyprotocol/p2p/interfaces"
	"github.com/symphonyprotocol/p2p/utils"
)

var logger = log.GetLogger("tcp")

type TCPConnection struct {
	*net.TCPConn
	stop			chan struct{} 
	isInbound		bool
}

func NewTCPConnection(conn *net.TCPConn, isInbound bool) *TCPConnection {
	return &TCPConnection {
		TCPConn		:	conn,
		isInbound	:	isInbound,
		stop		:	make(chan struct{}),
	}
}

type TCPService struct {
	interfaces.INetwork
	listener		*net.TCPListener
	connections		sync.Map				// map[string] *net.TCPConn	// string(ip.To16())	net.IP(ipStr)

	localNodeId		string
	ip 				net.IP
	port			int

	callbacks		sync.Map
}

func NewTCPService(localNodeId string, ip net.IP, port int) *TCPService {
	service := &TCPService {
		localNodeId	:	localNodeId,
		ip			:	ip,
		port		:	port,
	}

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{ IP: ip, Port: port })
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
	logger.Trace("Start listening TCP connections...")
	for {
		conn, err := tcp.listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		remoteAddr := conn.RemoteAddr()
		tcpAddr, err := net.ResolveTCPAddr(remoteAddr.Network(), remoteAddr.String())
		if err != nil {
			logger.Trace("Failed to parse remote addr from the new tcp connection\n")
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
		the_conn := NewTCPConnection(conn, true)
		tcp.connections.Store(the_key, the_conn)
		go tcp.handleConnection(the_conn)
	}
}

func (tcp *TCPService) handleConnection(conn *TCPConnection) {
	for {
		// check if we need to close this conn
		quit := false
		select {
		case <- conn.stop:
			quit = true
		default:
			// keep reading from conn
			data := make([]byte, 1280)
			n, err := conn.Read(data)
			if err != nil {
				if err != nil {
					logger.Trace("conn: read: %s", err)
				}
				break
			}

			remoteAddr := conn.RemoteAddr()
			tcpAddr, err := net.ResolveTCPAddr(remoteAddr.Network(), remoteAddr.String())
			rdata := data[:n]
			var diagram models.TCPDiagram
			utils.BytesToUDPDiagram(rdata, &diagram)
			if obj, ok := tcp.callbacks.Load(diagram.DCategory); ok {
				callback := obj.(func(models.TCPCallbackParams))
				callback(models.TCPCallbackParams{
					RemoteAddr: tcpAddr,
					Diagram:    diagram,
					Data:       rdata,
				})
			}
		}

		if quit {
			logger.Trace("TCP Connection to %v quit by signal\n", conn.RemoteAddr().String())
			break
		}
	}
}

func (tcp *TCPService) getConnection(ip net.IP, port int) (*TCPConnection, error) {
	the_key := tcp.getConnectionKey(ip, port)
	// 1. check if connection in map
	if _conn, ok := tcp.connections.Load(the_key); ok {
		if conn, ok := _conn.(*TCPConnection); ok {
			return conn, nil
		}
	} 

	// 2. create new connection
	localIP := &net.TCPAddr{ IP: tcp.ip, Port: tcp.port }
	remoteIP := &net.TCPAddr{ IP: ip, Port: port }
	conn, err := net.DialTCP("tcp", localIP, remoteIP)
	if err != nil {
		logger.Trace("Failed to open tcp connection to %v:%v\n", ip.String(), port)
		return nil, err
	}

	// 3. add connection to map
	the_conn := NewTCPConnection(conn, false)
	tcp.connections.Store(the_key, the_conn)

	// 4. start connection listener
	go tcp.handleConnection(the_conn)

	return the_conn, nil
}

func (tcp *TCPService) closeConnection(ip net.IP, port int) {
	key := tcp.getConnectionKey(ip, port)
	if _conn, ok := tcp.connections.Load(key); ok {
		if conn, ok := _conn.(*TCPConnection); ok {
			// 1. stop the handle Inbound Connection loop for this connection
			conn.stop <- struct{}{}
			// 2. close this connection
			conn.Close()
			// 3. remove from map
			tcp.connections.Delete(key)
			return
		}
	}

	logger.Trace("Failed to close connection %v", key)
}

func (c *TCPService) RegisterCallback(category string, callback func(models.CallbackParams)) {
	c.callbacks.Store(category, callback)
}

func (c *TCPService) RemoveCallback(category string) {
	c.callbacks.Delete(category)
}

func (c *TCPService) Send(ip net.IP, port int, bytes []byte) {
	conn, err := c.getConnection(ip, port)
	if err != nil {
		logger.Trace("Failed to send packet (%d) to %v:%v\n", len(bytes), ip.String(), port)
		return
	}

	// TODO: chunksize
	// TODO: encryption (can be done by tls on tcp connection?)
	length, err := conn.Write(bytes)
	if err != nil {
		logger.Trace("Failed to send packet (%d) to %v:%v\n", length, ip.String(), port)
	} else {
		logger.Trace("Packet (%d) sent to %v:%v\n", length, ip.String(), port)
	}
}

func (tcp *TCPService) Start() {
	go tcp.loop()
}
