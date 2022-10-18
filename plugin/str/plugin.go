package str

import (
	"context"
	"encoding/base64"
	"image"
	"log"
	"os"

	"go-keyboard-launcher/api"

	"golang.design/x/clipboard"
)

type Plugin struct {
	icon *image.Image
}

func (p *Plugin) LoadConfig(f func(interface{}) error) {
	// No configuration to load
}

func (p *Plugin) Catalog() error {
	// Nothing to do here
	return nil
}

func (p *Plugin) Name() string {
	return "str"
}

func (p *Plugin) Initialize() {
	p.loadIcon()
}

func (p *Plugin) loadIcon() {
	f, err := os.Open("plugin/str/icon.png")
	if err != nil {
		log.Printf("[ERROR] Could not read icon")
		return
	}

	defer f.Close()

	decoded, _, err := image.Decode(f)
	if err != nil {
		log.Printf("[ERROR] Could not decode icon")
		return
	}

	p.icon = &decoded
}

func (p Plugin) Icon() *image.Image {
	return p.icon
}

func (p Plugin) GetItems() ([]api.Item, error) {
	return []api.Item{{
		Label:       "String: Base64",
		Description: "",
		Category:    api.User,
		Target:      "",
		Data:        nil,
		Icon:        nil,
		ArgsHint:    api.Required,
	}}, nil
}

func (p Plugin) Suggest(ctx context.Context, input string, chain []api.Item, callback api.SuggestionCallback) {
	if len(chain) == 0 {
		return
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(input))

	suggestions := []api.Item{{
		Label:       encoded,
		Description: "Encoded string",
		Target:      encoded,
		Category:    api.User,
	}}

	decoded, err := base64.StdEncoding.DecodeString(input)

	if err == nil {
		suggestions = append(suggestions, api.Item{
			Label:       string(decoded),
			Description: "Decoded string",
			Target:      string(decoded),
			Category:    api.User,
		})
	}

	callback(suggestions, api.MatchAny)
}

func (p Plugin) Execute(item api.Item) {
	clipboard.Write(clipboard.FmtText, []byte(item.Target))
}
