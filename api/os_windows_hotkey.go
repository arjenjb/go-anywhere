//go:build windows

package api

import (
	windows "go-keyboard-launcher/api/internal"
	"log"
	"runtime"
)

var (
	reghotkey = moduser32.NewProc("RegisterHotKey")
)

func RegisterHotKey(hotkey Hotkey, onHotKeyPressed func()) {
	runtime.LockOSThread()

	r1, _, err := reghotkey.Call(
		0, 0, uintptr(hotkey.Modifiers), uintptr(hotkey.KeyCode))

	if r1 != 1 {
		log.Println("Failed to register", hotkey, ", error:", err)
		return
	}

	msg := new(windows.Msg)
	for {
		switch ret := windows.GetMessage(msg, 0, 0, 0); ret {
		case -1:
			log.Printf("[ERROR] Got -1 from hotkey message loop, stopping the loop")
			return
		case 0:
			break
		default:
			log.Printf("[DEBUG] Hotkey pressed\n")
			onHotKeyPressed()
		}

		windows.TranslateMessage(msg)
		windows.DispatchMessage(msg)
	}
}
