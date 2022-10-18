package api

import (
	"bytes"
	"fmt"
	"log"
	"strings"
)

const (
	ModAlt = 1 << iota
	ModCtrl
	ModShift
	ModWin
)

type Hotkey struct {
	Modifiers int // Mask of modifiers
	KeyCode   int // Key code, e.g. 'A'
}

// String returns a human-friendly display name of the hotkey
// such as "Hotkey[Id: 1, Alt+Ctrl+O]"
func (h *Hotkey) String() string {
	mod := &bytes.Buffer{}
	if h.Modifiers&ModAlt != 0 {
		mod.WriteString("Alt+")
	}
	if h.Modifiers&ModCtrl != 0 {
		mod.WriteString("Ctrl+")
	}
	if h.Modifiers&ModShift != 0 {
		mod.WriteString("Shift+")
	}
	if h.Modifiers&ModWin != 0 {
		mod.WriteString("Win+")
	}

	key := fmt.Sprintf("%c", h.KeyCode)

	switch h.KeyCode {
	case 0x20:
		key = "Space"
	}

	return fmt.Sprintf("Hotkey[%s%s]", mod, key)
}

func ParseHotkey(def string) Hotkey {
	a := strings.Split(def, "+")
	modifier := 0
	key := -1

	for i := range a {
		m := strings.ToLower(strings.TrimSpace(a[i]))

		if i < len(a)-1 {
			// Handle modifiers

			switch m {
			case "ctrl":
				modifier += ModCtrl
			case "shift":
				modifier += ModShift
			case "alt":
				modifier += ModAlt
			case "win":
				modifier += ModWin
			default:
				log.Fatalf("Hotkey modifier `%s` not supported", m)
			}

		} else {
			switch m {
			case "space":
				key = 0x20
			case "ins":
				fallthrough
			case "insert":
				key = 45
			case "del":
				fallthrough
			case "delete":
				key = 46
			case "home":
				key = 36
			case "end":
				key = 35
			case "backspace":
				key = 8
			case "esc":
				fallthrough
			case "escape":
				key = 27
			case "break":
				fallthrough
			case "pause":
				key = 19
			case "f1":
				key = 112
			case "f2":
				key = 113
			case "f3":
				key = 114
			case "f4":
				key = 115
			case "f5":
				key = 116
			case "f6":
				key = 117
			case "f7":
				key = 118
			case "f8":
				key = 119
			case "f9":
				key = 120
			case "f10":
				key = 121
			case "f11":
				key = 122
			case "f12":
				key = 123
			case "tab":
				key = 9
			case "enter":
				key = 13
			default:
				if len(m) == 1 {
					key = int(m[0])
				} else {
					log.Fatalf("Hotkey `%s` not supported", m)
				}
			}
		}
	}
	return Hotkey{Modifiers: modifier, KeyCode: key}
}
