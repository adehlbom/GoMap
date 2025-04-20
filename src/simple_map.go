package main

import (
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// createNetworkMapTab creates a simplified network map visualization
func createNetworkMapTab() fyne.CanvasObject {
	// Create a title
	title := widget.NewLabelWithStyle("Network Map Visualization",
		fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Create the scan button
	scanButton := widget.NewButton("Scan Network", func() {
		go scanNetwork()
	})

	// Results list
	networkList := widget.NewList(
		func() int { return len(networkNodes) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Device"),
				widget.NewLabel("Status"),
				widget.NewLabel("Ports"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			node := networkNodes[id]
			items := obj.(*fyne.Container).Objects
			items[0].(*widget.Label).SetText(fmt.Sprintf("%s (%s)", node.Hostname, node.IPAddress))
			items[1].(*widget.Label).SetText(node.DeviceType)
			items[2].(*widget.Label).SetText(fmt.Sprintf("%d open", len(node.OpenPorts)))
		},
	)

	// Create a details panel
	detailsLabel := widget.NewLabel("Select a device to view details")

	// Handle selection
	networkList.OnSelected = func(id widget.ListItemID) {
		node := networkNodes[id]
		details := fmt.Sprintf("Device: %s (%s)\nType: %s\n\nOpen Ports: %d\n",
			node.Hostname, node.IPAddress, node.DeviceType, len(node.OpenPorts))

		for _, port := range node.OpenPorts {
			service := getServiceNameByPort(port)
			details += fmt.Sprintf("- Port %d: %s\n", port, service)
		}

		if len(node.Vulnerabilities) > 0 {
			details += "\nPotential Vulnerabilities:\n"
			for _, vuln := range node.Vulnerabilities {
				details += fmt.Sprintf("- %s\n", vuln)
			}
		}

		detailsLabel.SetText(details)
	}

	// Main layout
	return container.NewBorder(
		container.NewVBox(
			title,
			container.NewHBox(
				scanButton,
				widget.NewButton("Refresh", func() {
					networkList.Refresh()
				}),
			),
			widget.NewSeparator(),
		),
		nil, nil, nil,
		container.NewHSplit(
			container.NewVScroll(networkList),
			container.NewVScroll(detailsLabel),
		),
	)
}

// Network nodes for the map
type NetworkNode struct {
	IPAddress       string
	Hostname        string
	DeviceType      string
	OpenPorts       []int
	Vulnerabilities []string
}

var (
	networkNodes []NetworkNode
	networkMutex sync.Mutex
)

// Scan the network and build the node list
func scanNetwork() {
	dialog.ShowInformation("Network Scan", "Starting network scan...", mainWindow)
	progressBar.SetValue(0.1)

	// Clear existing data
	networkMutex.Lock()
	networkNodes = []NetworkNode{}
	networkMutex.Unlock()

	// Get subnet info
	subnet, err := GetDefaultLocalSubnet()
	if err != nil {
		dialog.ShowError(err, mainWindow)
		return
	}

	// Add router node
	networkMutex.Lock()
	networkNodes = append(networkNodes, NetworkNode{
		IPAddress:  subnet.Gateway,
		Hostname:   "Router",
		DeviceType: "router",
	})
	networkMutex.Unlock()

	// Parse the CIDR to get start and end IPs
	startIP, endIP, err := parseCIDR(subnet.CIDRNotation)
	if err != nil {
		dialog.ShowError(err, mainWindow)
		return
	}

	// Scan for hosts using improved scanning method
	progressBar.SetValue(0.3)
	results := enhancedScanIPRange(startIP, endIP, defaultTimeout)
	progressBar.SetValue(0.6)

	// Process results
	for _, host := range results {
		// Skip router (already added)
		if host.IPAddress == subnet.Gateway {
			continue
		}

		// Scan for ports
		openPorts := portScanRange(host.IPAddress, 1, 1024)

		// Create a simple vulnerability list
		var vulns []string
		for _, port := range openPorts {
			switch port {
			case 21:
				vulns = append(vulns, "FTP service - Often has weak authentication")
			case 23:
				vulns = append(vulns, "Telnet service - Transmits data in cleartext")
			case 80:
				vulns = append(vulns, "HTTP service - Consider using HTTPS instead")
			case 3389:
				vulns = append(vulns, "RDP service - Limit access with firewall rules")
			}
		}

		if len(openPorts) > 10 {
			vulns = append(vulns, fmt.Sprintf("High number of open ports (%d)", len(openPorts)))
		}

		// Add to network nodes
		networkMutex.Lock()
		networkNodes = append(networkNodes, NetworkNode{
			IPAddress:       host.IPAddress,
			Hostname:        host.Hostname,
			DeviceType:      "computer",
			OpenPorts:       openPorts,
			Vulnerabilities: vulns,
		})
		networkMutex.Unlock()
	}

	progressBar.SetValue(1.0)
	dialog.ShowInformation("Scan Complete", fmt.Sprintf("Found %d hosts on the network", len(networkNodes)), mainWindow)
}
