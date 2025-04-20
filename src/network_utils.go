// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/src/network_utils.go
package main

import (
	"errors"
	"net"
	"strings"

	"GoMap/types"
)

// GetLocalSubnets detects all available local subnets on the system
func GetLocalSubnets() ([]types.SubnetInfo, error) {
	var subnets []types.SubnetInfo

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
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
				IPAddress:    ipNet.IP.String(),
				Netmask:      net.IP(ipNet.Mask).String(),
				CIDRNotation: cidr,
			}

			subnets = append(subnets, subnet)
		}
	}

	if len(subnets) == 0 {
		return nil, errors.New("no active network interfaces found")
	}

	return subnets, nil
}

// GetDefaultLocalSubnet returns the primary local subnet (usually the one with internet access)
func GetDefaultLocalSubnet() (types.SubnetInfo, error) {
	subnets, err := GetLocalSubnets()
	if err != nil {
		return types.SubnetInfo{}, err
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
	if len(subnets) > 0 {
		return subnets[0], nil
	}

	return types.SubnetInfo{}, errors.New("no suitable subnet found")
}

// ScanLocalSubnet performs a scan of the local subnet for active hosts
func ScanLocalSubnet() ([]types.HostResult, error) {
	// Get the default local subnet
	subnet, err := GetDefaultLocalSubnet()
	if err != nil {
		return nil, err
	}

	// Parse the CIDR to get start and end IPs
	startIP, endIP, err := parseCIDR(subnet.CIDRNotation)
	if err != nil {
		return nil, err
	}

	// Perform the scan
	localResults := scanIPRange(startIP, endIP, defaultTimeout)

	// Convert to our centralized types
	typedResults := make([]types.HostResult, len(localResults))
	for i, host := range localResults {
		typedResults[i] = types.HostResult{
			IPAddress: host.IPAddress,
			Hostname:  host.Hostname,
			Status:    host.Status,
		}
	}

	return typedResults, nil
}
