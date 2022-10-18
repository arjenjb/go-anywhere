package api

const (
	EventHotkey = iota
	EventShow
	EventHide
	EventSuggestionsChanged
)

type Event int8
