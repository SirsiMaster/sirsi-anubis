// Package main — pantheon-menubar
//
// icon.go — Embedded icon for the macOS menu bar.
//
// macOS menu bar icons should be "template images" — monochrome with alpha.
// The system tints them white on dark backgrounds and black on light.
// Size: 22×22 pixels is the standard for @1x, 44×44 for @2x.
package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// generateAnkhIcon creates a 22x22 ankh symbol as a macOS template icon.
// Uses solid black pixels on transparent — macOS will handle the tinting.
func generateAnkhIcon() []byte {
	const size = 22
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	// Ankh symbol (☥) — drawn as a pixel bitmap
	// The ankh has: a loop/oval at top, a crossbar, and a vertical stem
	black := color.NRGBA{0, 0, 0, 255}

	// Define the ankh shape row by row (y from top)
	// Each row is a list of x-coordinates that should be filled
	rows := map[int][]int{
		// Top of loop
		1: {9, 10, 11, 12},
		2: {7, 8, 13, 14},
		3: {6, 7, 14, 15},
		4: {6, 15},
		5: {6, 15},
		6: {6, 7, 14, 15},
		7: {7, 8, 13, 14},
		8: {8, 9, 12, 13},
		9: {9, 10, 11, 12},
		// Crossbar
		10: {4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
		11: {4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
		// Stem
		12: {10, 11},
		13: {10, 11},
		14: {10, 11},
		15: {10, 11},
		16: {10, 11},
		17: {10, 11},
		18: {10, 11},
		19: {10, 11},
		20: {10, 11},
	}

	for y, xs := range rows {
		for _, x := range xs {
			if x >= 0 && x < size && y >= 0 && y < size {
				img.Set(x, y, black)
			}
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// getIcon returns the menu bar icon bytes.
func getIcon() []byte {
	return generateAnkhIcon()
}
