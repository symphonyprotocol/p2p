package p2p

import (
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/p2p/tcp"
	ui "github.com/strawhatboy/termui"
	"time"
	"fmt"
	"strconv"
	"sort"
	"strings"
)

var dmLogger = log.GetLogger("dashboard")

type DashboardMiddleware struct {

}


func (d *DashboardMiddleware) Handle(ctx *tcp.P2PContext) {
	ctx.Next()
}
func (d *DashboardMiddleware) Start(ctx *tcp.P2PContext) {
	startTime := time.Now()
	err := ui.Init()
	if err != nil {
		panic(err)
	}

	defer ui.Close()

	localNode := ctx.LocalNode()
	ls := ui.NewTable()
	ls.Border = true
	ls.BorderLabel = "Local Node"
	ls.Separator = false

	tUdpPeers := ui.NewTable()
	tUdpPeers.Border = true
	tUdpPeers.BorderLabel = "UDP Peers"
	tUdpPeers.Separator = false

	tTcpConns := ui.NewTable()
	tTcpConns.Border = true
	tTcpConns.BorderLabel = "TCP Connections"
	tTcpConns.Separator = false

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, ls),
		),
		ui.NewRow(
			ui.NewCol(12, 0, tUdpPeers),
		),
		ui.NewRow(
			ui.NewCol(12, 0, tTcpConns),
		),
	)

	ticker := time.NewTicker(time.Second)
	go func() {
		for range ticker.C {
			uptime := d.upTime(startTime)
			udpPeers := ctx.NodeProvider().PeekNodes()
			tcpService, _ := ctx.Network().(*tcp.SecuredTCPService)
			tcpConns := tcpService.GetTCPConnections()

			dmLogger.Debug("Got peers: %v and conns: %v", len(udpPeers), len(tcpConns))

			ls.Rows = [][]string{
				[]string { "Id:", localNode.GetID() },
				[]string { "PubKey:", localNode.GetPublicKey() },
				[]string { "Local Address:", fmt.Sprintf("%v:%v", localNode.GetLocalIP().String(), localNode.GetLocalPort()) },
				[]string { "Remote Address:", fmt.Sprintf("%v:%v", localNode.GetRemoteIP().String(), localNode.GetRemotePort()) },
				[]string { "Up time:", fmt.Sprintf("%v", uptime) },
				[]string { "Block Height:", fmt.Sprintf("%v", BlockHeight) },
			}

			ls.Height = len(ls.Rows) + 2
			ls.Analysis()
			ls.SetSize()
			//ls.Align()

			tUdpPeers.Rows = [][]string{
			}

			for _, peer := range udpPeers {
				tUdpPeers.Rows = append(tUdpPeers.Rows, []string{
					" ", 
					peer.GetID(), 
					fmt.Sprintf("%v:%v", peer.GetRemoteIP().String(), peer.GetRemotePort()), 
					strconv.Itoa(peer.Latency),
					fmt.Sprintf("%v", peer.LastActiveTime),
				})
			}
			dmLogger.Debug("table peers rows: %v", len(tUdpPeers.Rows))

			tUdpPeers.Height = len(udpPeers) + 3
			sort.Slice(tUdpPeers.Rows[:], func(i, j int) bool {
				return strings.Compare(tUdpPeers.Rows[i][1], tUdpPeers.Rows[j][1]) < 0
			})

			tUdpPeers.Rows = append([][]string{[]string{ "", "Id", "RemoteAddr", "Latency(ms)", "LastActiveTime"}}, tUdpPeers.Rows...)
			tUdpPeers.Analysis()
			tUdpPeers.SetSize()

			for n, peer := range tUdpPeers.Rows {
				if peer[3] == "-1" {
					tUdpPeers.BgColors[n] = ui.ColorRed
				} else if n == 0 {
					tUdpPeers.BgColors[n] = ui.ColorWhite
					tUdpPeers.FgColors[n] = ui.ColorBlack
				} else {
					tUdpPeers.BgColors[n] = ui.ColorDefault
				}
			}
			//tUdpPeers.Align()

			tTcpConns.Rows = [][]string{
			}

			for _, tConn := range tcpConns {
				tTcpConns.Rows = append(tTcpConns.Rows, []string{
					" ",
					tConn.LocalAddr().String(),
					tConn.RemoteAddr().String(),
					strconv.FormatBool(tConn.GetIsInBound()),
					tConn.GetNodeID(),
					fmt.Sprintf("%v", tConn.GetLastActiveTime()),
				})
			}
			dmLogger.Debug("table conns rows: %v", len(tTcpConns.Rows))

			sort.Slice(tTcpConns.Rows[:], func(i, j int) bool {
				return strings.Compare(tTcpConns.Rows[i][4], tTcpConns.Rows[j][4]) < 0
			})

			tTcpConns.Rows = append([][]string{[]string{ "", "LocalAddr", "RemoteAddr", "IsInbound", "NodeId", "LastActiveTime" }}, tTcpConns.Rows...)

			tTcpConns.Height = len(tcpConns) + 3
			tTcpConns.Analysis()
			tTcpConns.SetSize()

			tTcpConns.BgColors[0] = ui.ColorWhite
			tTcpConns.FgColors[0] = ui.ColorBlack

			ui.Body.Align()
			ui.Render(ui.Body)
		}
	}()

	ui.Body.Align()
	
	ui.Handle("<Resize>", func(e ui.Event) {
		payload := e.Payload.(ui.Resize)
		ui.Body.Width = payload.Width
		ui.Body.Align()
		ui.Clear()
		ui.Render(ui.Body)
	})
	
	ui.Handle("q", func(ui.Event) {
		ui.StopLoop()
		ticker.Stop()
	})

	ui.Loop()
}
func (d *DashboardMiddleware) AcceptConnection(*tcp.TCPConnection) {

}
func (d *DashboardMiddleware) DropConnection(*tcp.TCPConnection) {

}

func (d *DashboardMiddleware) upTime(t time.Time) time.Duration {
	return time.Since(t)
}

