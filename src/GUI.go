// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/src/modern_GUI.go
package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ModernScanResult stores information about scanned ports with enhanced details
type ModernScanResult struct {
	Port     int
	Status   string
	Service  string
	Time     time.Duration
	Protocol string
}

// RecentScan tracks scan history
type RecentScan struct {
	Type        string    // "network", "host", "port"
	Target      string    // IP, range, etc.
	ResultCount int       // Number of hosts or ports found
	Timestamp   time.Time // When the scan occurred
	Duration    time.Duration
}

// Global variables for GUI components that need to be accessed from different functions
var (
	resultsList          *widget.Table
	hostResultsList      *widget.Table
	progressBar          *widget.ProgressBar
	scanResults          []ScanResult
	hostResults          []HostResult
	portRangeMin         = 1
	portRangeMax         = 1024
	scanActive           = false
	timeoutDuration      = 5 * time.Second
	currentNavItem       = "dashboard"
	scanConfig           = make(map[string]interface{})
	mainWindow           fyne.Window
	mainContent          *fyne.Container
	contentContainer     *fyne.Container
	statusPanel          *fyne.Container
	currentScanStatus    *widget.Label
	selectedHostIP       string
	scanSummaryLabel     *widget.Label
	lastScanTime         time.Time
	recentScans          []RecentScan
	totalHostsDiscovered int
	totalPortsDiscovered int
	appName              = "GoMap"
	appVersion           = "2.0"
	resourcePath         string
)

// Button style functions for consistent UI
func applyPrimaryStyle(button *widget.Button) *widget.Button {
	button.Importance = widget.HighImportance
	return button
}

func applySecondaryStyle(button *widget.Button) *widget.Button {
	button.Importance = widget.MediumImportance
	return button
}

func applyDangerStyle(button *widget.Button) *widget.Button {
	button.Importance = widget.DangerImportance
	return button
}

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

	currentScanStatus = widget.NewLabel("Ready")
	scanSummaryLabel = widget.NewLabel("No scans performed")

	return container.NewVBox(
		progressBar,
		container.NewHBox(
			widget.NewIcon(theme.InfoIcon()),
			widget.NewLabel("Status:"),
			currentScanStatus,
			layout.NewSpacer(),
			scanSummaryLabel,
		),
	)
}

