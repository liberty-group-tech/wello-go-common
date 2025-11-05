package helper

import (
	"net"
	"strings"
)

func GetLocalIP() string {

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil && strings.HasPrefix(ipnet.IP.String(), "10.") {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
