package interfaces

import (
	"net"

	"github.com/symphonyprotocol/p2p/models"
)

type INetwork interface {
	RemoveCallback(category string)
	RegisterCallback(category string, callback func(models.CallbackParams))
	Send(ip net.IP, port int, bytes []byte)
}
