package main

import (
	_ "embed"
	"errors"
	"log"
	"os"
	"path/filepath"

	"go-keyboard-launcher/api"

	"github.com/BurntSushi/toml"
)

//go:embed assets/config.example.toml
var configExampleData []byte

type config struct {
	Hotkey  string
	Plugins map[string]toml.Primitive `toml:"plugin"`
}

func ConfigDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	return filepath.Join(dir, ".go-anywhere")
}

func ConfigFile() string {
	return filepath.Join(ConfigDir(), "config.toml")
}

func (a *App) ReadConfiguration() error {
	err := ensureDirectoryExists(ConfigDir())
	if err != nil {
		return err
	}

	f := ConfigFile()
	if _, err := os.Stat(f); errors.Is(err, os.ErrNotExist) {
		// Create the configuration file

		err := os.WriteFile(f, configExampleData, 0644)
		if err != nil {
			return errors.New("could not write to configuration file")
		}
	}

	err = a.readConfigFile(f)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) processConfig(c config) {
	a.hotKey = api.ParseHotkey(c.Hotkey)
	log.Printf("[DEBUG] Configured hot key %s\n", a.hotKey.String())

}

func (a *App) readConfigFile(filename string) error {
	base := config{}

	md, err := toml.DecodeFile(filename, &base)
	if err != nil {
		return err
	}

	a.processConfig(base)

	for k, prim := range base.Plugins {
		p, found := a.pluginByName(k)
		if found {
			p.LoadConfig(func(i interface{}) error {
				return md.PrimitiveDecode(prim, i)
			})
		}
	}

	return nil
}
