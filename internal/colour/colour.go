// Package colour implements basic text colouring for showing text diffs.
package colour

import (
	"os"
	"sync"
)

// ANSI codes for coloured output, they are all the same length so as not to throw off
// alignment of [text/tabwriter].
const (
	codeRed    = "\x1b[0;0031m" // Red, used for diff lines starting with '-'
	codeHeader = "\x1b[1;0036m" // Bold cyan, used for diff headers starting with '@@'
	codeGreen  = "\x1b[0;0032m" // Green, used for diff lines starting with '+'
	codeReset  = "\x1b[000000m" // Reset all attributes
)

// getColourOnce is a [sync.OnceValues] function that returns the state of
// $NO_COLOR and $FORCE_COLOR, once and only once to avoid us calling
// os.Getenv on every call to a colour function.
var getColourOnce = sync.OnceValues(getColour)

// getColour returns whether $NO_COLOR and $FORCE_COLOR were set.
func getColour() (noColour bool, forceColour bool) {
	no := os.Getenv("NO_COLOR") != ""
	force := os.Getenv("FORCE_COLOR") != ""

	return no, force
}

// Header returns a diff header styled string.
func Header(text string) string {
	return sprint(codeHeader, text)
}

// Green returns a green styled string.
func Green(text string) string {
	return sprint(codeGreen, text)
}

// Red returns a red styled string.
func Red(text string) string {
	return sprint(codeRed, text)
}

// sprint returns a string with a given colour and the reset code.
//
// It handles checking for NO_COLOR and FORCE_COLOR.
func sprint(code, text string) string {
	noColor, forceColor := getColourOnce()

	// $FORCE_COLOR overrides $NO_COLOR
	if forceColor {
		return code + text + codeReset
	}

	// $NO_COLOR is next
	if noColor {
		return text
	}
	return code + text + codeReset
}
