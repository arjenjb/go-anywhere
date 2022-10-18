package main

import (
	"context"
	"fmt"
	"image"
	"log"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"go-keyboard-launcher/api"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/hashicorp/go-hclog"
)

const WindowWidth = 600
const SearchBoxHeight = 33

type App struct {
	log       hclog.Logger
	plugins   []api.Plugin
	rootItems []InternalItem

	// Item state
	catalogMutex sync.Mutex
	itemIndex    int
	suggestItems []SuggestItem
	itemStack    []StackEntry

	lastVisibleItems     int           // Number of items that is shown
	lastClickTime        time.Duration // Time of last click on item, to detect double clicks
	lastSearchCancelFunc *context.CancelFunc
	eventChannel         chan api.Event
	hotKey               api.Hotkey

	// Gui state
	isVisible  bool
	textInput  widget.Editor
	listWidget widget.List
	eventKey   event.Tag
}

type InternalItem struct {
	Item       api.Item
	lookupName string
	plugin     api.Plugin
}

type StackEntry struct {
	item       InternalItem
	searchText string // The text that was searched for
}

type SuggestItem struct {
	Item  InternalItem
	Score float64
}

func (i InternalItem) DisplayName() string {
	return i.Item.Label
}

func (i InternalItem) Icon() *image.Image {
	if i.Item.Icon != nil {
		return i.Item.Icon
	} else if i.plugin.Icon() != nil {
		return i.plugin.Icon()
	} else {
		return nil
	}
}

func (i InternalItem) execute() {
	i.plugin.Execute(i.Item)
}

func (i InternalItem) Description() string {
	return i.Item.Description
}

func NewApp(plugins []api.Plugin) *App {
	a := App{
		log: hclog.New(&hclog.LoggerOptions{
			Level: hclog.LevelFromString("DEBUG"),
		}),
		plugins:      plugins,
		isVisible:    false,
		eventChannel: make(chan api.Event),
		itemIndex:    -1,
	}

	a.textInput.SetCaret(0, 0)
	a.textInput.SetText("")
	a.textInput.SingleLine = true
	a.textInput.Alignment = text.Start
	a.textInput.Focus()
	a.textInput.Submit = false

	a.listWidget = widget.List{List: layout.List{
		Axis: layout.Vertical,
	}}

	// Initialize the plugins
	for _, p := range plugins {
		p.Initialize()
	}

	return &a
}

func (a *App) Catalog() {
	for _, p := range a.plugins {
		if err := p.Catalog(); err != nil {
			log.Printf("Failed to catalog plugin %s\n: %s", p.Name(), err)
		}
	}
}

func (a *App) catalog() []InternalItem {
	if a.rootItems == nil {
		a.rebuildCatalog()
	}

	return a.rootItems
}

func (a *App) itemSuggest() {
	if !a.HasCurrentItem() {
		return
	}
	item := a.CurrentItem()

	if item.Item.ArgsHint == api.Forbidden {
		return
	}

	// Push the item on the stack
	a.itemStack = append(a.itemStack, StackEntry{
		item:       item,
		searchText: a.textInput.Text(),
	})
	a.resetInput()

	a.Search("")
}

func (a *App) itemDown() {
	if a.itemIndex == len(a.suggestItems)-1 {
		// We're at the end of the list
		return
	}
	a.itemIndex++

	// Update the list scroll position if needed
	p := a.listWidget.Position.First + a.lastVisibleItems
	if a.itemIndex >= p {
		a.listWidget.Position.First = a.itemIndex - a.lastVisibleItems + 1
	}
}

func (a *App) itemUp() {
	if a.itemIndex <= 0 {
		// We are already at the start of the list
		return
	}
	a.itemIndex--

	if a.itemIndex < a.listWidget.Position.First {
		a.listWidget.Position.First = a.itemIndex
	}
}

func (a *App) cancel() {
	if len(a.textInput.Text()) > 0 {
		a.resetInput()
	} else if len(a.itemStack) > 0 {
		a.popItemStack()
	} else {
		a.Hide()
	}
}

