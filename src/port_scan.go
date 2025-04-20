package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Default timeout for port connections
var defaultTimeout = 5 * time.Second

// WaitGroup for synchronizing goroutines
var wg sync.WaitGroup

// tcp_scan scans a specific port and prints results to console
// Used in command-line mode
func tcp_scan(ip_address string, port int) {
	defer wg.Done()

	// Format the address as IP:PORT (using JoinHostPort for IPv6 compatibility)
	address := net.JoinHostPort(ip_address, strconv.Itoa(port))

	// Attempt to establish a connection with timeout
	conn, err := net.DialTimeout("tcp", address, defaultTimeout)
	if err != nil {
		// Port is closed or filtered
		return
	}
	defer conn.Close() // Always close the connection when done

	// Port is open, attempt to identify the service
	fmt.Printf("%d/tcp OPEN", port)

	// Try to get banner information
	serviceName := getServiceNameByPort(port)
	if serviceName != "" {
		fmt.Printf(" | SERVICE: %s", serviceName)
	}

	// Try banner grabbing for additional info
	buffer := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(defaultTimeout))
	bytesRead, err := conn.Read(buffer)

	if err == nil && bytesRead > 0 {
		// Clean up the banner data
		banner := cleanBanner(string(buffer[:bytesRead]))
		fmt.Printf(" | BANNER: %s", banner)
	}

	fmt.Println()
}

// isCommonService checks if port is associated with a commonly known service
func isCommonService(port int) bool {
	commonPorts := []int{21, 22, 23, 25, 53, 80, 110, 143, 443, 465, 587, 993, 995, 3306, 3389, 5432, 8080}
	for _, p := range commonPorts {
		if p == port {
			return true
		}
	}
	return false
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
		5900: "VNC",
		8080: "HTTP-Alt",
		8443: "HTTPS-Alt",
	}

	if service, found := portMap[port]; found {
		return service
	}
	return ""
}

// cleanBanner sanitizes banner information for display
func cleanBanner(banner string) string {
	// Limit length and clean non-printable characters
	if len(banner) > 50 {
		banner = banner[:50] + "..."
	}

	// Replace control characters and make it safe for output
	banner = strings.ReplaceAll(banner, "\n", " ")
	banner = strings.ReplaceAll(banner, "\r", "")

	return banner
}

// portScanRange performs a port scan over a range and returns open ports
// This can be used by other functions in the application
func portScanRange(ip string, startPort, endPort int) []int {
	var openPorts []int
	var mu sync.Mutex // Mutex to safely append to openPorts
	var wg sync.WaitGroup

	// Set a limit on concurrent goroutines to prevent overwhelming the network or target
	semaphore := make(chan struct{}, 100)

	for port := startPort; port <= endPort; port++ {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire a slot

		go func(p int) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the slot

			address := net.JoinHostPort(ip, strconv.Itoa(p))
			conn, err := net.DialTimeout("tcp", address, defaultTimeout)
			if err == nil {
				conn.Close()
				mu.Lock()
				openPorts = append(openPorts, p)
				mu.Unlock()
			}
		}(port)
	}

	wg.Wait()
	return openPorts
}
