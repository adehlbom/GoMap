// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/modern_app.go
package modernui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
)

// Export main window so it can be accessed from main package
var MainWindow fyne.Window

// LaunchModernApp creates and runs the modern UI version of the GoMap application
func LaunchModernApp() {
	goMap := app.New()

	// Apply the custom theme for a modern look
	goMap.Settings().SetTheme(NewCustomTheme())

	// Create main window
	MainWindow = goMap.NewWindow("GoMap - Network Scanner v2.0")
	MainWindow.Resize(fyne.NewSize(1024, 768))

	// Create the UI with the modern UI function
	content := createModernGUI()

	MainWindow.SetContent(content)
	MainWindow.CenterOnScreen()
	MainWindow.ShowAndRun()
}

// ShowErrorDialog displays an error message
func ShowErrorDialog(message string) {
	dialog.ShowError(fmt.Errorf(message), MainWindow)
}

// ShowInfoDialog displays an informational message
func ShowInfoDialog(title, message string) {
	dialog.ShowInformation(title, message, MainWindow)
}

// RunScan is a public API that can be called from the main package
func RunScan(ipAddress string, minPort, maxPort int) {
	selectedHostIP = ipAddress
	portRangeMin = minPort
	portRangeMax = maxPort

	// If already on port scanner view, start scan directly
	// Otherwise switch to port scanner and then start scan
	if currentNavItem != "port" {
		switchContent("port")
	}

	// Start the scan
	startPortScan()
}