// createSidebar creates the left sidebar navigation
func createSidebar() *fyne.Container {
	// App title
	appTitle := widget.NewLabelWithStyle(appName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	appVersion := widget.NewLabelWithStyle("Version "+appVersion, fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

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
		appVersion,
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
			showErrorDialog("No results to export")
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
				showErrorDialog("Invalid IP range: " + err.Error())
				return
			}
			startHostDiscovery(startIPRange, endIPRange, timeoutDuration)
		} else {
			showErrorDialog("Please enter an IP range")
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
		// Would implement filtering logic here
	}

	// Results table
	resultsLabel := widget.NewLabelWithStyle("Discovered Hosts", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Create table headers
	headers := []string{"IP Address", "Hostname", "Status", "Ports", "Actions"}
	tableHeader := createTableHeader(headers)

	// Host results table
	hostResultsList = widget.NewTable(
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
	hostResultsList.SetColumnWidth(0, 120)
	hostResultsList.SetColumnWidth(1, 150)
	hostResultsList.SetColumnWidth(2, 100)
	hostResultsList.SetColumnWidth(3, 80)
	hostResultsList.SetColumnWidth(4, 100)

	// Export button
	exportButton := widget.NewButton("Export Results", func() {
		if len(hostResults) > 0 {
			exportHostResults()
		} else {
			showErrorDialog("No results to export")
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
		container.NewVScroll(hostResultsList),
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

// createHostDetailsContent builds the host details view
func createHostDetailsContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Host Details", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Host details section
	hostIPLabel := widget.NewLabel("IP: Not selected")
	hostNameLabel := widget.NewLabel("Hostname: Not selected")
	hostStatusLabel := widget.NewLabel("Status: Unknown")
	lastSeenLabel := widget.NewLabel("Last seen: Never")

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
			// Auto-populate the port scanner with this host
			startScan(selectedHostIP, portRangeMin, portRangeMax)
		} else {
			showErrorDialog("No host selected")
		}
	})
	scanPortsButton.Importance = widget.HighImportance

	identifyOSButton := widget.NewButton("Identify OS", func() {
		if selectedHostIP != "" {
			// Would implement OS detection logic here
			showInfoDialog("OS Detection", "OS Detection not implemented yet")
		} else {
			showErrorDialog("No host selected")
		}
	})

	traceRouteButton := widget.NewButton("Trace Route", func() {
		if selectedHostIP != "" {
			showInfoDialog("Trace Route", "Trace route not implemented yet")
		} else {
			showErrorDialog("No host selected")
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

	openPortsList := widget.NewTable(
		func() (int, int) {
			return 0, 4 // Will be populated with ports from selected host
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(tci widget.TableCellID, co fyne.CanvasObject) {
			label := co.(*widget.Label)
			label.SetText("No port data") // Set a placeholder text until real port data is available
			// This would be populated with ports from selected host
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
			startScan(ipEntry.Text, portRangeMin, portRangeMax)
		} else {
			showErrorDialog("Please enter an IP address")
		}
	})
	scanButton.Importance = widget.HighImportance

	// Results filter
	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("Filter by port or service")
	filterEntry.OnChanged = func(text string) {
		// Would implement filtering logic here
	}

	// Results table
	resultsLabel := widget.NewLabelWithStyle("Scan Results", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Create table headers
	headers := []string{"Port", "Service", "Protocol", "Status", "Response Time"}
	tableHeader := createTableHeader(headers)

	// Port results table
	resultsList = widget.NewTable(
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
					label.SetText("TCP") // Currently only TCP is supported
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
	resultsList.SetColumnWidth(0, 80)
	resultsList.SetColumnWidth(1, 150)
	resultsList.SetColumnWidth(2, 80)
	resultsList.SetColumnWidth(3, 80)
	resultsList.SetColumnWidth(4, 100)

	// Export button
	exportButton := widget.NewButton("Export Results", func() {
		if len(scanResults) > 0 {
			exportScanResults()
		} else {
			showErrorDialog("No results to export")
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
		container.NewVScroll(resultsList),
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

// createSettingsContent builds the settings interface
func createSettingsContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Create form items
	parallelValue := widget.NewLabel("100")
	parallelSlider := widget.NewSlider(10, 500)
	parallelSlider.Step = 10
	parallelSlider.Value = 100
	parallelSlider.OnChanged = func(value float64) {
		parallelValue.SetText(fmt.Sprintf("%.0f", value))
		scanConfig["maxParallel"] = int(value)
	}

	timeoutValue := widget.NewLabel("5 sec")
	timeoutSlider := widget.NewSlider(1, 20)
	timeoutSlider.Step = 1
	timeoutSlider.Value = 5
	timeoutSlider.OnChanged = func(value float64) {
		timeoutValue.SetText(fmt.Sprintf("%.0f sec", value))
		timeoutDuration = time.Duration(value) * time.Second
	}

	// Visual theme
	themeSelect := widget.NewSelect([]string{
		"System Default",
		"Dark Theme",
		"Light Theme",
	}, func(theme string) {
		scanConfig["appTheme"] = theme
		// Would implement theme change logic
	})
	themeSelect.SetSelected("Dark Theme")

	// Port scanning behavior
	portScanBehavior := widget.NewRadioGroup([]string{
		"Fast (less accurate)",
		"Balance speed and accuracy",
		"Thorough (slower)",
	}, func(option string) {
		scanConfig["portScanBehavior"] = option
	})
	portScanBehavior.Selected = "Balance speed and accuracy"

	// Host discovery behavior
	hostDiscoveryBehavior := widget.NewRadioGroup([]string{
		"Quick ping only",
		"Ping and common ports (default)",
		"Extensive port checking",
	}, func(option string) {
		scanConfig["hostDiscoveryBehavior"] = option
	})
	hostDiscoveryBehavior.Selected = "Ping and common ports (default)"

	// Banner grabbing settings
	bannerGrabCheckbox := widget.NewCheck("Enable banner grabbing", func(value bool) {
		scanConfig["enableBannerGrab"] = value
	})
	bannerGrabCheckbox.Checked = true

	// Service info grabbing settings
	serviceInfoCheckbox := widget.NewCheck("Enable service detection", func(value bool) {
		scanConfig["enableServiceDetection"] = value
	})
	serviceInfoCheckbox.Checked = true

	// Save and restore defaults buttons
	saveButton := widget.NewButton("Save Settings", func() {
		// Would implement settings save logic
		showInfoDialog("Settings", "Settings saved successfully")
	})
	saveButton.Importance = widget.HighImportance

	defaultsButton := widget.NewButton("Restore Defaults", func() {
		// Would implement defaults restore logic
		portScanBehavior.SetSelected("Balance speed and accuracy")
		hostDiscoveryBehavior.SetSelected("Ping and common ports (default)")
		parallelSlider.SetValue(100)
		timeoutSlider.SetValue(5)
		bannerGrabCheckbox.SetChecked(true)
		serviceInfoCheckbox.SetChecked(true)
		themeSelect.SetSelected("Dark Theme")

		showInfoDialog("Settings", "Default settings restored")
	})

	// Build settings form
	settingsForm := widget.NewForm(
		widget.NewFormItem("Max Parallel Scans",
			container.NewBorder(nil, nil, nil, parallelValue, parallelSlider)),
		widget.NewFormItem("Default Timeout",
			container.NewBorder(nil, nil, nil, timeoutValue, timeoutSlider)),
		widget.NewFormItem("Application Theme", themeSelect),
	)

	// Add scanning behavior options
	scanBehaviorBox := container.NewVBox(
		widget.NewLabel("Port Scanning Behavior:"),
		portScanBehavior,
		widget.NewSeparator(),
		widget.NewLabel("Host Discovery Behavior:"),
		hostDiscoveryBehavior,
	)

	// Add feature toggles
	featureToggles := container.NewVBox(
		widget.NewLabelWithStyle("Feature Toggles", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		bannerGrabCheckbox,
		serviceInfoCheckbox,
	)

	// Add buttons
	buttonBox := container.NewHBox(layout.NewSpacer(), saveButton, defaultsButton, layout.NewSpacer())

	// About section
	aboutBox := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("About GoMap", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Version 2.0"),
		widget.NewLabel("© 2025 | https://github.com/adehlbom/GoMap"),
	)

	// Combine all settings sections
	settingsLayout := container.NewVBox(
		container.NewPadded(container.NewVBox(
			title,
			widget.NewSeparator(),
		)),
		container.NewPadded(settingsForm),
		widget.NewSeparator(),
		container.NewPadded(scanBehaviorBox),
		widget.NewSeparator(),
		container.NewPadded(featureToggles),
		container.NewPadded(buttonBox),
		container.NewPadded(aboutBox),
	)

	return container.NewScroll(settingsLayout)
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

// Custom layout for fixed width sidebar
type fixedWidthLayout struct {
	width float32
}

func (f *fixedWidthLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(0, 0)
	size := fyne.NewSize(f.width, containerSize.Height)

	for _, o := range objects {
		o.Resize(size)
		o.Move(pos)
	}
}

func (f *fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minHeight := float32(0)
	for _, o := range objects {
		if h := o.MinSize().Height; h > minHeight {
			minHeight = h
		}
	}
	return fyne.NewSize(f.width, minHeight)
}
