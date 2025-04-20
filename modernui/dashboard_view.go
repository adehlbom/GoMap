// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/dashboard_view.go
package modernui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// createDashboardContent builds the dashboard view
func createDashboardContent() fyne.CanvasObject {
	// Welcome banner with app description
	welcomeLabel := widget.NewLabelWithStyle("GoMap Network Scanner",
		fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	welcomeText := widget.NewRichTextFromMarkdown(`
**Welcome to GoMap!** An advanced network scanning and mapping tool.

This application allows you to:
- Discover active hosts on your network
- Scan ports on specific hosts
- Identify running services
- Map your network topology
	`)

	welcomeContainer := container.NewVBox(
		welcomeLabel,
		welcomeText,
	)

	// Stats cards
	hostCard := createStatCard("Hosts Discovered", strconv.Itoa(totalHostsDiscovered))
	portCard := createStatCard("Open Ports Found", strconv.Itoa(totalPortsDiscovered))

	// Format last scan time
	lastScanTimeStr := "Never"
	if !lastScanTime.IsZero() {
		lastScanTimeStr = lastScanTime.Format("2006-01-02 15:04:05")
	}
	timeCard := createStatCard("Last Scan", lastScanTimeStr)

	statsContainer := container.NewHBox(
		hostCard, portCard, timeCard,
	)

	// Quick action buttons
	quickScanBtn := widget.NewButton("Scan Local Network", func() {
		startLocalNetworkScan()
	})
	quickScanBtn.Importance = widget.HighImportance

	// Recent activity table
	recentActivityLabel := widget.NewLabelWithStyle("Recent Activity",
		fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	recentTable := widget.NewTable(
		func() (int, int) {
			return len(recentScans), 4
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(tci widget.TableCellID, co fyne.CanvasObject) {
			label := co.(*widget.Label)
			if tci.Row < len(recentScans) {
				scan := recentScans[len(recentScans)-tci.Row-1] // Show newest first

				switch tci.Col {
				case 0:
					label.SetText(scan.Timestamp.Format("15:04:05"))
				case 1:
					label.SetText(strings.Title(scan.Type))
				case 2:
					label.SetText(scan.Target)
				case 3:
					label.SetText(fmt.Sprintf("%d results", scan.ResultCount))
				}
			}
		},
	)

	// Set column widths
	recentTable.SetColumnWidth(0, 100)
	recentTable.SetColumnWidth(1, 100)
	recentTable.SetColumnWidth(2, 150)
	recentTable.SetColumnWidth(3, 100)

	// Network visualization (placeholder)
	networkVisualization := container.NewPadded(
		widget.NewLabelWithStyle("Network Visualization", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		canvas.NewText("Coming in future version", theme.ForegroundColor()),
	)

	// Layout the dashboard
	actionsContainer := container.NewHBox(layout.NewSpacer(), quickScanBtn, layout.NewSpacer())

	// Combine all elements in a scroll container
	return container.NewScroll(container.NewVBox(
		container.NewPadded(welcomeContainer),
		widget.NewSeparator(),
		container.NewPadded(statsContainer),
		widget.NewSeparator(),
		container.NewPadded(actionsContainer),
		widget.NewSeparator(),
		recentActivityLabel,
		container.NewVBox(
			createTableHeader([]string{"Time", "Type", "Target", "Results"}),
			container.NewVScroll(recentTable),
		),
		widget.NewSeparator(),
		container.NewPadded(networkVisualization),
	))
}
