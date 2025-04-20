// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/modern_interface.go
package modernui

import (
	"fyne.io/fyne/v2"
)

// CreateModernUI is the main entry point for creating the modern UI
// This function is called from the main package
func CreateModernUI(window fyne.Window) fyne.CanvasObject {
	// Store the window reference for dialogs and other window operations
	MainWindow = window

	// Apply the custom theme for a modern look
	window.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		// Global keyboard shortcuts could be implemented here
	})

	// Create and return the modern UI
	return createModernGUI()
}
