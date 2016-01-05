// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package text

import (
	"bytes"
)

// These constants are not really necessary but make the fold code easier to
// read and understand. If chars or lines are defined too small there will
// potentially be additional allocations needed but wasted space will be
// reduced. If chars or lines are defined too large the allocations will be
// reduced but unused space will be allocated.
//
// TODO: Tune chars and lines at runtime based on average text sizes being
// processed? Will need to set maximum limits to avoid runaway sizing based on
// deliberatly large text being sent by players causing a denial of service.
const (
	reset = 0  // Reset buffer to start (position zero) or test if at start
	space = 1  // Width in bytes of a space
	chars = 32 // Starting number of characters for initial word buffer sizing
	lines = 24 // Starting number of lines for page initial buffer sizing
)

// Fold takes a string and reformats it so lines have a maximum length of the
// passed width. Fold will handle multibyte runes. However it cannot handle
// 'wide' runes - those that are wider than a normal single character when
// displayed. This is because the required information is actually contained in
// the font files of the font in use at the 'client' end.
//
// For example the Chinese for 9 is 九 (U+4E5D). Even in a monospaced font 九
// will take up the space of two columns.
func Fold(in []byte, width int) []byte {

	// Can we take a short cut? Counting bytes is fine although we may end up
	// with a string shorter than we think it is if there are multibyte runes.
	if len(in) <= width {
		return []byte(in)
	}

	// Add extra line feed to end of input. Will cause final word and line to be
	// 'flushed' from the buffers. The extra line feed itself will not be output
	// because it will still be in the buffers - so we don't need to trim it off.
	in = append(in, '\n')

	var (
		word = bytes.NewBuffer(make([]byte, 0, chars))
		line = bytes.NewBuffer(make([]byte, 0, width+chars))
		page = bytes.NewBuffer(make([]byte, 0, len(in)+lines))
	)

	var (
		wordLen, lineLen, pageLen = 0, 0, 0 // word, line and output length in runes
		blank                     = true    // true when line is empty or only blanks
	)

	for _, r := range bytes.Runes(in) {

		if (r != ' ' && r != '\n') || (r == ' ' && blank == true) {
			word.WriteRune(r)
			wordLen++
			blank = r == ' '
			continue
		}

		if lineLen+space+wordLen > width {
			if pageLen != reset {
				page.WriteByte('\n')
				pageLen++
			}
			line.WriteTo(page)
			pageLen += lineLen
			lineLen = reset
			blank = true
		}

		if lineLen != reset {
			line.WriteByte(' ')
			lineLen++
		}
		word.WriteTo(line)
		lineLen += wordLen
		wordLen = reset

		if r == '\n' {
			if pageLen != reset {
				page.WriteByte('\n')
				pageLen++
			}
			line.WriteTo(page)
			pageLen += lineLen
			lineLen = reset
			blank = true
		}

	}

	return page.Bytes()
}
