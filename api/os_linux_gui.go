//go:build linux || freebsd || openbsd

package api

/*
#cgo freebsd openbsd CFLAGS: -I/usr/X11R6/include -I/usr/local/include
#cgo freebsd openbsd LDFLAGS: -L/usr/X11R6/lib -L/usr/local/lib
#cgo freebsd openbsd LDFLAGS: -lX11 -lxkbcommon -lxkbcommon-x11 -lX11-xcb -lXcursor -lXfixes
#cgo linux pkg-config: x11 xkbcommon xkbcommon-x11 x11-xcb xcursor xfixes

#include <stdlib.h>
#include <locale.h>
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/Xutil.h>
#include <X11/Xresource.h>
#include <X11/XKBlib.h>
#include <X11/Xlib-xcb.h>
#include <X11/extensions/Xfixes.h>
#include <X11/Xcursor/Xcursor.h>
#include <xkbcommon/xkbcommon-x11.h>
*/
import "C"

import (
	"fmt"
	"image"
	"log"
	"runtime"
	"unsafe"
)

func GetShellIconImage(path string) (im image.Image, err error) {
	err = fmt.Errorf("Not supported on linux")
	return
}

func ShellExecuteItem(cmd string) {

}

func RegisterHotKey(hotkey Hotkey, pressed func()) error {
	runtime.LockOSThread()

	dpy := C.XOpenDisplay(nil)
	defer C.XCloseDisplay(dpy)

	root := C.XDefaultRootWindow(dpy)

	//unsigned int    modifiers       = ControlMask | ShiftMask;
	//int             keycode         = XKeysymToKeycode(dpy,XK_Y);
	//Window          grab_window     =  root;
	//Bool            owner_events    = False;
	//int             pointer_mode    = GrabModeAsync;
	//int             keyboard_mode   = GrabModeAsync;
	keyCode := C.XKeysymToKeycode(dpy, C.ulong(hotkey.KeyCode))

	C.XGrabKey(dpy, C.int(keyCode), C.ControlMask, root, C.False, C.GrabModeAsync, C.GrabModeAsync)
	C.XSelectInput(dpy, root, C.KeyPressMask)

	log.Println("[DEBUG] Hotkey registered")

	for {
		var ev C.XEvent
		C.XNextEvent(dpy, &ev)

		t := (*C.XAnyEvent)(unsafe.Pointer(&ev))._type

		switch t {
		case C.KeyPress:
			log.Println("[DEBUG] Key pressed")
			pressed()
		}
	}
}
