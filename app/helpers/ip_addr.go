package helpers

import (
	"net"
	"net/http"
	"strings"
)

// IPAddr returns a best guess of the user's IP address.
func IPAddr(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		i := strings.IndexAny(ip, ", ")
		if i > 0 {
			return ip[:i]
		}
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
