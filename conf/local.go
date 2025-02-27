package conf

import (
	"net"
	"strconv"
)

type Local struct {
	PublicHttpServerPort int
}

func (l Local) PublicServerAddress() string {
	return net.JoinHostPort("0.0.0.0", strconv.Itoa(l.PublicHttpServerPort))
}
