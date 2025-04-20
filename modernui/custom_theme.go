// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/custom_theme.go
package modernui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme is a modernized theme for GoMap with a clean dark look
type CustomTheme struct {
	regularFont, boldFont, monoFont fyne.Resource
}

// NewCustomTheme creates a new custom theme instance
func NewCustomTheme() fyne.Theme {
	return &CustomTheme{
		regularFont: theme.DefaultTheme().Font(fyne.TextStyle{}),
		boldFont:    theme.DefaultTheme().Font(fyne.TextStyle{Bold: true}),
		monoFont:    theme.DefaultTheme().Font(fyne.TextStyle{Monospace: true}),
	}
}

// Color returns the color for the specified ThemeColorName
func (t *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 18, G: 21, B: 29, A: 255} // Dark blue-gray
	case theme.ColorNameButton:
		return color.NRGBA{R: 37, G: 99, B: 235, A: 255} // Bright blue
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 60, G: 60, B: 60, A: 255}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 240, G: 240, B: 250, A: 255} // Nearly white
	case theme.ColorNameHover:
		return color.NRGBA{R: 51, G: 153, B: 255, A: 255} // Light blue
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 140, G: 140, B: 140, A: 255} // Light gray
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0, G: 109, B: 217, A: 255} // Medium blue
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 80, G: 80, B: 80, A: 255}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font returns the specified font resource
func (t *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Bold {
		return t.boldFont
	}
	if style.Monospace {
		return t.monoFont
	}
	return t.regularFont
}

// Icon returns the icon resource for a named icon
func (t *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size for a specific element/name
func (t *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 10
	case theme.SizeNameScrollBar:
		return 8
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 20
	default:
		return theme.DefaultTheme().Size(name)
	}
}
