package main

import (
	"fyne.io/fyne/v2"
)

// HalfHeightLayout is a custom layout that arranges two objects
// vertically, each taking half the available height and full width.
type HalfHeightLayout struct{}

// NewHalfHeightLayout creates a new instance of HalfHeightLayout.
func NewHalfHeightLayout() fyne.Layout {
	return &HalfHeightLayout{}
}

// Layout arranges the objects within the container.
func (h *HalfHeightLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 2 {
		// Log an error or handle the case where there aren't exactly two objects
		// For simplicity, we'll just return here.
		return
	}

	halfHeight := size.Height / 2

	// Position and size the first object
	objects[0].Resize(fyne.NewSize(size.Width, halfHeight))
	objects[0].Move(fyne.NewPos(0, 0))

	// Position and size the second object
	objects[1].Resize(fyne.NewSize(size.Width, halfHeight))
	objects[1].Move(fyne.NewPos(0, halfHeight))
}

// MinSize calculates the minimum size required for the layout.
func (h *HalfHeightLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) != 2 {
		return fyne.NewSize(0, 0) // Or a sensible default
	}

	// The minimum width is the maximum of the two objects' minimum widths.
	// The minimum height is the sum of the two objects' minimum heights.
	minWidth := fyne.Max(objects[0].MinSize().Width, objects[1].MinSize().Width)
	minHeight := objects[0].MinSize().Height + objects[1].MinSize().Height

	return fyne.NewSize(minWidth, minHeight)
}
