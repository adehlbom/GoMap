// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/main.go
package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"GoMap/modernui"
)

func main() {
	// Check if user wants to use the modern UI
	useModernUI := true

	// Check for UI choice flag
	for _, arg := range os.Args {
		if arg == "--classic" || arg == "-c" {
			useModernUI = false
			break
		}
	}

	// Print banner
	fmt.Println("==========================================")
	fmt.Println("         GoMap Network Scanner            ")
	fmt.Println("==========================================")

	// Check if running in command-line mode or GUI mode
	if len(os.Args) > 1 && os.Args[1] != "--classic" && os.Args[1] != "-c" {
		// Command-line mode
		fmt.Println("Command-line mode not yet implemented in this version")
		os.Exit(1)
	} else {
		// Launch the application with the selected UI
		launchApp(useModernUI)
	}
}

// launchApp starts the application with the selected UI type
func launchApp(useModernUI bool) {
	// Create the application instance
	goMapApp := app.New()

	// Create the main window
	var mainWindow fyne.Window
	if useModernUI {
		fmt.Println("Starting with Modern UI...")
		mainWindow = goMapApp.NewWindow("GoMap - Network Scanner v2.0")
		mainWindow.Resize(fyne.NewSize(1024, 768))

		// Initialize the modern UI
		content := modernui.CreateModernUI(mainWindow)
		mainWindow.SetContent(content)
	} else {
		fmt.Println("Starting with Classic UI...")
		mainWindow = goMapApp.NewWindow("GoMap - Network Scanner")
		mainWindow.Resize(fyne.NewSize(900, 700))

		// For now, just show a message that classic UI isn't integrated yet
		label := widget.NewLabel("Classic UI not implemented in this version")
		content := container.NewCenter(label)
		mainWindow.SetContent(content)
	}

	mainWindow.CenterOnScreen()
	mainWindow.ShowAndRun()
}
