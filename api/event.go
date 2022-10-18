package api

const (
	EventHotkey = iota
	EventShow
	EventHide
	EventSuggestionsChanged
	EventQuit
)

type Event int8
