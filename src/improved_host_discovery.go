// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/src/improved_host_discovery.go
package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// enhancedScanIPRange improves host discovery by using multiple methods
func enhancedScanIPRange(startIP, endIP string, timeout time.Duration) []HostResult {
	var results []HostResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	log.Printf("Starting enhanced network scan from %s to %s...", startIP, endIP)

	// Convert IPs to integers for iteration
	startIPInt := ipToInt(net.ParseIP(startIP))
	endIPInt := ipToInt(net.ParseIP(endIP))

	// Calculate total hosts to scan
	totalHosts := int(endIPInt - startIPInt + 1)
	log.Printf("Scanning %d hosts...", totalHosts)

	// Rate limiting to avoid overwhelming the network
	// Increased from 100 to 250 for faster scanning
	semaphore := make(chan struct{}, 250)

	// Scan each IP in range
	for ipInt := startIPInt; ipInt <= endIPInt; ipInt++ {
		wg.Add(1)
		semaphore <- struct{}{} // acquire token

		go func(ip uint32) {
			defer wg.Done()
			defer func() { <-semaphore }() // release token

			ipStr := intToIP(ip).String()
			result := enhancedPingHost(ipStr, timeout)

			if result.Status == "Active" {
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}
		}(ipInt)
	}

	wg.Wait()
	log.Printf("Enhanced scan complete. Found %d active hosts.", len(results))
	return results
}

// enhancedPingHost uses multiple methods to check if a host is active
func enhancedPingHost(ip string, timeout time.Duration) HostResult {
	result := HostResult{
		IPAddress: ip,
		Status:    "Inactive",
		OpenPorts: 0,
	}

	// Method 1: ICMP ping (most reliable for devices that respond to ping)
	// Different ping parameters for macOS
	cmd := exec.Command("ping", "-c", "1", "-W", "1", "-t", "1", ip)
	if err := cmd.Run(); err == nil {
		result.Status = "Active"
	}

	// Method 2: TCP port scanning (for devices that block ICMP)
	if result.Status != "Active" {
		// Extended list of commonly open ports on various devices
		commonPorts := []int{
			// Web/App ports
			80, 443, 8080, 8443, 3000, 5000,
			// Remote access
			22, 23, 3389, 5900,
			// File sharing/SMB
			445, 139, 135,
			// IoT devices often use these
			1883, 8883, 5683,
			// Apple devices
			548, 62078,
			// Home automation
			1900, 8123, 8234,
			// Media devices
			1900, 5353, 7000,
			// DNS/DHCP
			53, 67, 68}

		for _, port := range commonPorts {
			address := fmt.Sprintf("%s:%d", ip, port)
			conn, err := net.DialTimeout("tcp", address, timeout/3) // Using shorter timeout for faster scanning
			if err == nil {
				conn.Close()
				result.Status = "Active"
				result.OpenPorts++
				break // Found one open port, that's enough to consider it active
			}
		}
	}

	// Method 3: ARP scanning (works by checking if the device has a MAC address in the ARP table)
	if result.Status != "Active" {
		// On macOS, use arp -n to check if the IP has an ARP entry
		cmd := exec.Command("arp", "-n", ip)
		output, err := cmd.Output()
		if err == nil && !strings.Contains(string(output), "no entry") {
			result.Status = "Active"
		}
	}

	// Try to resolve hostname if host is active
	if result.Status == "Active" {
		// Try DNS lookup
		names, err := net.LookupAddr(ip)
		if err == nil && len(names) > 0 {
			result.Hostname = strings.TrimSuffix(names[0], ".")
		} else {
			// If DNS lookup fails, use a generic name based on IP
			result.Hostname = "Device-" + strings.Replace(ip, ".", "-", -1)
		}
	}

	return result
}

// Try to determine device type based on open ports and other characteristics
func identifyDeviceType(ip string, openPorts []int) string {
	// This is a simplified version - you can expand this with more sophisticated detection

	// Check if it's likely a router
	if contains(openPorts, 80) && (contains(openPorts, 53) || contains(openPorts, 443)) {
		return "router"
	}

	// Check if it's likely a printer
	if contains(openPorts, 9100) || contains(openPorts, 515) {
		return "printer"
	}

	// Check if it's likely a media device
	if contains(openPorts, 5000) || contains(openPorts, 1900) || contains(openPorts, 7000) {
		return "media-device"
	}

	// Check if it's likely a mobile device
	if contains(openPorts, 62078) {
		return "mobile-device"
	}

	// Default to generic computer
	return "computer"
}

// Helper function to check if a slice contains a value
func contains(s []int, v int) bool {
	for _, a := range s {
		if a == v {
			return true
		}
	}
	return false
}
