// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/src/modern_handlers.go
package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ScanResult represents the result of a port scan
type ScanResult struct {
	Port     int
	Status   string
	Service  string
	Protocol string
	Time     time.Duration
}

// HostResult is already defined in host_discovery.go

// startLocalNetworkScan is a helper function to scan the local network with the modern UI
func startLocalNetworkScan() {
	currentNavItem = "network"

	go func() {
		// Update UI
		scanActive = true
		currentScanStatus.SetText("Scanning local network...")
		progressBar.SetValue(0.1)

		scanStartTime := time.Now()

		// Get local subnet info
		subnet, err := GetDefaultLocalSubnet()
		if err != nil {
			showErrorDialog("Failed to detect local network: " + err.Error())
			scanActive = false
			currentScanStatus.SetText("Ready")
			progressBar.SetValue(0)
			return
		}

		// Parse the CIDR to get start and end IPs
		startIP, endIP, err := parseCIDR(subnet.CIDRNotation)
		if err != nil {
			showErrorDialog("Failed to parse subnet: " + err.Error())
			scanActive = false
			currentScanStatus.SetText("Ready")
			progressBar.SetValue(0)
			return
		}

		// Update progress
		progressBar.SetValue(0.2)
		currentScanStatus.SetText(fmt.Sprintf("Scanning network %s...", subnet.CIDRNotation))

		// Perform the scan
		results := scanIPRange(startIP, endIP, timeoutDuration)

		// Update results and UI
		hostResults = results

		// Switch to network tab to show results
		switchContent("network")

		// If host results table exists, refresh it
		if hostResultsList != nil {
			hostResultsList.Refresh()
		}

		// Record duration
		scanDuration := time.Since(scanStartTime)

		// Add to recent scans
		addRecentScan("network", subnet.CIDRNotation, len(results), scanDuration)

		// Update statistics
		updateStatistics()

		// Finish up
		progressBar.SetValue(1.0)
		scanActive = false
		lastScanTime = time.Now()
		currentScanStatus.SetText("Ready")
		scanSummaryLabel.SetText(fmt.Sprintf("Found %d hosts on %s", len(results), subnet.CIDRNotation))

		// Show completion message
		showInfoDialog("Scan Complete", fmt.Sprintf("Found %d active hosts on local network", len(results)))
	}()
}

