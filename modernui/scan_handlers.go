// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/scan_handlers.go
package modernui

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Adapter functions to connect to the main scanning functionality

// startLocalNetworkScan is a helper function to scan the local network
func startLocalNetworkScan() {
	currentNavItem = "network"

	go func() {
		// Update UI
		scanActive = true
		scanStatus.SetText("Scanning local network...")
		progressBar.SetValue(0.1)

		scanStartTime := time.Now()

		// Get local subnet info
		subnet, err := getDefaultLocalSubnet()
		if err != nil {
			ShowErrorDialog("Failed to detect local network: " + err.Error())
			scanActive = false
			scanStatus.SetText("Ready")
			progressBar.SetValue(0)
			return
		}

		// Parse the CIDR to get start and end IPs
		startIP, endIP, err := parseCIDR(subnet)
		if err != nil {
			ShowErrorDialog("Failed to parse subnet: " + err.Error())
			scanActive = false
			scanStatus.SetText("Ready")
			progressBar.SetValue(0)
			return
		}

		// Update progress
		progressBar.SetValue(0.2)
		scanStatus.SetText(fmt.Sprintf("Scanning network %s...", subnet))

		// Debug logging before scan
		fmt.Printf("DEBUG: Starting scan from %s to %s\n", startIP, endIP)

		// Perform the scan
		results := scanIPRange(startIP, endIP)

		// Debug logging after scan
		fmt.Printf("DEBUG: Scan complete. Found %d hosts\n", len(results))
		for i, host := range results {
			fmt.Printf("DEBUG: Host %d: IP=%s, Hostname=%s, Status=%s, OpenPorts=%d\n",
				i, host.IPAddress, host.Hostname, host.Status, host.OpenPorts)
		}

		// Update results and UI
		hostResults = results
		fmt.Printf("DEBUG: Updated hostResults, length: %d\n", len(hostResults))

		// Switch to network tab to show results
		switchContent("network")

		// If host results table exists, refresh it
		if hostResultsTable != nil {
			hostResultsTable.Refresh()
		}

		// Record duration
		scanDuration := time.Since(scanStartTime)

		// Add to recent scans
		addRecentScan("network", subnet, len(results), scanDuration)

		// Update statistics
		updateStatistics()

		// Finish up
		progressBar.SetValue(1.0)
		scanActive = false
		lastScanTime = time.Now()
		scanStatus.SetText("Ready")
		scanSummaryLabel.SetText(fmt.Sprintf("Found %d hosts on %s", len(results), subnet))

		// Show completion message
		ShowInfoDialog("Scan Complete", fmt.Sprintf("Found %d active hosts on local network", len(results)))
	}()
}

// startHostDiscovery begins a host discovery scan with the modern UI
func startHostDiscovery(startIP, endIP string) {
	if scanActive {
		return
	}

	scanActive = true
	hostResults = []HostResult{}

	// If we have a host results table, refresh it
	if hostResultsTable != nil {
		hostResultsTable.Refresh()
	}

	progressBar.SetValue(0)
	scanStatus.SetText(fmt.Sprintf("Scanning IP range %s to %s...", startIP, endIP))

	scanStartTime := time.Now()

	// Start the scan in a goroutine
	go func() {
		// Perform the scan
		results := scanIPRange(startIP, endIP)

		// Update results and UI
		hostResults = results

		// If host results table exists, refresh it
		if hostResultsTable != nil {
			hostResultsTable.Refresh()
		}

		// Record scan duration
		scanDuration := time.Since(scanStartTime)

		// Add to recent scans
		addRecentScan("host", fmt.Sprintf("%s to %s", startIP, endIP), len(results), scanDuration)

		// Update statistics
		updateStatistics()

		progressBar.SetValue(1.0)
		scanActive = false
		lastScanTime = time.Now()
		scanStatus.SetText("Ready")
		scanSummaryLabel.SetText(fmt.Sprintf("Found %d hosts in range %s-%s", len(results), startIP, endIP))

		// Show completion message
		ShowInfoDialog("Scan Complete", fmt.Sprintf("Scan completed: Found %d active hosts", len(results)))
	}()
}

// getDefaultLocalSubnet gets information about the default local subnet
func getDefaultLocalSubnet() (string, error) {
	// This is a simplified implementation - in production it would
	// use the network utils from the main package

	// Get local IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("could not determine local IP: %v", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ipAddress := localAddr.IP.String()

	// Assume a standard /24 subnet - in production this should be detected more accurately
	parts := strings.Split(ipAddress, ".")
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid IP format")
	}

	cidr := fmt.Sprintf("%s.%s.%s.0/24", parts[0], parts[1], parts[2])
	return cidr, nil
}

// parseCIDR parses a CIDR notation and returns the start and end IPs
func parseCIDR(cidr string) (string, string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", err
	}

	// Get start IP (network address)
	startIP := ipNet.IP.String()

	// Calculate end IP based on mask
	mask := ipNet.Mask
	maskSize, _ := mask.Size()

	// Calculate the number of host addresses available
	numAddresses := 1 << (32 - maskSize)

	// Convert IP to 32-bit integer
	ipInt := ip4ToInt(ipNet.IP)

	// Calculate broadcast IP (end IP)
	endIPInt := ipInt + uint32(numAddresses) - 1
	endIP := intToIP4(endIPInt).String()

	// Remove network and broadcast addresses for usable range
	if maskSize < 31 { // Normal subnets where network/broadcast exist
		startParts := strings.Split(startIP, ".")
		lastOctet, _ := strconv.Atoi(startParts[3])
		startParts[3] = strconv.Itoa(lastOctet + 1)
		startIP = strings.Join(startParts, ".")

		endParts := strings.Split(endIP, ".")
		lastOctet, _ = strconv.Atoi(endParts[3])
		endParts[3] = strconv.Itoa(lastOctet - 1)
		endIP = strings.Join(endParts, ".")
	}

	return startIP, endIP, nil
}

// ip4ToInt converts an IPv4 address to a uint32
func ip4ToInt(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// intToIP4 converts a uint32 to an IPv4 address
func intToIP4(nn uint32) net.IP {
	ip := make(net.IP, 4)
	ip[0] = byte(nn >> 24)
	ip[1] = byte(nn >> 16)
	ip[2] = byte(nn >> 8)
	ip[3] = byte(nn)
	return ip
}

// scanIPRange scans a range of IPs for active hosts
func scanIPRange(startIP, endIP string) []HostResult {
	var results []HostResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Print debug information
	fmt.Printf("DEBUG: Scanning IP range from %s to %s\n", startIP, endIP)

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

			ipStr := intToIP4(ip).String()
			result := pingHost(ipStr, timeoutDuration)

			if result.Status == "Active" {
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				fmt.Printf("DEBUG: Found active host: %s (%s)\n", result.IPAddress, result.Hostname)
			}
		}(ipInt)
	}

	wg.Wait()
	fmt.Printf("DEBUG: Scan complete. Found %d active hosts.\n", len(results))
	return results
}

// exportScanResults exports the port scan results to a file
func exportScanResults() {
	// This would implement export functionality
	ShowInfoDialog("Export", "Results exported successfully (placeholder)")
}

// exportHostResults exports the host discovery results to a file
func exportHostResults() {
	// This would implement export functionality
	ShowInfoDialog("Export", "Host results exported successfully (placeholder)")
}

// showHelp displays a help dialog
func showHelp() {
	// This would implement help dialog
	ShowInfoDialog("Help", "GoMap Help\n\nThis is a network scanning and mapping tool.\n\nUse the sidebar to navigate between different functions.")
}
