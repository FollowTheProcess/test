// Package colour implements basic text colouring for showing text diffs.
package colour

import (
	"github.com/fatih/color"
)

var (
	header = color.New(color.FgCyan, color.Bold)
	green  = color.New(color.FgGreen)
	red    = color.New(color.FgRed)
)

// Header returns a diff header styled string.
func Header(text string) string {
	return header.Sprint(text)
}

// Green returns a green styled string.
func Green(text string) string {
	return green.Sprint(text)
}

// Red returns a red styled string.
func Red(text string) string {
	return red.Sprint(text)
}
