package util

import (
	"fmt"
	"strings"
)

// Get the shortened unit and the base from a byte value
func ByteDecodeUnit(num uint64) (string, uint64) {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	base := uint64(1)
	var unit string

	for i, u := range units {
		if num < base*1000 || i == len(units)-1 {
			unit = u
			break
		}
		base *= 1000
	}

	return unit, base
}

// Get the address (host:port) from a host and a port
func GetAddrFromHostPort(host, port string) string {
	return fmt.Sprintf("%s:%s", host, port)
}

// Get the host and the port from an address (host:port)
func GetHostPortFromAddr(addr string) (string, string, error) {
	if len(strings.Split(addr, ":")) != 2 {
		return "", "", fmt.Errorf("%s: wrong address format, it should be host:port\n", addr)
	}

	host := strings.Split(addr, ":")[0]
	port := strings.Split(addr, ":")[1]

	return host, port, nil
}
