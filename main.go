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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"GoMap/modernui"
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

// SubnetInfo contains information about a detected subnet
type SubnetInfo struct {
	InterfaceName string
	IPAddress     string
	SubnetMask    string
	CIDRNotation  string // e.g., 192.168.1.0/24
	Gateway       string // May be empty if not detected
}

// HostResult represents a scanned host with its status and information
type HostResult struct {
	IPAddress string
	Hostname  string
	Status    string
	OpenPorts int
}

func main() {
	// Check if user wants to use the modern UI
	useModernUI := true

	// Check for UI choice flag
	for _, arg := range os.Args {
		if arg == "--classic" || arg == "-c" {
			useModernUI = false
			break
		}
	}

	// Print banner
	fmt.Println("==========================================")
	fmt.Println("         GoMap Network Scanner            ")
	fmt.Println("==========================================")

	// Check if running in command-line mode or GUI mode
	if len(os.Args) > 1 && os.Args[1] != "--classic" && os.Args[1] != "-c" {
		// Command-line mode
		handleCLICommands(os.Args[1:])
	} else {
		// Launch the application with the selected UI
		launchApp(useModernUI)
	}
}

// launchApp starts the application with the selected UI type
func launchApp(useModernUI bool) {
	// Create the application instance
	goMapApp := app.New()

	// Create the main window
	var mainWindow fyne.Window
	if useModernUI {
		fmt.Println("Starting with Modern UI...")
		mainWindow = goMapApp.NewWindow("GoMap - Network Scanner v2.0")
		mainWindow.Resize(fyne.NewSize(1024, 768))

		// Create a wrapper function to match the expected ScannerFunc signature
		scannerWrapper := func(subnetOverride string, portRangeStr string) (interface{}, error) {
			return runNetworkScan(subnetOverride, portRangeStr)
		}

		// Register the scanner implementation with the modernui package
		modernui.RegisterScanner(scannerWrapper)

		// Initialize the modern UI
		content := modernui.CreateModernUI(mainWindow)
		mainWindow.SetContent(content)
	} else {
		fmt.Println("Starting with Classic UI...")
		mainWindow = goMapApp.NewWindow("GoMap - Network Scanner")
		mainWindow.Resize(fyne.NewSize(900, 700))

		// For now, just show a message that classic UI isn't integrated yet
		label := widget.NewLabel("Classic UI not implemented in this version")
		content := container.NewCenter(label)
		mainWindow.SetContent(content)
	}

	mainWindow.CenterOnScreen()
	mainWindow.ShowAndRun()
}

// handleCLICommands processes command-line arguments and executes the appropriate action
func handleCLICommands(args []string) {
	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	command := args[0]

	switch command {
	case "scan":
		// Execute network scan using our built-in functionality
		fmt.Println("Starting network scan...")
		executeNetworkScan(args[1:])
	case "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

// printHelp displays usage information for command-line operation
func printHelp() {
	fmt.Println("Usage: gomap [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  scan            Perform subnet discovery, host discovery, and port scanning")
	fmt.Println("  help            Display this help message")
	fmt.Println("\nOptions for 'scan':")
	fmt.Println("  --output, -o    Specify output file path for JSON results")
	fmt.Println("  --ports, -p     Specify port range (e.g., 20-1024, default)")
	fmt.Println("  --subnet, -s    Specify subnet to scan (e.g., 192.168.1.0/24)")
	fmt.Println("                  If not specified, auto-detection will be used")
	fmt.Println("\nExamples:")
	fmt.Println("  gomap scan")
	fmt.Println("  gomap scan -o results.json")
	fmt.Println("  gomap scan -p 1-1000 -s 10.0.0.0/24")
}

// executeNetworkScan performs a full network scan (subnet detection, host discovery, port scanning)
func executeNetworkScan(args []string) {
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
func runNetworkScan(subnetOverride string, portRangeStr string) (ScanResult, error) {
	var result ScanResult
	result.ScanTime = time.Now().Format(time.RFC3339)

	// Step 1: Detect subnet
	var subnet SubnetInfo
	var err error

	if subnetOverride != "" {
		// Parse user-provided subnet
		subnet = SubnetInfo{CIDRNotation: subnetOverride}
		// Extract IP address from CIDR notation
		if ip, _, err := net.ParseCIDR(subnetOverride); err == nil {
			subnet.IPAddress = ip.String()
		}
	} else {
		// Auto-detect subnet
		subnetInfo, err := GetDefaultLocalSubnet()
		if err == nil {
			// Convert from scanner_bridge SubnetInfo to local SubnetInfo
			subnet = SubnetInfo{
				InterfaceName: subnetInfo.InterfaceName,
				IPAddress:     subnetInfo.IPAddress,
				SubnetMask:    subnetInfo.SubnetMask,
				CIDRNotation:  subnetInfo.CIDRNotation,
				Gateway:       subnetInfo.Gateway,
			}
		}
		if err != nil {
			return result, fmt.Errorf("subnet detection failed: %v", err)
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
		hostScan := HostScan{
			IPAddress: host.IPAddress,
			Hostname:  host.Hostname,
			OpenPorts: []PortInfo{},
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
func discoverLiveHosts(ips []string) ([]HostResult, error) {
	fmt.Printf("Discovering live hosts among %d IPs...\n", len(ips))

	var results []HostResult
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

			result := pingHost(ipAddr, 1*time.Second)

			if result.Status == "Active" {
				// Convert from types.HostResult to local HostResult
				localResult := HostResult{
					IPAddress: result.IPAddress,
					Hostname:  result.Hostname,
					Status:    result.Status,
					OpenPorts: result.OpenPorts,
				}

				mu.Lock()
				results = append(results, localResult)
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
func scanPortsOnHost(hostIP string, startPort, endPort int) ([]PortInfo, error) {
	var openPorts []PortInfo
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

			portInfo := PortInfo{
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

	return cleanBannerText(banner)
}

// cleanBannerText cleans a banner string for display
func cleanBannerText(banner string) string {
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

// displayScanResults prints scan results to the console
func displayScanResults(result ScanResult) {
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
func saveResultsToJSON(result ScanResult, filePath string) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, jsonData, 0644)
}
