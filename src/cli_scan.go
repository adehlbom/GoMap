package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"GoMap/types"
)

// RunCLIScan handles the CLI scan command with all options
func RunCLIScan(args []string) {
	var outputFile string
	var portRange string = "20-1024" // Default port range
	var subnetOverride string

	// Parse scan command options
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--output", "-o":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		case "--ports", "-p":
			if i+1 < len(args) {
				portRange = args[i+1]
				i++
			}
		case "--subnet", "-s":
			if i+1 < len(args) {
				subnetOverride = args[i+1]
				i++
			}
		}
	}

	// Run the network scan workflow
	results, err := runNetworkScan(subnetOverride, portRange)
	if err != nil {
		fmt.Printf("Error during scan: %v\n", err)
		os.Exit(1)
	}

	// Display results to console
	displayScanResults(results)

	// Save results to file if specified
	if outputFile != "" {
		err := saveResultsToJSON(results, outputFile)
		if err != nil {
			fmt.Printf("Error saving results to file: %v\n", err)
		} else {
			fmt.Printf("Results saved to %s\n", outputFile)
		}
	}
}

// runNetworkScan performs a complete network scan workflow
func runNetworkScan(subnetOverride string, portRangeStr string) (types.ScanResult, error) {
	var result types.ScanResult
	result.ScanTime = time.Now().Format(time.RFC3339)

	// Step 1: Detect subnet
	var subnet types.SubnetInfo
	var err error

	if subnetOverride != "" {
		// Parse user-provided subnet
		subnet = types.SubnetInfo{CIDRNotation: subnetOverride}
		// Extract IP address from CIDR notation
		if ip, _, err := net.ParseCIDR(subnetOverride); err == nil {
			subnet.IPAddress = ip.String()
		}
	} else {
		// Auto-detect subnet
		// Import GetDefaultLocalSubnet from network_utils
		localSubnet, err := GetDefaultLocalSubnet()
		if err != nil {
			return result, fmt.Errorf("subnet detection failed: %v", err)
		}
		// Convert to our types package SubnetInfo
		subnet = types.SubnetInfo{
			IPAddress:    localSubnet.IPAddress,
			CIDRNotation: localSubnet.CIDRNotation,
		}
	}

	fmt.Printf("Detected subnet: %s\n", subnet.CIDRNotation)
	result.Subnet = subnet.CIDRNotation

	// Step 2: Generate all IPs in the subnet
	ips, err := generateIPsFromCIDR(subnet.CIDRNotation)
	if err != nil {
		return result, fmt.Errorf("failed to generate IP list: %v", err)
	}

	fmt.Printf("Scanning %d IP addresses in subnet %s\n", len(ips), subnet.CIDRNotation)

	// Step 3: Discover live hosts using ping
	fmt.Println("Discovering live hosts...")
	liveHosts, err := discoverLiveHosts(ips)
	if err != nil {
		return result, fmt.Errorf("host discovery failed: %v", err)
	}

	fmt.Printf("Found %d live hosts\n", len(liveHosts))

	// Step 4: Parse port range
	startPort, endPort, err := parsePortRange(portRangeStr)
	if err != nil {
		return result, fmt.Errorf("invalid port range: %v", err)
	}

	// Step 5: Scan ports on each live host
	fmt.Printf("Scanning ports %d-%d on each live host...\n", startPort, endPort)
	for _, host := range liveHosts {
		fmt.Printf("Scanning %s...\n", host.IPAddress)
		hostScan := types.HostScan{
			IPAddress: host.IPAddress,
			Hostname:  host.Hostname,
			OpenPorts: []types.PortInfo{},
		}

		// Scan ports
		openPorts, err := scanPortsOnHost(host.IPAddress, startPort, endPort)
		if err != nil {
			fmt.Printf("Warning: Error scanning ports on %s: %v\n", host.IPAddress, err)
		}

		hostScan.OpenPorts = openPorts
		result.HostsFound = append(result.HostsFound, hostScan)
	}

	return result, nil
}

// generateIPsFromCIDR generates all IP addresses in a CIDR subnet
func generateIPsFromCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		ips = append(ips, ip.String())
	}

	// Remove network address and broadcast address for IPv4
	if len(ips) > 2 {
		return ips[1 : len(ips)-1], nil
	}

	return ips, nil
}

