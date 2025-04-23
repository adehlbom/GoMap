// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/host_ping.go
package modernui

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

// pingHost checks if a host is active and returns information about it
func pingHost(ip string, timeout time.Duration) HostResult {
	result := HostResult{
		IPAddress: ip,
		Status:    "Inactive",
		OpenPorts: 0,
	}

	// First try ICMP ping (doesn't require open ports)
	cmd := exec.Command("ping", "-c", "1", "-W", "1", ip)
	err := cmd.Run()
	if err == nil {
		result.Status = "Active"
	}

	// If ping didn't work, try common ports to determine if host is active
	if result.Status != "Active" {
		// Extended list of commonly open ports on various devices
		commonPorts := []int{80, 443, 22, 3389, 21, 23, 25, 53, 8080, 8443, 445, 139, 135, 5900, 5901, 62078, 7000}
		for _, port := range commonPorts {
			address := fmt.Sprintf("%s:%d", ip, port)
			conn, err := net.DialTimeout("tcp", address, timeout/2) // Reduced timeout for faster scanning
			if err == nil {
				conn.Close()
				result.Status = "Active"
				result.OpenPorts++
				break // Found an open port, no need to check more
			}
		}
	}

	// Try to resolve hostname if host is active
	if result.Status == "Active" {
		// Try DNS lookup
		names, err := net.LookupAddr(ip)
		if err == nil && len(names) > 0 {
			result.Hostname = strings.TrimSuffix(names[0], ".")
		} else {
			// If DNS lookup fails, use a placeholder hostname
			result.Hostname = "Unknown-" + strings.Replace(ip, ".", "-", -1)
		}
	}

	return result
}

// ipToInt converts an IP address to an integer
func ipToInt(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}
