package startmenu

import (
	"context"
	"image"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"go-keyboard-launcher/api"
)

type Plugin struct {
	icon  *image.Image
	items []api.Item
}

func (p *Plugin) LoadConfig(f func(interface{}) error) {
	// No configuration to load
}

func (p *Plugin) Catalog() error {
	directory := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "Start Menu", "Programs")
	err := collectApplicationsFrom(directory, &p.items)
	if err != nil {
		return err
	}

	directory = filepath.Join(os.Getenv("AppData"), "Microsoft", "Windows", "Start Menu", "Programs")
	err = collectApplicationsFrom(directory, &p.items)
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) Name() string {
	return "win32-apps"
}

func (p *Plugin) Initialize() {

}

func (p *Plugin) Icon() *image.Image {
	return p.icon
}

func (p *Plugin) Suggest(ctx context.Context, input string, chain []api.Item, callback api.SuggestionCallback) {
	return
}

func (p *Plugin) Execute(item api.Item) {

}

func (p *Plugin) GetItems() ([]api.Item, error) {
	return p.items, nil
}

func collectApplicationsFrom(directory string, result *[]api.Item) error {
	return filepath.Walk(directory, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		name := filepath.Base(info.Name())
		extension := filepath.Ext(name)

		if extension == ".lnk" || extension == ".url" {
			var iconImage *image.Image

			if i, err := api.GetShellIconImage(path); err == nil {
				iconImage = &i
			} else {
				log.Printf("Could not load image for %s\n", path)
			}

			*result = append(*result, api.Item{
				Label:    name[0 : len(name)-len(extension)],
				Category: api.File,
				Data:     path,
				Icon:     iconImage,
			})
		}

		return nil
	})
}
