package api

import (
	"context"
	"image"
)

type ItemCategory int
type ItemArgsHint int
type ItemHitHint int
type Match int

const (
	MatchAny Match = iota
	MatchFuzzy
)

const (
	Error ItemCategory = iota
	Keyword
	File
	Url
	User = 1000
)

const (
	Forbidden ItemArgsHint = iota
	Accepted
	Required
)

type Item struct {
	Label       string
	Description string
	Category    ItemCategory
	Target      string
	Data        interface{}
	Icon        *image.Image
	ArgsHint    ItemArgsHint
}

func (i Item) String() string {
	return i.Label
}

type SuggestionCallback func([]Item, Match)

type Plugin interface {
	Initialize()
	Catalog() error
	Icon() *image.Image
	GetItems() ([]Item, error)
	Suggest(ctx context.Context, input string, chain []Item, callback SuggestionCallback)
	Execute(Item)
	Name() string

	// LoadConfig receives a loader function that when invoked loads the configuration in the struct pointer `config` that it's passed
	LoadConfig(func(interface{}) error)
}
