package helpers

import (
	"encoding/binary"
	"fmt"
	"net"
)

func Ipv4ToInt(ip net.IP) (uint32, error) {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return 0, fmt.Errorf("not an IPv4")
	}
	return binary.BigEndian.Uint32(ipv4), nil
}