// startHostDiscovery begins a host discovery scan with the modern UI
func startHostDiscovery(startIP, endIP string, timeout time.Duration) {
	if scanActive {
		return
	}

	scanActive = true
	hostResults = []HostResult{}

	// If we have a host results table, refresh it
	if hostResultsList != nil {
		hostResultsList.Refresh()
	}

	progressBar.SetValue(0)
	currentScanStatus.SetText(fmt.Sprintf("Scanning IP range %s to %s...", startIP, endIP))

	scanStartTime := time.Now()

	// Start the scan in a goroutine
	go func() {
		// Perform the scan
		results := scanIPRange(startIP, endIP, timeout)

		// Update results and UI
		hostResults = results

		// If host results table exists, refresh it
		if hostResultsList != nil {
			hostResultsList.Refresh()
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
		currentScanStatus.SetText("Ready")
		scanSummaryLabel.SetText(fmt.Sprintf("Found %d hosts in range %s-%s", len(results), startIP, endIP))

		// Show completion message
		showInfoDialog("Scan Complete", fmt.Sprintf("Scan completed: Found %d active hosts", len(results)))
	}()
}

// startScan initiates a port scan with the modern UI
func startScan(ip string, minPort, maxPort int) {
	if scanActive {
		return
	}

	scanActive = true
	scanResults = []ScanResult{}

	// If we have a results table, refresh it
	if resultsList != nil {
		resultsList.Refresh()
	}

	progressBar.SetValue(0)
	currentScanStatus.SetText("Scanning ports...")

	scanStartTime := time.Now()

	// Start the scan in a goroutine
	go func() {
		totalPorts := maxPort - minPort + 1
		portsDone := 0

		// Initialize wait group
		var wg sync.WaitGroup
		wg.Add(totalPorts)

		// Limit concurrency
		maxConcurrent := 100
		if val, ok := scanConfig["maxParallel"]; ok {
			if intVal, ok := val.(int); ok && intVal > 0 {
				maxConcurrent = intVal
			}
		}
		semaphore := make(chan struct{}, maxConcurrent)

		// Start scan for each port
		for port := minPort; port <= maxPort; port++ {
			semaphore <- struct{}{} // Acquire token

			go func(p int) {
				defer func() {
					<-semaphore // Release token
					portsDone++
					progress := float64(portsDone) / float64(totalPorts)
					progressBar.SetValue(progress)
					wg.Done()
				}()

				scan(ip, p)
			}(port)
		}

		// Wait for all scans to complete
		wg.Wait()

		// Record scan duration
		scanDuration := time.Since(scanStartTime)

		// Add to recent scans
		addRecentScan("port", ip, len(scanResults), scanDuration)

		// Update statistics
		updateStatistics()

		scanActive = false
		currentScanStatus.SetText("Ready")
		lastScanTime = time.Now()
		scanSummaryLabel.SetText(fmt.Sprintf("Found %d open ports on %s", len(scanResults), ip))

		// If we have a results table, refresh it
		if resultsList != nil {
			resultsList.Refresh()
		}

		// Show completion message
		showInfoDialog("Scan Complete", fmt.Sprintf("Scan complete. Found %d open ports.", len(scanResults)))
	}()
}

// scan performs the actual port scan and updates the results
func scan(ip string, port int) {
	startTime := time.Now()
	result := tcp_scan_gui(ip, port)
	scanDuration := time.Since(startTime)

	// Add timing information
	result.Time = scanDuration

	if result.Status == "OPEN" {
		// Add to results
		scanResults = append(scanResults, result)

		// Try to refresh UI if available
		if resultsList != nil {
			resultsList.Refresh()
		}
	}
}

// cancelScan cancels any active scanning operations
func cancelScan() {
	if !scanActive {
		return
	}

	// Cannot truly cancel goroutines, but we can mark as inactive and update UI
	scanActive = false
	progressBar.SetValue(0)
	currentScanStatus.SetText("Scan cancelled")
	showInfoDialog("Operation Cancelled", "Scan operation cancelled by user")
}

// updateHostDetails updates the Host Details view with information about the selected host
func updateHostDetails(host HostResult) {
	selectedHostIP = host.IPAddress

	// If we're not on the host details screen, switch to it
	if currentNavItem != "host" {
		switchContent("host")
	}

	// Refresh the display (will be implemented when the UI gets updated)
	// The new UI will handle this better than updating the container objects directly
}

// Dialog helper functions

// showErrorDialog displays an error message
func showErrorDialog(message string) {
	dialog.ShowError(fmt.Errorf(message), mainWindow)
}

// showInfoDialog displays an informational message
func showInfoDialog(title string, message string) {
	dialog.ShowInformation(title, message, mainWindow)
}

// showHelp displays a help dialog
func showHelp() {
	helpWindow := fyne.CurrentApp().NewWindow("GoMap Help")

	tabs := container.NewAppTabs(
		container.NewTabItem("Overview", createHelpTab("Overview",
			"GoMap is a network scanning tool designed to help you discover "+
				"and analyze devices on your network. You can scan for hosts, "+
				"open ports, and service information.")),

		container.NewTabItem("Dashboard", createHelpTab("Dashboard",
			"The dashboard provides an overview of your scanning activity and results. "+
				"It shows statistics about discovered hosts and open ports, recent scan "+
				"history, and quick access to common scanning functions.")),

		container.NewTabItem("Network Scanner", createHelpTab("Network Scanner",
			"The Network Scanner allows you to discover active hosts on your network. "+
				"You can scan your local network or specify a custom IP range using CIDR "+
				"notation (e.g., 192.168.1.0/24) or range format (e.g., 192.168.1.1-192.168.1.254).")),

		container.NewTabItem("Host Details", createHelpTab("Host Details",
			"The Host Details view shows detailed information about a selected host, "+
				"including its IP address, hostname, status, and open ports. You can also "+
				"initiate port scans and other analyses from this screen.")),

		container.NewTabItem("Port Scanner", createHelpTab("Port Scanner",
			"The Port Scanner allows you to check for open ports on a specific host. "+
				"You can choose from preset port ranges or define a custom range. "+
				"For each open port, GoMap will attempt to identify the running service.")),

		container.NewTabItem("Settings", createHelpTab("Settings",
			"The Settings screen allows you to configure scan behavior, including "+
				"timeout duration, parallelism, and detection methods. You can also "+
				"adjust visual preferences and enable or disable certain features.")),

		container.NewTabItem("About", createHelpTab("About",
			"GoMap Network Scanner\nVersion 2.0\n\n"+
				"Created by Anders Dehlbom\n"+
				"GitHub: https://github.com/adehlbom/GoMap\n\n"+
				"© 2025")),
	)

	closeButton := widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {
		helpWindow.Close()
	})

	helpWindow.SetContent(container.NewBorder(
		nil,
		container.NewCenter(closeButton),
		nil, nil,
		tabs,
	))

	helpWindow.Resize(fyne.NewSize(600, 500))
	helpWindow.Show()
}

