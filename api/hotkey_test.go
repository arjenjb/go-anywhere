package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHotkey(t *testing.T) {
	var hotkey Hotkey

	ParseHotkey("Ctrl+Shift+Space")
	ParseHotkey("Ctrl+Shift +  space ")

	hotkey = ParseHotkey("ALt+Space ")
	assert.Equal(t, ModAlt, hotkey.Modifiers)
	assert.Equal(t, 32, hotkey.KeyCode)

	hotkey = ParseHotkey("alt+x")
	assert.Equal(t, ModAlt, hotkey.Modifiers)
	assert.Equal(t, 120, hotkey.KeyCode)

	hotkey = ParseHotkey("ctrl+win+k")
	assert.Equal(t, ModCtrl|ModWin, hotkey.Modifiers)
	assert.Equal(t, 107, hotkey.KeyCode)
}
