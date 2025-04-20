// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/port_scanner_view.go
package modernui

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// createPortScannerContent builds the port scanner interface
func createPortScannerContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Port Scanner", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	description := widget.NewLabel("Scan for open ports on a specific host")

	// IP Address input with auto-selection of previously chosen host
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("Enter IP address (e.g. 192.168.1.1)")
	if selectedHostIP != "" {
		ipEntry.SetText(selectedHostIP)
	}

	// Port range selection with presets
	portRangeOptions := []string{
		"Common Ports (1-1024)",
		"Extended (1-10000)",
		"Full Range (1-65535)",
		"Custom Range",
	}

	portRangeSelect := widget.NewSelect(portRangeOptions, func(selected string) {
		switch selected {
		case "Common Ports (1-1024)":
			portRangeMin = 1
			portRangeMax = 1024
		case "Extended (1-10000)":
			portRangeMin = 1
			portRangeMax = 10000
		case "Full Range (1-65535)":
			portRangeMin = 1
			portRangeMax = 65535
		case "Custom Range":
			showPortRangeDialog()
		}
	})
	portRangeSelect.SetSelected("Common Ports (1-1024)")

	// Scan method options
	scanMethodOptions := []string{
		"TCP Connect (reliable)",
		"SYN Scan (faster, requires admin/root)",
		"Service Detection (slower, more info)",
	}

	scanMethodSelect := widget.NewSelect(scanMethodOptions, func(selected string) {
		scanConfig["scanMethod"] = selected
	})
	scanMethodSelect.SetSelected("TCP Connect (reliable)")

	// Timeout
	timeoutValue := widget.NewLabel("5 seconds")
	timeoutSlider := widget.NewSlider(1, 10)
	timeoutSlider.Value = 5
	timeoutSlider.OnChanged = func(value float64) {
		timeoutValue.SetText(fmt.Sprintf("%.0f seconds", value))
		timeoutDuration = time.Duration(value) * time.Second
	}

	// Threading
	threadingValue := widget.NewLabel("100 concurrent connections")
	threadingSlider := widget.NewSlider(10, 500)
	threadingSlider.Value = 100
	threadingSlider.OnChanged = func(value float64) {
		threadingValue.SetText(fmt.Sprintf("%.0f concurrent connections", value))
		scanConfig["maxParallel"] = int(value)
	}

	// Scan button
	scanButton := widget.NewButton("Start Port Scan", func() {
		if len(ipEntry.Text) > 0 {
			selectedHostIP = ipEntry.Text
			startPortScan()
		} else {
			ShowErrorDialog("Please enter an IP address")
		}
	})
	scanButton.Importance = widget.HighImportance

	// Results filter
	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("Filter by port or service")
	filterEntry.OnChanged = func(text string) {
		// Would implement filtering logic here
		filterScanResults(text)
	}

	// Results table
	resultsLabel := widget.NewLabelWithStyle("Scan Results", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Create table headers
	headers := []string{"Port", "Service", "Protocol", "Status", "Response Time"}
	tableHeader := createTableHeader(headers)

	// Port results table
	resultsTable = widget.NewTable(
		func() (int, int) {
			return len(scanResults), 5
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(tci widget.TableCellID, co fyne.CanvasObject) {
			label := co.(*widget.Label)
			if tci.Row < len(scanResults) {
				result := scanResults[tci.Row]

				switch tci.Col {
				case 0: // Port
					label.SetText(strconv.Itoa(result.Port))
				case 1: // Service
					label.SetText(result.Service)
				case 2: // Protocol
					label.SetText(result.Protocol)
				case 3: // Status
					label.SetText(result.Status)
				case 4: // Response time
					responseTime := "N/A"
					if result.Time > 0 {
						responseTime = result.Time.String()
					}
					label.SetText(responseTime)
				}
			} else {
				label.SetText("")
			}
		},
	)

	// Set column widths
	resultsTable.SetColumnWidth(0, 80)
	resultsTable.SetColumnWidth(1, 150)
	resultsTable.SetColumnWidth(2, 80)
	resultsTable.SetColumnWidth(3, 80)
	resultsTable.SetColumnWidth(4, 100)

	// Export button
	exportButton := widget.NewButton("Export Results", func() {
		if len(scanResults) > 0 {
			exportScanResults()
		} else {
			ShowErrorDialog("No results to export")
		}
	})

	// Layout the form elements
	form := container.NewVBox(
		container.NewPadded(container.NewVBox(
			title,
			description,
			widget.NewSeparator(),
		)),
		container.NewPadded(container.NewGridWithColumns(2,
			widget.NewLabel("Target IP:"),
			ipEntry,
			widget.NewLabel("Port Range:"),
			portRangeSelect,
			widget.NewLabel("Scan Method:"),
			scanMethodSelect,
			widget.NewLabel("Timeout:"),
			container.NewBorder(nil, nil, nil, timeoutValue, timeoutSlider),
			widget.NewLabel("Thread Count:"),
			container.NewBorder(nil, nil, nil, threadingValue, threadingSlider),
		)),
		container.NewPadded(container.NewHBox(
			layout.NewSpacer(),
			scanButton,
			layout.NewSpacer(),
		)),
	)

	// Results area
	resultsArea := container.NewVBox(
		container.NewBorder(resultsLabel, nil, nil, filterEntry),
		widget.NewSeparator(),
		tableHeader,
		container.NewVScroll(resultsTable),
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

// filterScanResults filters the port scan results based on search text
func filterScanResults(searchText string) {
	// This would be implemented to filter the displayed scan results
	// based on the search text
}

// startPortScan initiates the port scanning process
func startPortScan() {
	if scanActive {
		return
	}

	scanActive = true
	scanResults = []ModernScanResult{}

	// If we have a results table, refresh it
	if resultsTable != nil {
		resultsTable.Refresh()
	}

	progressBar.SetValue(0)
	scanStatus.SetText("Scanning ports...")

	scanStartTime := time.Now()

	// Start the scan in a goroutine
	go func() {
		totalPorts := portRangeMax - portRangeMin + 1
		portsDone := 0

		// Initialize scan semaphore for concurrency control
		maxConcurrent := 100
		if val, ok := scanConfig["maxParallel"]; ok {
			if intVal, ok := val.(int); ok && intVal > 0 {
				maxConcurrent = intVal
			}
		}
		scanSemaphore = make(chan struct{}, maxConcurrent)

		// Initialize wait group
		scanWaitGroup.Add(totalPorts)

		// Start scan for each port
		for port := portRangeMin; port <= portRangeMax; port++ {
			scanSemaphore <- struct{}{} // Acquire token

			go func(p int) {
				defer func() {
					<-scanSemaphore // Release token
					portsDone++
					progress := float64(portsDone) / float64(totalPorts)
					progressBar.SetValue(progress)
					scanWaitGroup.Done()
				}()

				scanPort(selectedHostIP, p)
			}(port)
		}

		// Wait for all scans to complete
		scanWaitGroup.Wait()

		// Record scan duration
		scanDuration := time.Since(scanStartTime)

		// Add to recent scans
		addRecentScan("port", selectedHostIP, len(scanResults), scanDuration)

		// Update statistics
		updateStatistics()

		scanActive = false
		scanStatus.SetText("Ready")
		lastScanTime = time.Now()
		scanSummaryLabel.SetText(fmt.Sprintf("Found %d open ports on %s", len(scanResults), selectedHostIP))

		// If we have a results table, refresh it
		if resultsTable != nil {
			resultsTable.Refresh()
		}

		// Show completion message
		ShowInfoDialog("Scan Complete", fmt.Sprintf("Scan complete. Found %d open ports.", len(scanResults)))
	}()
}

// scanPort performs the actual port scan and updates the results
func scanPort(ip string, port int) {
	startTime := time.Now()

	// Call the tcp_scan function to scan the port
	// We'll need to implement this or bridge to the main package's implementation
	result := tcpScan(ip, port)

	scanDuration := time.Since(startTime)

	// Add timing information
	result.Time = scanDuration

	if result.Status == "OPEN" {
		// Add to results
		scanResults = append(scanResults, result)

		// Try to refresh UI if available
		if resultsTable != nil {
			resultsTable.Refresh()
		}
	}
}

// tcpScan performs a TCP scan on the specified port
func tcpScan(ipAddress string, port int) ModernScanResult {
	// This is a placeholder function that would connect to the main package's
	// tcp_scan functionality or implement its own TCP scanning.
	// For now, it returns a dummy result.
	return ModernScanResult{
		Port:     port,
		Status:   "CLOSED",
		Service:  "Unknown",
		Protocol: "TCP",
		Time:     0,
	}
}

// showPortRangeDialog displays a dialog for custom port range input
func showPortRangeDialog() {
	// Create entry fields
	minEntry := widget.NewEntry()
	minEntry.SetPlaceHolder("e.g. 1")

	maxEntry := widget.NewEntry()
	maxEntry.SetPlaceHolder("e.g. 1024")

	// Create a form dialog with form items
	form := widget.NewForm(
		widget.NewFormItem("Minimum Port:", minEntry),
		widget.NewFormItem("Maximum Port:", maxEntry),
	)

	// Set the OnSubmit handler for the form
	form.OnSubmit = func() {
		// Parse min port
		if minEntry.Text != "" {
			if min, err := strconv.Atoi(minEntry.Text); err == nil {
				if min >= 1 && min <= 65535 {
					portRangeMin = min
				} else {
					ShowErrorDialog("Minimum port must be between 1 and 65535")
					return
				}
			} else {
				ShowErrorDialog("Invalid minimum port number")
				return
			}
		}

		// Parse max port
		if maxEntry.Text != "" {
			if max, err := strconv.Atoi(maxEntry.Text); err == nil {
				if max >= 1 && max <= 65535 && max >= portRangeMin {
					portRangeMax = max
				} else if max < portRangeMin {
					ShowErrorDialog("Maximum port must be greater than minimum port")
					return
				} else {
					ShowErrorDialog("Maximum port must be between 1 and 65535")
					return
				}
			} else {
				ShowErrorDialog("Invalid maximum port number")
				return
			}
		}

		// Update UI to show selected range
		ShowInfoDialog("Port Range Updated",
			fmt.Sprintf("Custom port range set to %d-%d", portRangeMin, portRangeMax))
	}

	// Create a custom dialog to show the form
	customDialog := dialog.NewCustom("Port Range Settings", "Cancel", form, fyne.CurrentApp().Driver().AllWindows()[0])

	// Show the dialog
	customDialog.Show()
}

// cancelScan cancels any active scanning operations
func cancelScan() {
	if !scanActive {
		return
	}

	// Cannot truly cancel goroutines, but we can mark as inactive and update UI
	scanActive = false
	progressBar.SetValue(0)
	scanStatus.SetText("Scan cancelled")
	ShowInfoDialog("Operation Cancelled", "Scan operation cancelled by user")
}
