package main

import (
	_ "embed"
	"log"
	"os"

	"go-keyboard-launcher/api"
	"go-keyboard-launcher/plugin/expr"
	"go-keyboard-launcher/plugin/github"
	"go-keyboard-launcher/plugin/str"

	"gioui.org/app"
	"github.com/getlantern/systray"
)

//go:embed assets/icon.ico
var iconWindows []byte

func (a *App) Show() {
	go func() { a.eventChannel <- api.EventShow }()
}

func (a *App) Hide() {
	go func() { a.eventChannel <- api.EventHide }()
}

func main() {
	plugins := []api.Plugin{
		//&startmenu.Plugin{},
		&expr.Plugin{},
		&str.Plugin{},
		&github.Plugin{},
	}

	a := NewApp(plugins)
	err := a.ReadConfiguration()

	if err != nil {
		log.Fatalln(err)
		return
	}

	a.Catalog()

	go func() {
		a.registerHotkey()
		a.isVisible = true

		if err := a.run(); err != nil {
			log.Fatalf("[FATAL] Application error: %v\n", err)
		}

		log.Println("Mainloop is done")
		systray.Quit()
	}()

	go func() {
		var trayReady = func() {
			setupTray(
				func() {
					a.Show()
				},
				func() {
					a.Quit()
				},
			)
		}

		systray.Run(trayReady, func() {
			log.Println("Tray has been properly shutdown, now we can exit")
			os.Exit(0)
		})
	}()

	app.Main()
}

func setupTray(doShow func(), doQuit func()) {
	systray.SetTemplateIcon(iconWindows, iconWindows)
	systray.SetTitle("Go Anywhere")
	systray.SetTooltip("Go Anywhere")

	itemOpen := systray.AddMenuItem("Open", "")
	itemQuit := systray.AddMenuItem("Quit", "")

	for {
		select {
		case <-itemOpen.ClickedCh:
			doShow()
		case <-itemQuit.ClickedCh:
			doQuit()
		}
	}
}
