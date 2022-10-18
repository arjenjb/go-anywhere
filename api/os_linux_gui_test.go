//go:build linux || freebsd || openbsd

package api

import "testing"

func Test_RegisterHotkey(t *testing.T) {
	h := Hotkey{
		Modifiers: ModAlt,
		KeyCode:   'X',
	}

	RegisterHotKey(h, nil)
}