// createHelpTab creates a tab for the help dialog
func createHelpTab(title string, content string) fyne.CanvasObject {
	helpText := widget.NewLabel(content)
	helpText.Wrapping = fyne.TextWrapWord

	return container.NewScroll(container.NewPadded(helpText))
}

// exportScanResults exports the port scan results to a file
func exportScanResults() {
	if len(scanResults) == 0 {
		showErrorDialog("No scan results to export")
		return
	}

	// Use a save dialog to specify where to save the file
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			showErrorDialog(err.Error())
			return
		}
		if writer == nil {
			return // User cancelled
		}

		// Create CSV content
		csvContent := "Port,Status,Service,Protocol,Response Time\n"
		for _, result := range scanResults {
			csvContent += fmt.Sprintf("%d,%s,%s,%s,%s\n",
				result.Port,
				result.Status,
				result.Service,
				"TCP", // Currently only TCP is supported
				result.Time.String(),
			)
		}

		// Write to file
		_, err = writer.Write([]byte(csvContent))
		writer.Close()

		if err != nil {
			showErrorDialog("Failed to export results: " + err.Error())
			return
		}

		showInfoDialog("Export Successful", "Scan results exported successfully")
	}, mainWindow)
}

// exportHostResults exports the host discovery results to a file
func exportHostResults() {
	if len(hostResults) == 0 {
		showErrorDialog("No host results to export")
		return
	}

	// Use a save dialog to specify where to save the file
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			showErrorDialog(err.Error())
			return
		}
		if writer == nil {
			return // User cancelled
		}

		// Create CSV content
		csvContent := "IP Address,Hostname,Status,Open Ports\n"
		for _, result := range hostResults {
			csvContent += fmt.Sprintf("%s,%s,%s,%d\n",
				result.IPAddress,
				result.Hostname,
				result.Status,
				result.OpenPorts,
			)
		}

		// Write to file
		_, err = writer.Write([]byte(csvContent))
		writer.Close()

		if err != nil {
			showErrorDialog("Failed to export results: " + err.Error())
			return
		}

		showInfoDialog("Export Successful", "Host results exported successfully")
	}, mainWindow)
}

