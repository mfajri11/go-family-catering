package web

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

func RealIP(r *http.Request) string {
	var (
		err   error
		netIP net.IP
		ip    string
	)

	xRealIP := r.Header.Get("X-Real-IP")
	xForwardedIP := r.Header.Get("X-Forwarded-For")

	// trying to get IP from RemoteAddr
	if xRealIP == "" && xForwardedIP == "" {
		ip, _, err = net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			return ""
		}
		netIP = net.ParseIP(ip)
		if netIP != nil {
			return netIP.String()
		}
	}
	// trying to get IP from X-Forwarded-For header
	ips := strings.Split(xForwardedIP, ",")
	fmt.Println("ips: ", ips)
	for _, ip = range ips {
		netIP = net.ParseIP(ip)
		if netIP != nil {
			return netIP.String()
		}
	}

	return xRealIP
}
