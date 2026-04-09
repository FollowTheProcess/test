package diff

import (
	"unicode/utf8"
)

// Segment is a contiguous run of text tagged as equal or changed.
type Segment struct {
	Text    []byte
	Changed bool
}

// InlineChange holds the char-level breakdown for a removed/added line pair.
type InlineChange struct {
	Removed []Segment
	Added   []Segment
}

// CharDiff computes character-level segments for a removed/added line pair.
// Uses O(mn) LCS on []rune — acceptable since individual lines are short.
// Strips and restores trailing \n before diffing.
// If either side exceeds 500 runes, returns a single-segment fallback (whole-line).
func CharDiff(removed, added []byte) InlineChange {
	// Strip trailing newline, remember whether each side had one.
	removedNL := len(removed) > 0 && removed[len(removed)-1] == '\n'
	addedNL := len(added) > 0 && added[len(added)-1] == '\n'

	removedCore := removed
	if removedNL {
		removedCore = removed[:len(removed)-1]
	}

	addedCore := added
	if addedNL {
		addedCore = added[:len(added)-1]
	}

	oldRunes := []rune(string(removedCore))
	newRunes := []rune(string(addedCore))

	// Safety cap: avoid O(mn) on large inputs (minified/generated content).
	if len(oldRunes) > 500 || len(newRunes) > 500 {
		return fallback(removed, added)
	}

	segs := lcsSegments(oldRunes, newRunes)

	removedSegs := reattachNL(segs.removed, removedNL)
	addedSegs := reattachNL(segs.added, addedNL)

	return InlineChange{Removed: removedSegs, Added: addedSegs}
}

// fallback returns a single Changed segment per side (whole-line fallback).
// Copies the input slices to avoid aliasing the caller's memory.
func fallback(removed, added []byte) InlineChange {
	result := InlineChange{}

	if len(removed) > 0 {
		cp := make([]byte, len(removed))
		copy(cp, removed)
		result.Removed = []Segment{{Text: cp, Changed: true}}
	}

	if len(added) > 0 {
		cp := make([]byte, len(added))
		copy(cp, added)
		result.Added = []Segment{{Text: cp, Changed: true}}
	}

	return result
}

// reattachNL returns a new segment slice with a newline reattached after the
// last segment. The input slice is not modified.
//
// If the last segment is Changed, the newline is appended as a separate
// unchanged segment rather than being included in the highlight span — a
// highlighted \n causes the ANSI background colour to bleed onto the next
// terminal line.
func reattachNL(segs []Segment, hadNL bool) []Segment {
	if !hadNL || len(segs) == 0 {
		return segs
	}

	last := segs[len(segs)-1]

	if last.Changed {
		result := make([]Segment, 0, len(segs)+1)
		result = append(result, segs...)

		return append(result, Segment{Text: []byte{'\n'}, Changed: false})
	}

	result := make([]Segment, len(segs))
	copy(result, segs)

	last.Text = append(append([]byte(nil), last.Text...), '\n')
	result[len(result)-1] = last

	return result
}

// sides holds parallel removed/added segment lists built during backtracking.
type sides struct {
	removed []Segment
	added   []Segment
}

// op is a single edit operation produced during LCS backtracking.
type op struct {
	r    rune
	kind byte // 'e' equal, 'd' delete, 'i' insert
}

// lcsSegments builds the LCS table and backtracks to produce Equal/Delete/Insert
// edit operations, then merges consecutive same-kind ops into Segment runs.
func lcsSegments(old, newText []rune) sides {
	m, n := len(old), len(newText)

	// Build LCS DP table using a flat slice to avoid m+1 separate allocations.
	// dp[i*(n+1)+j] = length of LCS of old[:i] and newText[:j].
	dp := make([]int, (m+1)*(n+1))

	stride := n + 1
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if old[i-1] == newText[j-1] {
				dp[i*stride+j] = dp[(i-1)*stride+(j-1)] + 1
			} else {
				dp[i*stride+j] = max(dp[(i-1)*stride+j], dp[i*stride+(j-1)])
			}
		}
	}

	// Backtrack to build edit ops, then reverse.
	ops := make([]op, 0, m+n)

	i, j := m, n
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && old[i-1] == newText[j-1]:
			ops = append(ops, op{r: old[i-1], kind: 'e'})
			i--
			j--
		case j > 0 && (i == 0 || dp[i*stride+(j-1)] >= dp[(i-1)*stride+j]):
			ops = append(ops, op{r: newText[j-1], kind: 'i'})
			j--
		default:
			ops = append(ops, op{r: old[i-1], kind: 'd'})
			i--
		}
	}

	// Reverse ops (they were built backwards).
	for l, r := 0, len(ops)-1; l < r; l, r = l+1, r-1 {
		ops[l], ops[r] = ops[r], ops[l]
	}

	return mergeSegments(ops)
}

// mergeSegments merges consecutive same-kind ops into Segment runs.
func mergeSegments(ops []op) sides {
	// Pre-allocate with a reasonable capacity to reduce re-allocations.
	removedSegs := make([]Segment, 0, 8)
	addedSegs := make([]Segment, 0, 8)

	var buf [utf8.UTFMax]byte
	for _, o := range ops {
		nb := utf8.EncodeRune(buf[:], o.r)
		r := buf[:nb]

		switch o.kind {
		case 'e':
			if len(removedSegs) > 0 && !removedSegs[len(removedSegs)-1].Changed {
				removedSegs[len(removedSegs)-1].Text = append(removedSegs[len(removedSegs)-1].Text, r...)
			} else {
				removedSegs = append(removedSegs, Segment{Text: append([]byte(nil), r...), Changed: false})
			}

			if len(addedSegs) > 0 && !addedSegs[len(addedSegs)-1].Changed {
				addedSegs[len(addedSegs)-1].Text = append(addedSegs[len(addedSegs)-1].Text, r...)
			} else {
				addedSegs = append(addedSegs, Segment{Text: append([]byte(nil), r...), Changed: false})
			}
		case 'd':
			if len(removedSegs) > 0 && removedSegs[len(removedSegs)-1].Changed {
				removedSegs[len(removedSegs)-1].Text = append(removedSegs[len(removedSegs)-1].Text, r...)
			} else {
				removedSegs = append(removedSegs, Segment{Text: append([]byte(nil), r...), Changed: true})
			}
		case 'i':
			if len(addedSegs) > 0 && addedSegs[len(addedSegs)-1].Changed {
				addedSegs[len(addedSegs)-1].Text = append(addedSegs[len(addedSegs)-1].Text, r...)
			} else {
				addedSegs = append(addedSegs, Segment{Text: append([]byte(nil), r...), Changed: true})
			}
		default:
			// op.kind is always 'e', 'd', or 'i' — lcsSegments is the only producer.
		}
	}

	// Handle empty inputs: produce a single unchanged empty segment so callers
	// always have at least one segment to work with.
	if len(removedSegs) == 0 {
		removedSegs = append(removedSegs, Segment{Text: []byte{}, Changed: false})
	}

	if len(addedSegs) == 0 {
		addedSegs = append(addedSegs, Segment{Text: []byte{}, Changed: false})
	}

	return sides{removed: removedSegs, added: addedSegs}
}