func (a *App) popItemStack() {
	// Pop an item off the stack
	top := a.itemStack[len(a.itemStack)-1]

	a.itemStack = a.itemStack[:len(a.itemStack)-1]

	// Restore the last search text
	a.textInput.SetText(top.searchText)
	a.textInput.SetCaret(len(top.searchText), len(top.searchText))
}

func (a *App) resetInput() {
	a.cancelLastSearch()
	a.clearSuggestions()

	a.textInput.SetText("")
}

func (a *App) enter() {
	if !a.HasCurrentItem() {
		return
	}
	item := a.CurrentItem()

	if item.Item.ArgsHint == api.Required {
		// Push the item on the stack
		a.itemStack = append(a.itemStack, StackEntry{
			item:       item,
			searchText: a.textInput.Text(),
		})

		a.resetInput()
		return
	}

	switch item.Item.Category {
	case api.File:
		log.Printf("Launching file item %s", item.Item)
		api.ShellExecuteItem(item.Item.Target)

	case api.Url:
		log.Printf("Launching URL item %s", item.Item)
		executeUrl(item.Item.Target)

	default:
		if item.Item.Category >= api.User {
			log.Printf("Executing user item %s", item.DisplayName())
			go item.execute()
		} else {
			log.Printf("Error: Cannot handle this type of category")
		}
	}

	a.Hide()
}

func executeUrl(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Fatal(err)
	}
}

func (a *App) Search(input string) {
	a.cancelLastSearch()

	// Remove whitespace and lower for search matching
	search := strings.ToLower(strings.Trim(input, " "))

	// Search query is empty
	if len(search) > 0 && len(a.itemStack) == 0 {
		// Find items that match directly
		var suggestions []SuggestItem

		for _, item := range a.catalog() {
			if s := MatchScore(search, item.lookupName); s > 0.0 {
				suggestions = append(suggestions, SuggestItem{Item: item, Score: s})
			}
		}

		a.setSuggestions(suggestions)
	} else {
		// If there is no search query, first empty the suggestions
		a.clearSuggestions()
	}

	// Dispatch search to plugins
	ctx, cancel := context.WithCancel(context.Background())
	a.lastSearchCancelFunc = &cancel

	// Let the plugins do some suggestions
	if len(a.itemStack) > 0 {
		p := a.itemStack[0].item.plugin
		stack := collectItems(a.itemStack)
		log.Printf("Stack size: %d\n", len(stack))
		// The original user input is passed
		go p.Suggest(ctx, input, stack, func(items []api.Item, match api.Match) {
			internalItems := createSuggestions(items, match, p, search)
			a.addSuggestions(internalItems)
		})

	} else {
		// Broadcast the search query to all plugins
		for _, p := range a.plugins {
			go p.Suggest(ctx, input, nil, func(items []api.Item, match api.Match) {
				internalItems := createSuggestions(items, match, p, search)
				a.addSuggestions(internalItems)
			})
		}
	}
}

func createSuggestions(items []api.Item, match api.Match, p api.Plugin, search string) []SuggestItem {
	var internalItems []SuggestItem
	for _, i := range items {
		ii := asInternalItem(i, p)

		var score float64
		if match == api.MatchAny {
			score = 1.0

		} else {
			if len(search) == 0 {
				score = 1.0
			} else {
				score = MatchScore(search, ii.lookupName)

				if score == 0.0 {
					continue
				}
			}
		}

		internalItems = append(internalItems, SuggestItem{
			Item:  ii,
			Score: score,
		})
	}
	return internalItems
}

func collectItems(stack []StackEntry) []api.Item {
	result := make([]api.Item, len(stack))
	for idx, e := range stack {
		result[idx] = e.item.Item
	}
	return result
}

func (a *App) cancelLastSearch() {
	if a.lastSearchCancelFunc != nil {
		// Cancel last search request
		(*a.lastSearchCancelFunc)()
		a.lastSearchCancelFunc = nil
	}
}

