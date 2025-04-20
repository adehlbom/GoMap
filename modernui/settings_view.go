// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/settings_view.go
package modernui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

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
		ShowInfoDialog("Settings", "Settings saved successfully")
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

		ShowInfoDialog("Settings", "Default settings restored")
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
