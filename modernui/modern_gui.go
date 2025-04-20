// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/modern_gui.go
package modernui

import (
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// createModernGUI creates the main GUI with a sidebar navigation and content area
func createModernGUI() fyne.CanvasObject {
	// Initialize the resource path to load icons
	resourcePath = filepath.Join("resources", "icons")

	// Initialize app statistics
	if len(recentScans) == 0 {
		recentScans = make([]RecentScan, 0)
	}

	// Create the sidebar navigation
	sidebar := createSidebar()

	// Create initial content (dashboard)
	contentContainer = container.NewMax(createDashboardContent())

	// Create status bar
	statusPanel = createStatusPanel()

	// Create main layout with sidebar and content
	mainContent = container.NewBorder(
		nil,
		statusPanel,
		sidebar,
		nil,
		contentContainer,
	)

	return mainContent
}

// createStatusPanel creates the status bar at the bottom of the window
func createStatusPanel() *fyne.Container {
	progressBar = widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 1

	scanStatus = widget.NewLabel("Ready")
	scanSummaryLabel = widget.NewLabel("No scans performed")

	return container.NewVBox(
		progressBar,
		container.NewHBox(
			widget.NewIcon(theme.InfoIcon()),
			widget.NewLabel("Status:"),
			scanStatus,
			layout.NewSpacer(),
			scanSummaryLabel,
		),
	)
}

// createSidebar creates the left sidebar navigation
func createSidebar() *fyne.Container {
	// App title
	appTitle := widget.NewLabelWithStyle(appName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	appVersionLabel := widget.NewLabelWithStyle("Version "+appVersion, fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	// Navigation items
	navItems := []struct {
		icon     fyne.Resource
		title    string
		id       string
		onTapped func()
	}{
		{theme.HomeIcon(), "Dashboard", "dashboard", func() { switchContent("dashboard") }},
		{theme.SearchIcon(), "Network Scan", "network", func() { switchContent("network") }},
		{theme.ComputerIcon(), "Host Details", "host", func() { switchContent("host") }},
		{theme.ViewRestoreIcon(), "Port Scanner", "port", func() { switchContent("port") }},
		{theme.SettingsIcon(), "Settings", "settings", func() { switchContent("settings") }},
	}

	// Create navigation buttons
	var navButtons []fyne.CanvasObject

	// Logo and title area
	titleArea := container.NewVBox(
		appTitle,
		appVersionLabel,
		widget.NewSeparator(),
	)

	// Build navigation items
	for _, item := range navItems {
		btn := widget.NewButtonWithIcon(item.title, item.icon, item.onTapped)
		btn.Alignment = widget.ButtonAlignLeading

		// Add to navigation list
		navButtons = append(navButtons, btn)
	}

	// Actions section
	actionsLabel := widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	quickScanBtn := widget.NewButtonWithIcon("Quick Scan", theme.MediaPlayIcon(), func() {
		startLocalNetworkScan()
	})

	helpBtn := widget.NewButtonWithIcon("Help", theme.HelpIcon(), func() {
		showHelp()
	})

	exportBtn := widget.NewButtonWithIcon("Export Results", theme.DocumentSaveIcon(), func() {
		if currentNavItem == "port" && len(scanResults) > 0 {
			exportScanResults()
		} else if (currentNavItem == "network" || currentNavItem == "host") && len(hostResults) > 0 {
			exportHostResults()
		} else {
			ShowErrorDialog("No results to export")
		}
	})

	cancelBtn := widget.NewButtonWithIcon("Cancel Scan", theme.CancelIcon(), func() {
		if scanActive {
			cancelScan()
		}
	})
	cancelBtn.Importance = widget.DangerImportance

	// Create action buttons section
	actionBtns := container.NewVBox(
		widget.NewSeparator(),
		actionsLabel,
		quickScanBtn,
		exportBtn,
		cancelBtn,
		helpBtn,
	)

	// Build the sidebar with fixed width
	navBox := container.NewVBox(append([]fyne.CanvasObject{titleArea}, navButtons...)...)
	sidebar := container.NewBorder(navBox, actionBtns, nil, nil, layout.NewSpacer())

	// Use a fixed width for the sidebar
	sidebarContainer := container.New(
		&fixedWidthLayout{width: 200},
		sidebar,
	)

	return sidebarContainer
}

// switchContent changes the main content area based on selected navigation item
func switchContent(contentID string) {
	currentNavItem = contentID

	var content fyne.CanvasObject

	switch contentID {
	case "dashboard":
		content = createDashboardContent()
	case "network":
		content = createNetworkScanContent()
	case "host":
		content = createHostDetailsContent()
	case "port":
		content = createPortScannerContent()
	case "settings":
		content = createSettingsContent()
	default:
		content = createDashboardContent()
	}

	contentContainer.Objects = []fyne.CanvasObject{content}
	contentContainer.Refresh()
}

// Helper functions

// createStatCard creates a card-like UI element to display statistics
func createStatCard(title string, value string) fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	valueLabel := widget.NewLabelWithStyle(value, fyne.TextAlignCenter, fyne.TextStyle{})

	card := container.NewVBox(
		titleLabel,
		valueLabel,
	)

	// Wrap in a styled container to look like a card
	return container.NewPadded(container.NewVBox(
		card,
	))
}

// createTableHeader creates a consistent header row for tables
func createTableHeader(columns []string) fyne.CanvasObject {
	var headers []fyne.CanvasObject

	for _, col := range columns {
		headers = append(headers, widget.NewLabelWithStyle(col, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	}

	return container.NewHBox(headers...)
}

// Utility Functions

// addRecentScan adds a new scan to the recent scans list
func addRecentScan(scanType string, target string, resultCount int, duration time.Duration) {
	recentScans = append(recentScans, RecentScan{
		Type:        scanType,
		Target:      target,
		ResultCount: resultCount,
		Timestamp:   time.Now(),
		Duration:    duration,
	})

	// Limit to most recent 10 scans
	if len(recentScans) > 10 {
		recentScans = recentScans[len(recentScans)-10:]
	}
}

// updateStatistics updates the overall application statistics
func updateStatistics() {
	// Update total hosts discovered
	totalHostsDiscovered = len(hostResults)

	// Count total open ports across all hosts
	totalPortsDiscovered = len(scanResults)

	// If on the dashboard, refresh it to show updated stats
	if currentNavItem == "dashboard" {
		switchContent("dashboard")
	}
}
