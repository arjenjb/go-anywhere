package aws

//
//import (
//	"bytes"
//	"context"
//	_ "embed"
//	"fmt"
//	"image"
//	"log"
//
//	"go-keyboard-launcher/api"
//)
//
////go:embed logo.png
//var iconData []byte
//
//const (
//	BranchesCategory = api.User + 1
//)
//
//const (
//	KeywordConfigure uint8 = iota
//)
//
//type Plugin struct {
//	icon *image.Image
//}
//
//func (p *Plugin) Initialize() {
//	decoded, _, err := image.Decode(bytes.NewReader(iconData))
//	if err == nil {
//		p.icon = &decoded
//	}
//}
//
//func (p *Plugin) Icon() *image.Image {
//	return p.icon
//}
//
//func (p *Plugin) GetItems() ([]api.Item, error) {
//	result := make([]api.Item, 0)
//
//	result = append(result, api.Item{
//		Label:    "Plugin Github: Configure",
//		Category: api.Keyword,
//		ArgsHint: api.Forbidden,
//		Data:     KeywordConfigure,
//	})
//	return result, nil
//}
//
//func (p *Plugin) Execute(item api.Item) {
//	log.Printf("I don't know how to execute item %s", item.String())
//}
//
//func (p *Plugin) Suggest(ctx context.Context, input string, chain []api.Item, setSuggestions api.SuggestionCallback) {
//	if len(chain) == 0 {
//		return
//	}
//
//	setSuggestions([]api.Item{
//		{
//			Label:    fmt.Sprintf("Branches"),
//			Category: BranchesCategory,
//		},
//		{
//			Label:    fmt.Sprintf("Pull requests"),
//			Category: api.Url,
//			Target:   fmt.Sprintf("%s/pulls", chain[0].Target),
//		},
//	}, api.MatchFuzzy)
//}
