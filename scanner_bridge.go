// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/scanner_bridge_new.go
package main

import (
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"GoMap/types" // Import the types package
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
