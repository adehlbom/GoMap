// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/host_details_view.go
package modernui

import (
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// createHostDetailsContent builds the host details view
func createHostDetailsContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Host Details", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Host details section
	hostIPLabel := widget.NewLabel("IP: Not selected")
	hostNameLabel := widget.NewLabel("Hostname: Not selected")
	hostStatusLabel := widget.NewLabel("Status: Unknown")
	lastSeenLabel := widget.NewLabel("Last seen: Never")

	// Set values if we have a selected host
	if selectedHostIP != "" {
		// Find the host in our results
		for _, host := range hostResults {
			if host.IPAddress == selectedHostIP {
				hostIPLabel.SetText("IP: " + host.IPAddress)
				hostNameLabel.SetText("Hostname: " + host.Hostname)
				hostStatusLabel.SetText("Status: " + host.Status)
				lastSeenLabel.SetText("Last seen: " + time.Now().Format("2006-01-02 15:04:05"))
				break
			}
		}
	}

	// Create info cards panel
	infoPanel := container.NewVBox(
		container.NewHBox(
			widget.NewIcon(theme.ComputerIcon()),
			hostIPLabel,
		),
		container.NewHBox(
			widget.NewIcon(theme.InfoIcon()),
			hostNameLabel,
		),
		container.NewHBox(
			widget.NewIcon(theme.VisibilityIcon()),
			hostStatusLabel,
		),
		container.NewHBox(
			widget.NewIcon(theme.HistoryIcon()),
			lastSeenLabel,
		),
	)

	// Action buttons
	scanPortsButton := widget.NewButton("Scan Ports", func() {
		if selectedHostIP != "" {
			switchContent("port")
			// Starts port scan automatically for this host
			startPortScan()
		} else {
			ShowErrorDialog("No host selected")
		}
	})
	scanPortsButton.Importance = widget.HighImportance

	identifyOSButton := widget.NewButton("Identify OS", func() {
		if selectedHostIP != "" {
			// Would implement OS detection logic here
			ShowInfoDialog("OS Detection", "OS Detection not implemented yet")
		} else {
			ShowErrorDialog("No host selected")
		}
	})

	traceRouteButton := widget.NewButton("Trace Route", func() {
		if selectedHostIP != "" {
			ShowInfoDialog("Trace Route", "Trace route not implemented yet")
		} else {
			ShowErrorDialog("No host selected")
		}
	})

	// Action buttons container
	actionButtons := container.NewHBox(
		layout.NewSpacer(),
		scanPortsButton,
		identifyOSButton,
		traceRouteButton,
		layout.NewSpacer(),
	)

	// Open ports section
	openPortsLabel := widget.NewLabelWithStyle("Open Ports", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Create port table
	headers := []string{"Port", "Service", "Protocol", "Status"}
	tableHeader := createTableHeader(headers)

	// Get open ports for this host from scan results
	var hostPorts []ModernScanResult
	if selectedHostIP != "" {
		for _, result := range scanResults {
			if result.Status == "OPEN" {
				hostPorts = append(hostPorts, result)
			}
		}
	}

	openPortsList := widget.NewTable(
		func() (int, int) {
			return len(hostPorts), 4
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(tci widget.TableCellID, co fyne.CanvasObject) {
			label := co.(*widget.Label)

			if tci.Row < len(hostPorts) {
				port := hostPorts[tci.Row]

				switch tci.Col {
				case 0:
					label.SetText(strconv.Itoa(port.Port))
				case 1:
					label.SetText(port.Service)
				case 2:
					label.SetText(port.Protocol)
				case 3:
					label.SetText(port.Status)
				}
			} else {
				label.SetText("")
			}
		},
	)

	// Set column widths
	openPortsList.SetColumnWidth(0, 80)
	openPortsList.SetColumnWidth(1, 150)
	openPortsList.SetColumnWidth(2, 80)
	openPortsList.SetColumnWidth(3, 80)

	// Vulnerability section (placeholder for future)
	vulnLabel := widget.NewLabelWithStyle("Potential Vulnerabilities", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	vulnList := widget.NewLabel("No vulnerability scan performed")

	vulnContainer := container.NewVBox(
		vulnLabel,
		container.NewPadded(vulnList),
	)

	// Combine all elements
	detailsLayout := container.NewVBox(
		container.NewPadded(container.NewVBox(
			title,
			widget.NewSeparator(),
		)),
		container.NewPadded(infoPanel),
		container.NewPadded(actionButtons),
		widget.NewSeparator(),
		container.NewPadded(openPortsLabel),
		tableHeader,
		container.NewVScroll(openPortsList),
		widget.NewSeparator(),
		container.NewPadded(vulnContainer),
	)

	return container.NewScroll(detailsLayout)
}
