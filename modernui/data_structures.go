// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/data_structures.go
package modernui

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
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

// HostResult struct must match the one in types/network_types.go
type HostResult struct {
	IPAddress string  `json:"ip_address"`
	Hostname  string  `json:"hostname"`
	Status    string  `json:"status"`
	RTT       float64 `json:"rtt"`
	MAC       string  `json:"mac,omitempty"`
	Vendor    string  `json:"vendor,omitempty"`
	OpenPorts int     `json:"open_ports"`
}

// Global variables for GUI components that need to be accessed from different functions
var (
	resultsTable         *widget.Table
	hostResultsTable     *widget.Table
	progressBar          *widget.ProgressBar
	scanResults          []ModernScanResult
	hostResults          []HostResult
	portRangeMin         = 1
	portRangeMax         = 1024
	scanActive           = false
	timeoutDuration      = 5 * time.Second
	currentNavItem       = "dashboard"
	scanConfig           = make(map[string]interface{})
	mainContent          *fyne.Container
	contentContainer     *fyne.Container
	statusPanel          *fyne.Container
	scanStatus           *widget.Label
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

// Button style names defined as string constants
// These can be used for custom styling in your application
const (
	PrimaryButtonStyle   = "primary"
	SecondaryButtonStyle = "secondary"
	DangerButtonStyle    = "danger"
)

// Synchronization variables
var (
	scanWaitGroup sync.WaitGroup
	scanSemaphore chan struct{}
)
