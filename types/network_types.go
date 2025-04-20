// Package types provides centralized type definitions for the GoMap project
package types

// SubnetInfo contains information about a network subnet
type SubnetInfo struct {
	InterfaceName string `json:"interface_name"`
	IPAddress     string `json:"ip_address"`
	SubnetMask    string `json:"netmask"`
	Gateway       string `json:"gateway"`
	CIDRNotation  string `json:"cidr_notation"`
}

// HostResult represents information about a discovered host
type HostResult struct {
	IPAddress string  `json:"ip_address"`
	Hostname  string  `json:"hostname"`
	Status    string  `json:"status"`
	RTT       float64 `json:"rtt"`
	MAC       string  `json:"mac,omitempty"`
	Vendor    string  `json:"vendor,omitempty"`
	OpenPorts int     `json:"open_ports"`
}

// NetworkNode represents a node in the network visualization
type NetworkNode struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Label     string `json:"label"`
	IP        string `json:"ip"`
	MAC       string `json:"mac"`
	Vendor    string `json:"vendor"`
	Status    string `json:"status"`
	OpenPorts int    `json:"open_ports"`
}
