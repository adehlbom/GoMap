// filepath: /Users/andersdehlbom/Coding/Privat/GoMap/modernui/custom_layout.go
package modernui

import (
	"fyne.io/fyne/v2"
)

// Custom layouts for the modernized UI

// fixedWidthLayout implements a layout with a fixed width container
type fixedWidthLayout struct {
	width float32
}

// Layout positions the objects within the specified container size
func (f *fixedWidthLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(0, 0)
	size := fyne.NewSize(f.width, containerSize.Height)

	for _, o := range objects {
		o.Resize(size)
		o.Move(pos)
	}
}

// MinSize determines the smallest size that satisfies all objects
func (f *fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minHeight := float32(0)
	for _, o := range objects {
		if h := o.MinSize().Height; h > minHeight {
			minHeight = h
		}
	}
	return fyne.NewSize(f.width, minHeight)
}

// cardLayout is a custom layout for card-like UI elements
type cardLayout struct {
	padding float32
}

// Layout arranges the objects with padding in a card-like container
func (c *cardLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(c.padding, c.padding)
	size := fyne.NewSize(
		containerSize.Width-(2*c.padding),
		containerSize.Height-(2*c.padding),
	)

	for _, o := range objects {
		o.Resize(size)
		o.Move(pos)
	}
}

// MinSize determines the smallest size that satisfies all objects with padding
func (c *cardLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minW, minH := float32(0), float32(0)

	for _, o := range objects {
		childSize := o.MinSize()
		if childSize.Width > minW {
			minW = childSize.Width
		}
		if childSize.Height > minH {
			minH = childSize.Height
		}
	}

	return fyne.NewSize(minW+(2*c.padding), minH+(2*c.padding))
}

// newCardContainer creates a container with the card layout and styling
func newCardContainer(content fyne.CanvasObject) *fyne.Container {
	return fyne.NewContainer(content)
}
