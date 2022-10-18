package main

import (
	"gioui.org/app"
	"go-keyboard-launcher/api"
	"go-keyboard-launcher/plugin/expr"
	"go-keyboard-launcher/plugin/startmenu"
	"go-keyboard-launcher/plugin/str"
	"log"
	"testing"
)

func TestApp_ItemStack(t *testing.T) {
	p := str.Plugin{}
	plugins := []api.Plugin{
		&p,
		&startmenu.Plugin{},
		&expr.Plugin{},
	}

	go func() {
		a := NewApp(plugins)
		a.isVisible = true

		err := a.run()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	app.Main()
}
