// Package types provides centralized type definitions for the GoMap project
package types

import (
	"time"
)

// ScanResult represents the result of a complete network scan
type ScanResult struct {
	Subnet     string     `json:"subnet"`
	ScanTime   string     `json:"scan_time"`
	HostsFound []HostScan `json:"hosts_found"`
}

// HostScan represents a single host scan result with open ports
type HostScan struct {
	IPAddress string     `json:"ip_address"`
	Hostname  string     `json:"hostname"`
	OpenPorts []PortInfo `json:"open_ports"`
}

// PortInfo contains information about an open port
type PortInfo struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	Banner   string `json:"banner,omitempty"`
}

// ModernScanResult represents scan results for the modern UI
type ModernScanResult struct {
	ID           string    `json:"id"`
	Subnet       string    `json:"subnet"`
	ScanTime     time.Time `json:"scan_time"`
	HostsFound   int       `json:"hosts_found"`
	OpenPorts    int       `json:"open_ports"`
	ScanDuration string    `json:"scan_duration"`
}

// RecentScan holds basic information about a recent scan for UI display
type RecentScan struct {
	ID         string    `json:"id"`
	Subnet     string    `json:"subnet"`
	ScanTime   time.Time `json:"scan_time"`
	HostsFound int       `json:"hosts_found"`
}

// VulnerabilityInfo contains information about a potential vulnerability
type VulnerabilityInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	CVEID       string `json:"cve_id"`
	Port        int    `json:"port"`
	Service     string `json:"service"`
	References  string `json:"references"`
}
