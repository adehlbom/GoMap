package main

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// HostResult represents a scanned host with its status and information
type HostResult struct {
	IPAddress string
	Hostname  string
	Status    string
	OpenPorts int
}

// scanIPRange scans an IP range and returns active hosts
func scanIPRange(startIP, endIP string, timeout time.Duration) []HostResult {
	var results []HostResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Convert IPs to integers for iteration
	startIPInt := ipToInt(net.ParseIP(startIP))
	endIPInt := ipToInt(net.ParseIP(endIP))

	// Rate limiting to avoid overwhelming the network
	semaphore := make(chan struct{}, 100)

	// Scan each IP in range
	for ipInt := startIPInt; ipInt <= endIPInt; ipInt++ {
		wg.Add(1)
		semaphore <- struct{}{} // acquire token

		go func(ip uint32) {
			defer wg.Done()
			defer func() { <-semaphore }() // release token

			ipStr := intToIP(ip).String()
			result := pingHost(ipStr, timeout)

			if result.Status == "Active" {
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}
		}(ipInt)
	}

	wg.Wait()
	return results
}

// pingHost checks if a host is active
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
			// If DNS lookup fails, try to get hostname via NetBIOS (Windows) or mDNS
			hostnames := tryAlternativeHostnameMethods(ip)
			if hostnames != "" {
				result.Hostname = hostnames
			} else {
				result.Hostname = "Unknown-" + strings.Replace(ip, ".", "-", -1)
			}
		}
	}

	return result
}

// parseIPRange parses user input to determine the start and end IPs
func parseIPRange(ipRange string) (string, string, error) {
	// Check if it's a CIDR notation
	if strings.Contains(ipRange, "/") {
		return parseCIDR(ipRange)
	}

	// Check if it's a range with hyphen
	if strings.Contains(ipRange, "-") {
		parts := strings.Split(ipRange, "-")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
		}
	}

	// If it's just a single IP, use it as both start and end
	if net.ParseIP(ipRange) != nil {
		return ipRange, ipRange, nil
	}

	return "", "", fmt.Errorf("invalid IP range format")
}

// parseCIDR parses a CIDR notation and returns the first and last IPs
func parseCIDR(cidr string) (string, string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", err
	}

	// Get the first IP in the range (network address)
	firstIP := ip.Mask(ipnet.Mask)

	// Get the last IP in the range (broadcast address)
	mask := ipnet.Mask
	lastIP := make(net.IP, len(firstIP))
	for i := 0; i < len(mask); i++ {
		lastIP[i] = firstIP[i] | ^mask[i]
	}

	// Skip the network and broadcast addresses for IPv4
	if len(firstIP) == 4 {
		// Convert to integers, add 1 to first, subtract 1 from last
		firstIPInt := ipToInt(firstIP)
		lastIPInt := ipToInt(lastIP)

		// Only adjust if it's not a /31 or /32
		if lastIPInt-firstIPInt > 1 {
			firstIPInt++
			lastIPInt--
		}

		return intToIP(firstIPInt).String(), intToIP(lastIPInt).String(), nil
	}

	return firstIP.String(), lastIP.String(), nil
}

// ipToInt converts an IP address to an integer
func ipToInt(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// intToIP converts an integer to an IP address
func intToIP(i uint32) net.IP {
	return net.IPv4(byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
}

// generateIPRange generates a range of IPs from a base IP by varying the last octet
func generateIPRange(baseIP string, lastOctetStart, lastOctetEnd int) (string, string) {
	parts := strings.Split(baseIP, ".")
	if len(parts) != 4 {
		return "", ""
	}

	basePrefix := strings.Join(parts[:3], ".")
	startIP := fmt.Sprintf("%s.%d", basePrefix, lastOctetStart)
	endIP := fmt.Sprintf("%s.%d", basePrefix, lastOctetEnd)

	return startIP, endIP
}

// ipRangeToString creates a user-friendly string representation of an IP range
func ipRangeToString(startIP, endIP string) string {
	// Check if IPs are in the same subnet
	startParts := strings.Split(startIP, ".")
	endParts := strings.Split(endIP, ".")

	if len(startParts) == 4 && len(endParts) == 4 &&
		startParts[0] == endParts[0] &&
		startParts[1] == endParts[1] &&
		startParts[2] == endParts[2] {
		// Only the last octet differs
		return fmt.Sprintf("%s.%s-%s",
			strings.Join(startParts[:3], "."),
			startParts[3], endParts[3])
	}

	// Different subnets
	return startIP + " - " + endIP
}

// countHosts calculates the number of hosts in the given IP range
func countHosts(startIP, endIP string) int {
	startIPInt := ipToInt(net.ParseIP(startIP))
	endIPInt := ipToInt(net.ParseIP(endIP))

	// Add 1 because the range is inclusive
	return int(endIPInt - startIPInt + 1)
}
