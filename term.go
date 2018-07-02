package goterm

import (
	"fmt"
	"os"
	"unicode/utf8"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	keyCtrlD     = 004
	keyEnter     = 015
	keyEscape    = 033
	keyBackspace = 0177
	keyUnknown   = 0xd800 /* UTF-16 surrogate area */ + iota
	keyUp
	keyDown
	keyRight
	keyLeft
)

// TermConfig is configure settings for a Term.
type TermConfig struct {
	// History indicates whether the input history is enabled.
	History bool
}

// Term is the terminal emulator supporting full-width (zenkaku in Japanese) characters.
type Term struct {
	f                     *os.File
	stateBefore, stateRaw *terminal.State
	keybuf                []byte
	prompt                string
	config                TermConfig
	input                 []rune
	cursor                int
	history               [][]rune
	historyIdx            int
}

// NewTerm returns a new Term.
func NewTerm(f *os.File, prompt string, config TermConfig) (*Term, error) {
	stateBefore, err := terminal.MakeRaw(int(f.Fd()))
	if err != nil {
		return nil, err
	}
	stateRaw, err := terminal.GetState(int(f.Fd()))
	if err != nil {
		return nil, err
	}
	term := &Term{
		f:           f,
		stateBefore: stateBefore,
		stateRaw:    stateRaw,
		prompt:      prompt,
		config:      config,
	}
	term.changeRawMode(false)
	return term, nil
}

func (term *Term) changeRawMode(raw bool) {
	state := term.stateBefore
	if raw {
		state = term.stateRaw
	}
	terminal.Restore(int(term.f.Fd()), state)
}

func (term *Term) clearLine() {
	term.f.Write([]byte("\r\033[2K"))
}

func (term *Term) clearLineRemain() {
	term.f.Write([]byte("\033[0K"))
}

func (term *Term) eraseLast() bool {
	if len(term.input) > 0 {
		term.f.Write([]byte(fmt.Sprintf("\033[%dD\033[0K", RuneWidth(term.input[len(term.input)-1]))))
		term.input = term.input[:len(term.input)-1]
	}
	return len(term.input) > 0
}

func (term *Term) fillKeybuf() error {
	if len(term.keybuf) == 0 {
		term.keybuf = make([]byte, 256)
	}
	n, err := term.f.Read(term.keybuf)
	if err != nil {
		return err
	}
	term.keybuf = term.keybuf[:n]
	return nil
}

func (term *Term) readKey() (rune, error) {
	if len(term.keybuf) == 0 {
		if err := term.fillKeybuf(); err != nil {
			return utf8.RuneError, err
		}
	}
	if term.keybuf[0] != keyEscape {
		r, l := utf8.DecodeRune(term.keybuf)
		term.keybuf = term.keybuf[l:]
		return r, nil
	}
	if len(term.keybuf) >= 3 && term.keybuf[1] == '[' {
		switch term.keybuf[2] {
		case 'A':
			term.keybuf = term.keybuf[3:]
			return keyUp, nil
		case 'B':
			term.keybuf = term.keybuf[3:]
			return keyDown, nil
		case 'C':
			term.keybuf = term.keybuf[3:]
			return keyRight, nil
		case 'D':
			term.keybuf = term.keybuf[3:]
			return keyLeft, nil
		case 'E', 'F':
			term.keybuf = term.keybuf[3:]
			return keyUnknown, nil
		}
	}
	term.keybuf = term.keybuf[1:]
	return keyUnknown, nil
}

// ReadLine returns the read line ending with '\n'.
// If EOF is hit, the empty string and no error are returned.
func (term *Term) ReadLine() (string, error) {
	term.changeRawMode(true)
	defer term.changeRawMode(false)
	term.input = term.input[:0]
	term.cursor = 0
	term.historyIdx = len(term.history)
	var swappedInput []rune
	term.clearLine()
	term.f.Write([]byte(term.prompt))
	for {
		key, err := term.readKey()
		if err != nil {
			// The error should be reported immediately.
			return "", err
		}
		if key == keyEnter {
			// Emit the input sequence with the newline character.
			term.f.Write([]byte("\r\n"))
			term.input = append(term.input, '\n')
			break
		}
		switch key {
		case keyBackspace:
			if term.cursor == 0 {
				break
			} else if term.cursor == len(term.input) {
				// Remove the last character.
				term.eraseLast()
				term.cursor--
			} else {
				// Remove the last character followed by the cursor.
				term.cursor--
				term.f.Write([]byte(fmt.Sprintf("\033[%dD", RuneWidth(term.input[term.cursor]))))
				copy(term.input[term.cursor:], term.input[term.cursor+1:])
				term.input = term.input[:len(term.input)-1]
				term.refreshLineRemain()
				term.f.Write([]byte(fmt.Sprintf("\033[%dD", RuneWidth(term.input[term.cursor]))))
			}
		case keyCtrlD:
			// EOS signal.
			return "", nil
		case keyDown:
			// History forward.
			if !term.config.History || term.historyIdx == len(term.history) {
				break
			}
			for term.eraseLast() {
			}
			term.historyIdx++
			if term.historyIdx == len(term.history) {
				term.input = swappedInput
			} else {
				input := term.history[term.historyIdx]
				term.input = make([]rune, len(input))
				copy(term.input, input)
			}
			term.f.Write([]byte(string(term.input)))
			term.cursor = len(term.input)
		case keyLeft:
			// Cursor backward.
			if term.cursor == 0 {
				break
			}
			term.cursor--
			r := term.input[term.cursor]
			term.f.Write([]byte(fmt.Sprintf("\033[%dD", RuneWidth(r))))
		case keyRight:
			// Cursor forward.
			if term.cursor == len(term.input) {
				break
			}
			// Simply using "\033[xC" won't stride the terminal border.
			term.refreshLineRemain()
			term.cursor++
		case keyUp:
			// History backward.
			if !term.config.History || term.historyIdx == 0 {
				break
			}
			for term.eraseLast() {
			}
			if term.historyIdx == len(term.history) {
				swappedInput = term.input
			}
			term.historyIdx--
			input := term.history[term.historyIdx]
			term.input = make([]rune, len(input))
			copy(term.input, input)
			term.f.Write([]byte(string(term.input)))
			term.cursor = len(term.input)
		default:
			if !(key >= 0x20 && !(0xd800 <= key && key <= 0xdbff)) {
				// Unhandled control characters.
				break
			}
			if term.cursor == len(term.input) {
				// Append the character.
				term.input = append(term.input, key)
			} else {
				// Insert the character.
				term.input = append(term.input[:term.cursor], append([]rune{key}, term.input[term.cursor:]...)...)
			}
			term.refreshLineRemain()
			term.cursor++
		}
	}
	if term.config.History {
		input := make([]rune, len(term.input)-1)
		copy(input, term.input)
		term.history = append(term.history, input)
	}
	return string(term.input), nil
}

func (term *Term) refreshLineRemain() {
	term.clearLineRemain()
	term.f.Write([]byte(string(term.input[term.cursor:])))
	for c := len(term.input) - 1; c > term.cursor; c-- {
		term.f.Write([]byte(fmt.Sprintf("\033[%dD", RuneWidth(term.input[c]))))
	}
}
