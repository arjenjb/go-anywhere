package expr

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"image"
	"log"
	"strconv"

	"go-keyboard-launcher/api"

	"github.com/hashicorp/go-hclog"
	"golang.design/x/clipboard"
)

//go:embed icon.png
var iconData []byte

const (
	ExpressionCategory = api.User + 1
)

type Plugin struct {
	log  hclog.Logger
	icon *image.Image
}

func (p *Plugin) Initialize(log hclog.Logger) {
	p.log = log
	decoded, _, err := image.Decode(bytes.NewReader(iconData))
	if err == nil {
		p.icon = &decoded
	}
}

func (p *Plugin) LoadConfig(f func(interface{}) error) {
	// No configuration to load
}

func (p *Plugin) Catalog(context.Context) error {
	// Items are dynamic, nothing to do yet
	return nil
}

func (p *Plugin) Name() string {
	return "expr"
}

func (p *Plugin) Icon() *image.Image {
	return p.icon
}

func (p *Plugin) GetItems() ([]api.Item, error) {
	return nil, nil
}

func (p *Plugin) Execute(item api.Item) {
	if item.Category == ExpressionCategory {
		clipboard.Write(clipboard.FmtText, []byte(item.Target))

	} else {
		log.Printf("I don't know how to execute item %s", item.String())
	}
}

func (p *Plugin) Suggest(ctx context.Context, input string, chain []api.Item, setSuggestions api.SuggestionCallback) {
	// Try to parse the input as an expression
	expr := &Expression{}
	if err := parser.ParseString("", input, expr); err != nil {
		return
	}

	result := expr.Eval()

	if result == float64(int64(result)) {
		intValue := int64(result)

		setSuggestions([]api.Item{{
			Label:       fmt.Sprintf("= %d", intValue),
			Description: "Press Enter to copy the result",
			Category:    ExpressionCategory,
			Target:      fmt.Sprintf("%d", intValue),
		}, {
			Label:       fmt.Sprintf("= 0x%s", strconv.FormatInt(intValue, 16)),
			Description: "Press Enter to copy the result",
			Category:    ExpressionCategory,
			Target:      fmt.Sprintf("%d", intValue),
		}, {
			Label:       fmt.Sprintf("= 0b%s", strconv.FormatInt(intValue, 2)),
			Description: "Press Enter to copy the result",
			Category:    ExpressionCategory,
			Target:      fmt.Sprintf("%d", intValue),
		}}, api.MatchAny)
	} else {
		setSuggestions([]api.Item{{
			Label:       fmt.Sprintf("= %f", result),
			Description: "Press Enter to copy the result",
			Category:    ExpressionCategory,
			Target:      fmt.Sprintf("%f", result),
		}}, api.MatchAny)
	}
}