// showPortRangeDialog displays a dialog for custom port range input
func showPortRangeDialog() {
	// Create entry fields
	minEntry := widget.NewEntry()
	minEntry.SetPlaceHolder("e.g. 1")

	maxEntry := widget.NewEntry()
	maxEntry.SetPlaceHolder("e.g. 1024")

	// Create the dialog
	portDialog := dialog.NewCustom("Custom Port Range", "Apply", container.NewVBox(
		widget.NewLabel("Specify the port range to scan:"),
		container.NewGridWithColumns(2,
			widget.NewLabel("Minimum Port:"),
			minEntry,
			widget.NewLabel("Maximum Port:"),
			maxEntry,
		),
	), mainWindow)

	// Add a callback for the Apply button
	portDialog.SetOnClosed(func() {
		// Parse min port
		if minEntry.Text != "" {
			if min, err := strconv.Atoi(minEntry.Text); err == nil {
				if min >= 1 && min <= 65535 {
					portRangeMin = min
				} else {
					showErrorDialog("Minimum port must be between 1 and 65535")
					return
				}
			} else {
				showErrorDialog("Invalid minimum port number")
				return
			}
		}

		// Parse max port
		if maxEntry.Text != "" {
			if max, err := strconv.Atoi(maxEntry.Text); err == nil {
				if max >= 1 && max <= 65535 && max >= portRangeMin {
					portRangeMax = max
				} else if max < portRangeMin {
					showErrorDialog("Maximum port must be greater than minimum port")
					return
				} else {
					showErrorDialog("Maximum port must be between 1 and 65535")
					return
				}
			} else {
				showErrorDialog("Invalid maximum port number")
				return
			}
		}

		// Update UI to show selected range
		showInfoDialog("Port Range Updated",
			fmt.Sprintf("Custom port range set to %d-%d", portRangeMin, portRangeMax))
	})

	portDialog.Show()
}

// tcp_scan_gui is a modernized version of tcp_scan for the GUI
func tcp_scan_gui(ip_address string, port int) ScanResult {
	result := ScanResult{
		Port:     port,
		Status:   "CLOSED",
		Service:  "",
		Protocol: "TCP",
		Time:     0,
	}

	// Use JoinHostPort which handles both IPv4 and IPv6 correctly
	address := net.JoinHostPort(ip_address, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", address, timeoutDuration)
	if err != nil {
		return result
	}

	// Port is open
	result.Status = "OPEN"

	// Try to get service information
	serviceName := getServiceName(port)
	if serviceName != "" {
		result.Service = serviceName
	} else {
		// Try banner grabbing
		buffer := make([]byte, 4096)
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		numBytesRead, err := conn.Read(buffer)
		if err == nil && numBytesRead > 0 {
			// Clean banner for display
			banner := string(buffer[0:numBytesRead])
			// Limit length for display
			if len(banner) > 50 {
				banner = banner[:50] + "..."
			}
			result.Service = banner
		}
	}

	conn.Close()
	return result
}

// getServiceName returns a service name for common ports
func getServiceName(port int) string {
	commonServices := map[int]string{
		20:    "FTP-Data",
		21:    "FTP",
		22:    "SSH",
		23:    "Telnet",
		25:    "SMTP",
		53:    "DNS",
		67:    "DHCP-Server",
		68:    "DHCP-Client",
		69:    "TFTP",
		80:    "HTTP",
		88:    "Kerberos",
		110:   "POP3",
		123:   "NTP",
		137:   "NetBIOS-NS",
		138:   "NetBIOS-DGM",
		139:   "NetBIOS-SSN",
		143:   "IMAP",
		161:   "SNMP",
		389:   "LDAP",
		443:   "HTTPS",
		445:   "SMB",
		465:   "SMTPS",
		514:   "Syslog",
		546:   "DHCPv6",
		547:   "DHCPv6-Server",
		587:   "SMTP-Submission",
		631:   "IPP",
		636:   "LDAPS",
		993:   "IMAPS",
		995:   "POP3S",
		1194:  "OpenVPN",
		1433:  "MSSQL",
		1723:  "PPTP",
		3306:  "MySQL",
		3389:  "RDP",
		5432:  "PostgreSQL",
		5900:  "VNC",
		8080:  "HTTP-Alt",
		8443:  "HTTPS-Alt",
		27017: "MongoDB",
	}

	if service, ok := commonServices[port]; ok {
		return service
	}
	return ""
}
