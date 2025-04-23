package main

import (
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"GoMap/types"
)

// GetDefaultLocalSubnet returns the primary local subnet (usually the one with internet access)
func GetDefaultLocalSubnet() (types.SubnetInfo, error) {
	// Find all local subnets
	var subnets []types.SubnetInfo

	interfaces, err := net.Interfaces()
	if err != nil {
		return types.SubnetInfo{}, err
	}

	for _, iface := range interfaces {
		// Skip loopback, non-active, and interfaces without addresses
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			// Check if this is an IP network address
			ipNet, ok := addr.(*net.IPNet)

			// Skip if it's not an IP network or if it's an IPv6 address
			if !ok || ipNet.IP.To4() == nil {
				continue
			}

			// Skip loopback addresses
			if ipNet.IP.IsLoopback() {
				continue
			}

			// Skip link-local addresses (169.254.x.x)
			if ipNet.IP[0] == 169 && ipNet.IP[1] == 254 {
				continue
			}

			// Get CIDR notation
			cidr := ipNet.String()

			subnet := types.SubnetInfo{
				InterfaceName: iface.Name,
				IPAddress:     ipNet.IP.String(),
				SubnetMask:    net.IP(ipNet.Mask).String(),
				CIDRNotation:  cidr,
			}

			subnets = append(subnets, subnet)
		}
	}

	if len(subnets) == 0 {
		return types.SubnetInfo{}, nil
	}

	// Prioritize typical home/office network subnets over others
	for _, subnet := range subnets {
		// Typical home/office subnets often begin with 192.168 or 10
		if strings.HasPrefix(subnet.IPAddress, "192.168.") ||
			strings.HasPrefix(subnet.IPAddress, "10.") {
			return subnet, nil
		}
	}

	// If no specific type of subnet found, return the first one
	return subnets[0], nil
}

// NOTE: The functions below are already implemented elsewhere or causing redeclarations.
// Implementing them as new functions with different names to avoid conflicts.

// ScanNetworkSubnet scans a subnet for active hosts and returns a list of discovered hosts
// This is an alternative to any existing ScanSubnet function that might exist
func ScanNetworkSubnet(subnet types.SubnetInfo, concurrency int, timeout time.Duration) []types.HostResult {
	// Extract network information from CIDR
	ip, ipNet, err := net.ParseCIDR(subnet.CIDRNotation)
	if err != nil {
		return nil
	}

	var results []types.HostResult
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// Use a buffered channel as a semaphore to limit concurrency
	semaphore := make(chan struct{}, concurrency)

	// Increment the IP address
	inc := func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}

	// Start with the first IP in the subnet
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
		// Skip the network and broadcast addresses
		if ip[len(ip)-1] == 0 || ip[len(ip)-1] == 255 {
			continue
		}

		// Copy the IP so it doesn't change during the goroutine execution
		currentIP := make(net.IP, len(ip))
		copy(currentIP, ip)

		wg.Add(1)
		semaphore <- struct{}{} // Acquire a semaphore slot

		go func(ip net.IP) {
			defer func() {
				<-semaphore // Release a semaphore slot
				wg.Done()
			}()

			ipStr := ip.String()
			result := pingHost(ipStr, timeout)

			if result.Status == "Active" {
				mutex.Lock()
				results = append(results, result)
				mutex.Unlock()
			}
		}(currentIP)
	}

	wg.Wait()
	return results
}

// ScanHostPortRange scans a specific host for open ports within the given range
// This is an alternative to any existing ScanHostPorts function that might exist
func ScanHostPortRange(host types.HostResult, startPort, endPort int, concurrency int, timeout time.Duration) types.HostResult {
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Use a buffered channel as a semaphore to limit concurrency
	semaphore := make(chan struct{}, concurrency)

	result := host
	result.OpenPorts = 0
	// We'll only track the count since HostResult doesn't have a field for port details

	for port := startPort; port <= endPort; port++ {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire a semaphore slot

		go func(p int) {
			defer func() {
				<-semaphore // Release a semaphore slot
				wg.Done()
			}()

			address := net.JoinHostPort(host.IPAddress, strconv.Itoa(p))
			conn, err := net.DialTimeout("tcp", address, timeout)

			if err != nil {
				return // Port is closed or filtered
			}

			// Port is open
			// Note: We're not saving port details since HostResult doesn't have a field for them

			// Try to get banner if it's a common service that might offer one
			// Note: We're not saving banners since HostResult doesn't have a field for them

			conn.Close()

			mutex.Lock()
			result.OpenPorts++
			mutex.Unlock()
		}(port)
	}

	wg.Wait()
	return result
}

// DetermineDeviceType attempts to determine the device type based on open ports
// This is an alternative to any existing GetDeviceType function that might exist
func DetermineDeviceType(host types.HostResult) string {
	// Check for common device signatures
	if host.OpenPorts == 0 {
		return "Unknown Device"
	}

	// Since HostResult doesn't have detailed port information,
	// we can only make a general determination based on the number of open ports
	if host.OpenPorts > 5 {
		return "Server or Gateway Device"
	} else if host.OpenPorts > 2 {
		return "Workstation or Server"
	} else {
		return "End Device"
	}
}

// pingHost checks if a host is active and returns a HostResult
func pingHost(ip string, timeout time.Duration) types.HostResult {
	result := types.HostResult{
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

	// If ping didn't work, try TCP ports to determine if host is active
	if result.Status != "Active" {
		// Extended list of commonly open ports on various devices
		commonPorts := []int{80, 443, 22, 3389, 21, 23, 25, 53, 8080}
		for _, port := range commonPorts {
			address := net.JoinHostPort(ip, strconv.Itoa(port))
			conn, err := net.DialTimeout("tcp", address, timeout/2)
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
			result.Hostname = "Unknown"
		}
	}

	return result
}

// getServiceNameByPort returns the service name typically associated with a port
func getServiceNameByPort(port int) string {
	// Map of common ports to their service names
	portMap := map[int]string{
		20:   "FTP-data",
		21:   "FTP",
		22:   "SSH",
		23:   "Telnet",
		25:   "SMTP",
		53:   "DNS",
		67:   "DHCP",
		68:   "DHCP",
		80:   "HTTP",
		110:  "POP3",
		123:  "NTP",
		137:  "NetBIOS",
		138:  "NetBIOS",
		139:  "NetBIOS",
		143:  "IMAP",
		161:  "SNMP",
		162:  "SNMP",
		389:  "LDAP",
		443:  "HTTPS",
		445:  "SMB",
		465:  "SMTPS",
		514:  "Syslog",
		587:  "SMTP",
		636:  "LDAPS",
		993:  "IMAPS",
		995:  "POP3S",
		1433: "MSSQL",
		1521: "Oracle DB",
		3306: "MySQL",
		3389: "RDP",
		5432: "PostgreSQL",
		8080: "HTTP-Alt",
		8443: "HTTPS-Alt",
	}

	if service, exists := portMap[port]; exists {
		return service
	}

	return "Unknown"
}

// cleanBanner cleans a banner string for display
func cleanBanner(banner string) string {
	// Replace non-printable characters
	cleanedBanner := strings.Map(func(r rune) rune {
		if r < 32 || r > 126 {
			return ' '
		}
		return r
	}, banner)

	// Trim spaces and control characters
	cleanedBanner = strings.TrimSpace(cleanedBanner)

	return cleanedBanner
}