// incrementIP increments an IP address by 1
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// discoverLiveHosts finds which hosts in the IP list are online
func discoverLiveHosts(ips []string) ([]types.HostResult, error) {
	fmt.Printf("Discovering live hosts among %d IPs...\n", len(ips))

	var results []types.HostResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Use a semaphore to limit concurrency
	semaphore := make(chan struct{}, 100)

	for _, ip := range ips {
		wg.Add(1)
		semaphore <- struct{}{} // acquire token

		go func(ipAddr string) {
			defer wg.Done()
			defer func() { <-semaphore }() // release token

			localResult := pingHost(ipAddr, 1*time.Second)

			if localResult.Status == "Active" {
				// Convert to our types package HostResult
				typedResult := types.HostResult{
					IPAddress: localResult.IPAddress,
					Hostname:  localResult.Hostname,
					Status:    localResult.Status,
				}

				mu.Lock()
				results = append(results, typedResult)
				mu.Unlock()
			}
		}(ip)
	}

	wg.Wait()
	return results, nil
}

// parsePortRange parses a port range string (e.g., "20-1024") into start and end ports
func parsePortRange(portRange string) (int, int, error) {
	parts := strings.Split(portRange, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid port range format: %s (expected format: start-end)", portRange)
	}

	startPort, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start port: %v", err)
	}

	endPort, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end port: %v", err)
	}

	if startPort < 1 || startPort > 65535 || endPort < 1 || endPort > 65535 || startPort > endPort {
		return 0, 0, fmt.Errorf("port range must be between 1-65535 and start must be <= end")
	}

	return startPort, endPort, nil
}

// scanPortsOnHost scans a range of ports on a single host
func scanPortsOnHost(hostIP string, startPort, endPort int) ([]types.PortInfo, error) {
	var openPorts []types.PortInfo
	var mu sync.Mutex
	var scanWg sync.WaitGroup

	// Use a semaphore to limit concurrency
	portSemaphore := make(chan struct{}, 100)

	for port := startPort; port <= endPort; port++ {
		scanWg.Add(1)
		portSemaphore <- struct{}{} // acquire token

		go func(p int) {
			defer scanWg.Done()
			defer func() { <-portSemaphore }() // release token

			// Scan the port
			address := net.JoinHostPort(hostIP, strconv.Itoa(p))
			conn, err := net.DialTimeout("tcp", address, 1*time.Second)

			if err != nil {
				return // Port is closed or filtered
			}

			// Port is open
			defer conn.Close()

			portInfo := types.PortInfo{
				Port:     p,
				Protocol: "tcp",
				Service:  getServiceNameByPort(p),
			}

			// Try banner grabbing
			banner := getBanner(conn)
			if banner != "" {
				portInfo.Banner = banner
			}

			mu.Lock()
			openPorts = append(openPorts, portInfo)
			mu.Unlock()
		}(port)
	}

	scanWg.Wait()

	// Sort ports for consistent output
	sort.Slice(openPorts, func(i, j int) bool {
		return openPorts[i].Port < openPorts[j].Port
	})

	return openPorts, nil
}

// getBanner attempts to grab a service banner from an open connection
func getBanner(conn net.Conn) string {
	// Set a 1-second read deadline
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)

	if err != nil {
		return ""
	}

	// Clean and truncate the banner
	banner := strings.TrimSpace(string(buffer[:n]))
	if len(banner) > 100 {
		banner = banner[:100] + "..."
	}

	return cleanBanner(banner)
}

// displayScanResults prints scan results to the console
func displayScanResults(result types.ScanResult) {
	fmt.Println("\n=============================================")
	fmt.Printf("Scan Results for Subnet: %s\n", result.Subnet)
	fmt.Printf("Scan Time: %s\n", result.ScanTime)
	fmt.Printf("Hosts Found: %d\n", len(result.HostsFound))
	fmt.Println("=============================================")

	for _, host := range result.HostsFound {
		fmt.Printf("\nHost: %s", host.IPAddress)
		if host.Hostname != "" {
			fmt.Printf(" (%s)", host.Hostname)
		}
		fmt.Printf("\nOpen Ports: %d\n", len(host.OpenPorts))

		if len(host.OpenPorts) > 0 {
			fmt.Println("-----------------------------------------")
			fmt.Println("PORT\tSTATE\tSERVICE\tBANNER")
			fmt.Println("-----------------------------------------")

			for _, port := range host.OpenPorts {
				fmt.Printf("%d/%s\topen\t%s", port.Port, port.Protocol, port.Service)
				if port.Banner != "" {
					fmt.Printf("\t%s", port.Banner)
				}
				fmt.Println()
			}
		}
	}
	fmt.Println("=============================================")
}

// saveResultsToJSON saves scan results to a JSON file
func saveResultsToJSON(result types.ScanResult, filePath string) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, jsonData, 0644)
}