func (a *App) HasCurrentItem() bool {
	return a.itemIndex != -1
}

func (a *App) CurrentItem() InternalItem {
	return a.suggestItems[a.itemIndex].Item
}

func (a *App) rebuildCatalog() {
	var catalog []InternalItem

	for _, plugin := range a.plugins {
		items, err := plugin.GetItems()

		if err != nil {
			fmt.Printf("Failed to load plugin x")
		} else {
			for _, each := range items {
				catalog = append(catalog, asInternalItem(each, plugin))
			}
		}
	}

	log.Println("Catalog rebuild")
	a.rootItems = catalog
}

func (a *App) clearSuggestions() {
	log.Println("clearSuggestions()")

	a.catalogMutex.Lock()
	a.suggestItems = nil
	a.itemIndex = -1
	a.catalogMutex.Unlock()

	go func() { a.eventChannel <- api.EventSuggestionsChanged }()
}

func (a *App) setSuggestions(suggestions []SuggestItem) {
	log.Println("settingSuggestions(", len(suggestions), ")")

	a.catalogMutex.Lock()
	log.Println(" - acquired lock")

	// Sort items
	sort.SliceStable(suggestions, func(i, j int) bool {
		return suggestions[i].Score > suggestions[j].Score
	})

	a.suggestItems = suggestions

	// If we have any results, select the first item
	if a.itemIndex == -1 && len(suggestions) > 0 {
		a.itemIndex = 0
	}

	a.catalogMutex.Unlock()
	log.Println(" - lock released")

	go func() { a.eventChannel <- api.EventSuggestionsChanged }()
}

func (a *App) addSuggestions(suggestions []SuggestItem) {
	log.Println("addSuggestions(", len(suggestions), ")")
	if len(suggestions) == 0 {
		return
	}

	// Deal with locking
	a.catalogMutex.Lock()

	var result []SuggestItem

	for _, each := range suggestions {
		result = append(result, each)
	}
	for _, each := range a.suggestItems {
		result = append(result, each)
	}

	// Sort items
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	log.Println("Found some extra suggestions")
	a.suggestItems = result

	// If we have any results, select the first item
	if a.itemIndex == -1 && len(result) > 0 {
		a.itemIndex = 0
	}

	a.catalogMutex.Unlock()

	go func() { a.eventChannel <- api.EventSuggestionsChanged }()
}

func (a *App) doShow(w *app.Window) {
	a.isVisible = true

	w.Option(app.Size(unit.Dp(WindowWidth), unit.Dp(SearchBoxHeight)))
	w.Perform(system.ActionCenter)
	a.textInput.Focus()
}

func (a *App) doHide(w *app.Window) {
	a.isVisible = false

	w.Option(Hidden(true))
	a.resetStack()
	a.resetInput()
}

func (a *App) resetStack() {
	a.itemStack = nil
}

func (a *App) run() error {
	w := app.NewWindow(
		app.Decorated(false),
		app.Title("Go Anywhere"),
		Hidden(!a.isVisible),
	)

	// Somehow, on Windows, setting the window size in NewWindow will create a window that is too large
	w.Option(app.Size(unit.Dp(WindowWidth), unit.Dp(SearchBoxHeight)))

	return a.eventLoop(w)
}

func (a *App) registerHotkey() {
	// api.RegisterHotKey should run as a go routine since it locks to the OS thread to allow it to have a fixed
	// thread listening to Windows API messages
	go api.RegisterHotKey(a.hotKey, func() {
		a.eventChannel <- api.EventHotkey
	})
}

func (a *App) pluginByName(k string) (api.Plugin, bool) {
	for _, each := range a.plugins {
		if each.Name() == k {
			return each, true
		}
	}

	return nil, false
}

func (a *App) Quit() {
	a.eventChannel <- api.EventQuit
}

func asInternalItem(each api.Item, plugin api.Plugin) InternalItem {
	return InternalItem{
		Item:       each,
		lookupName: strings.ToLower(each.Label),
		plugin:     plugin,
	}
}
