package goterm

import (
	"io"
	"os"
	"unicode"

	"golang.org/x/crypto/ssh/terminal"
)

// RuneTwoWidthTables is the slice of unicode.RangeTable containing the unicode characters whose display width is two.
// Notice that, hankaku (half-width) katakana (in Japanese) characters are contained in unicode.Katakana, but the display width should be one.
var RuneTwoWidthTables = []*unicode.RangeTable{
	unicode.Han, unicode.Hangul, unicode.Hiragana, unicode.Katakana,
	{R16: []unicode.Range16{{0x3000, 0x303f, 1}}}, // CJK symbols and punctuations
	{R16: []unicode.Range16{{0x30a0, 0x30ff, 1}}}, // zenkaku katakana
	{R16: []unicode.Range16{{0xff01, 0xff60, 1}}}, // halfwidth and fullwidth forms
}

// RuneZeroWidthTables is the slice of unicode.RangeTable containing the unicode characters whose display width is zero.
var RuneZeroWidthTables = []*unicode.RangeTable{
	unicode.Mn, unicode.Me, unicode.Cc, unicode.Cf,
}

// IsTerminal returns ture if w writes to a terminal.
func IsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return terminal.IsTerminal(int(f.Fd()))
}

// RuneWidth returns the display width of the given rune.
func RuneWidth(r rune) int {
	if unicode.IsOneOf(RuneZeroWidthTables, r) {
		return 0
	} else if unicode.IsOneOf(RuneTwoWidthTables, r) {
		// Handle the hankaku (half-idth) katakana (in Japanese) characters.
		if '\uff61' <= r && r <= '\uff9f' {
			return 1
		}
		return 2
	}
	return 1
}
