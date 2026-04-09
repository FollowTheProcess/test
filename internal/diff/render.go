package diff

import (
	"go.followtheprocess.codes/hue"
)

const (
	styleHeaderBold       = hue.Bold
	styleRemovedHeader    = hue.Red
	styleAddedHeader      = hue.Green
	styleRemovedLine      = hue.Red
	styleAddedLine        = hue.Green
	styleRemovedHighlight = hue.Black | hue.Bold | hue.RedBackground
	styleAddedHighlight   = hue.Black | hue.Bold | hue.GreenBackground
)

// Render formats a []Line as a colourised string suitable for terminal output.
// Returns an empty string if lines is nil or empty.
func Render(lines []Line) string {
	if len(lines) == 0 {
		return ""
	}

	var buf []byte

	i := 0
	for i < len(lines) {
		line := lines[i]

		switch line.Kind {
		case KindHeader:
			switch {
			case len(line.Content) >= 3 && string(line.Content[:3]) == "---":
				buf = styleRemovedHeader.AppendText(buf, line.Content)
			case len(line.Content) >= 3 && string(line.Content[:3]) == "+++":
				buf = styleAddedHeader.AppendText(buf, line.Content)
			default:
				buf = styleHeaderBold.AppendText(buf, line.Content)
			}
			i++

		case KindContext:
			buf = append(buf, ' ', ' ')
			buf = append(buf, line.Content...)
			i++

		case KindRemoved:
			// Collect the full consecutive run of removed lines, then any trailing added lines.
			start := i
			for i < len(lines) && lines[i].Kind == KindRemoved {
				i++
			}

			removedEnd := i
			for i < len(lines) && lines[i].Kind == KindAdded {
				i++
			}

			removed := lines[start:removedEnd]
			added := lines[removedEnd:i]

			if len(removed) == len(added) {
				buf = renderInlinePairs(buf, removed, added)
			} else {
				buf = renderWholeLine(buf, removed, added)
			}

		case KindAdded:
			// Standalone added block — no preceding removed lines in this hunk.
			start := i
			for i < len(lines) && lines[i].Kind == KindAdded {
				i++
			}

			buf = renderWholeLine(buf, nil, lines[start:i])

		default:
			// no action for unknown line kinds
			i++
		}
	}

	return string(buf)
}

// renderInlinePairs renders 1:1 paired removed/added lines with character-level inline diff.
func renderInlinePairs(buf []byte, removed, added []Line) []byte {
	for k := range removed {
		ic := CharDiff(removed[k].Content, added[k].Content)

		buf = styleRemovedLine.AppendText(buf, []byte("- "))

		for _, seg := range ic.Removed {
			if seg.Changed {
				buf = styleRemovedHighlight.AppendText(buf, seg.Text)
			} else {
				buf = styleRemovedLine.AppendText(buf, seg.Text)
			}
		}

		buf = styleAddedLine.AppendText(buf, []byte("+ "))

		for _, seg := range ic.Added {
			if seg.Changed {
				buf = styleAddedHighlight.AppendText(buf, seg.Text)
			} else {
				buf = styleAddedLine.AppendText(buf, seg.Text)
			}
		}
	}

	return buf
}

// renderWholeLine renders removed/added lines with whole-line colour (no inline diff).
func renderWholeLine(buf []byte, removed, added []Line) []byte {
	for _, r := range removed {
		buf = styleRemovedLine.AppendText(buf, []byte("- "))
		buf = styleRemovedLine.AppendText(buf, r.Content)
	}

	for _, a := range added {
		buf = styleAddedLine.AppendText(buf, []byte("+ "))
		buf = styleAddedLine.AppendText(buf, a.Content)
	}

	return buf
}
