// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/connect_scanner.go
package modernui

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// performRealScan performs an actual network scan instead of using sample data
// This replaces the placeholder implementation in scanIPRange
func performRealScan(startIP, endIP string) []HostResult {
	var results []HostResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	fmt.Printf("DEBUG: Starting real network scan from %s to %s\n", startIP, endIP)

	// Convert IPs to integers for iteration
	startIPInt := convertIPToInt(net.ParseIP(startIP))
	endIPInt := convertIPToInt(net.ParseIP(endIP))

	// Rate limiting to avoid overwhelming the network
	semaphore := make(chan struct{}, 100)

	// Scan each IP in range
	for ipInt := startIPInt; ipInt <= endIPInt; ipInt++ {
		wg.Add(1)
		semaphore <- struct{}{} // acquire token

		go func(ip uint32) {
			defer wg.Done()
			defer func() { <-semaphore }() // release token

			ipStr := convertIntToIP(ip).String()
			result := checkHostActive(ipStr, timeoutDuration)

			if result.Status == "Active" {
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				fmt.Printf("DEBUG: Found active host: %s (%s)\n", result.IPAddress, result.Hostname)
			}
		}(ipInt)
	}

	wg.Wait()
	fmt.Printf("DEBUG: Real scan complete. Found %d active hosts.\n", len(results))
	return results
}

// checkHostActive checks if a host is active and returns a HostResult
func checkHostActive(ip string, timeout time.Duration) HostResult {
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

	// If ping didn't work, try more ports to determine if host is active
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
			result.Hostname = "Unknown-" + strings.Replace(ip, ".", "-", -1)
		}
	}

	return result
}

// convertIPToInt converts an IP address to an integer
func convertIPToInt(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// convertIntToIP converts an integer to an IP address
func convertIntToIP(i uint32) net.IP {
	return net.IPv4(byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
}
