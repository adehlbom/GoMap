// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/network_view.go
package modernui

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// createNetworkScanContent builds the network/host discovery screen
func createNetworkScanContent() fyne.CanvasObject {
	// Title
	title := widget.NewLabelWithStyle("Network Scanner", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	description := widget.NewLabel("Discover active hosts on your network by scanning IP ranges")

	// Form inputs
	ipRangeEntry := widget.NewEntry()
	ipRangeEntry.SetPlaceHolder("Enter IP range (e.g., 192.168.1.1-254 or 192.168.1.0/24)")

	// Adding example button
	exampleBtn := widget.NewButton("Use Example", func() {
		ipRangeEntry.SetText("192.168.1.0/24")
	})
	exampleBtn.Importance = widget.LowImportance

	// Host discovery options
	methodOptions := []string{
		"TCP Connect (default)",
		"ICMP Ping (requires admin/root)",
		"Combined (more accurate, slower)",
	}

	methodSelect := widget.NewSelect(methodOptions, func(selected string) {
		scanConfig["discoveryMethod"] = selected
	})
	methodSelect.SetSelected("TCP Connect (default)")

	// Timeout
	timeoutValue := widget.NewLabel("5 seconds")
	timeoutSlider := widget.NewSlider(1, 10)
	timeoutSlider.Step = 1
	timeoutSlider.SetValue(5)
	timeoutSlider.OnChanged = func(value float64) {
		timeoutValue.SetText(fmt.Sprintf("%.0f seconds", value))
		timeoutDuration = time.Duration(value) * time.Second
	}

	// Scan button
	scanButton := widget.NewButton("Scan Network", func() {
		if len(ipRangeEntry.Text) > 0 {
			startIPRange, endIPRange, err := parseIPRange(ipRangeEntry.Text)
			if err != nil {
				ShowErrorDialog("Invalid IP range: " + err.Error())
				return
			}
			startHostDiscovery(startIPRange, endIPRange)
		} else {
			ShowErrorDialog("Please enter an IP range")
		}
	})
	scanButton.Importance = widget.HighImportance

	// Local network scan button
	localScanButton := widget.NewButton("Scan Local Network", func() {
		startLocalNetworkScan()
	})

	// Filter
	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("Filter results (by IP, hostname, or status)")
	filterEntry.OnChanged = func(text string) {
		// Will implement filtering logic when we populate host results
		filterHostResults(text)
	}

	// Results table
	resultsLabel := widget.NewLabelWithStyle("Discovered Hosts", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Create table headers
	headers := []string{"IP Address", "Hostname", "Status", "Ports", "Actions"}
	tableHeader := createTableHeader(headers)

	// Host results table
	hostResultsTable = widget.NewTable(
		func() (int, int) {
			return len(hostResults), 5
		},
		func() fyne.CanvasObject {
			detailsBtn := widget.NewButton("Details", func() {})
			detailsBtn.Importance = widget.MediumImportance
			return container.NewMax(widget.NewLabel(""), detailsBtn)
		},
		func(tci widget.TableCellID, co fyne.CanvasObject) {
			if tci.Row < len(hostResults) {
				host := hostResults[tci.Row]

				switch tci.Col {
				case 0: // IP
					label := widget.NewLabel(host.IPAddress)
					container.NewMax(co).Objects[0] = label
				case 1: // Hostname
					label := widget.NewLabel(host.Hostname)
					container.NewMax(co).Objects[0] = label
				case 2: // Status
					status := widget.NewLabel(host.Status)
					container.NewMax(co).Objects[0] = status
				case 3: // Open ports
					ports := widget.NewLabel(strconv.Itoa(host.OpenPorts))
					container.NewMax(co).Objects[0] = ports
				case 4: // Actions button
					btn := widget.NewButton("Details", func() {
						// Store selected host and switch to host details view
						selectedHostIP = host.IPAddress
						switchContent("host")
						updateHostDetails(host)
					})
					btn.Importance = widget.MediumImportance
					container.NewMax(co).Objects[0] = btn
				}
			}
		},
	)

	// Set column widths
	hostResultsTable.SetColumnWidth(0, 120)
	hostResultsTable.SetColumnWidth(1, 150)
	hostResultsTable.SetColumnWidth(2, 100)
	hostResultsTable.SetColumnWidth(3, 80)
	hostResultsTable.SetColumnWidth(4, 100)

	// Export button
	exportButton := widget.NewButton("Export Results", func() {
		if len(hostResults) > 0 {
			exportHostResults()
		} else {
			ShowErrorDialog("No results to export")
		}
	})

	// Layout the form elements
	ipRangeContainer := container.NewBorder(nil, nil, nil, exampleBtn, ipRangeEntry)

	form := container.NewVBox(
		container.NewPadded(container.NewVBox(
			title,
			description,
			widget.NewSeparator(),
		)),
		container.NewPadded(container.NewGridWithColumns(2,
			widget.NewLabel("IP Range:"),
			ipRangeContainer,
			widget.NewLabel("Discovery Method:"),
			methodSelect,
			widget.NewLabel("Timeout:"),
			container.NewBorder(nil, nil, nil, timeoutValue, timeoutSlider),
		)),
		container.NewPadded(container.NewHBox(
			layout.NewSpacer(),
			scanButton,
			localScanButton,
			layout.NewSpacer(),
		)),
	)

	// Results area
	resultsArea := container.NewVBox(
		container.NewBorder(resultsLabel, nil, nil, filterEntry),
		widget.NewSeparator(),
		tableHeader,
		container.NewVScroll(hostResultsTable),
		container.NewHBox(layout.NewSpacer(), exportButton),
	)

	// Combine form and results
	splitContainer := container.NewVSplit(
		form,
		resultsArea,
	)
	splitContainer.Offset = 0.4

	return splitContainer
}

// filterHostResults filters the host results based on search text
func filterHostResults(searchText string) {
	// This would be implemented to filter the displayed host results
	// based on the search text
}

// updateHostDetails updates the host details view with the selected host information
func updateHostDetails(host HostResult) {
	// This will be implemented in the host details view
	// It will populate the host details view with information about the selected host
}

// parseIPRange parses an IP range string into start and end IP addresses
func parseIPRange(ipRange string) (string, string, error) {
	// This is a stub that would be implemented to parse IP ranges
	// For now, we'll return a placeholder to allow compilation
	return "192.168.1.1", "192.168.1.254", nil
}
