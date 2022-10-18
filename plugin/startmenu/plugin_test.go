package startmenu

import "testing"

func TestStartMenuPlugin_GetItems(t *testing.T) {
	p := Plugin{}
	items, _ := p.GetItems()
	for _, each := range items {
		println(each.Label)
	}
}
