package main

import (
	"image"
	"image/color"
	"log"
	"time"

	"go-keyboard-launcher/api"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type C = layout.Context
type D = layout.Dimensions

func Hidden(isHidden bool) app.Option {
	return func(m unit.Metric, cnf *app.Config) {
		if isHidden {
			cnf.Mode = app.Hidden
		}
	}
}

// Styling
var colorBackground = color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xff}
var colorBorder = color.NRGBA{R: 0x50, G: 0x50, B: 0x50, A: 255}
var colorHighlightItem = color.NRGBA{R: 0x2e, G: 0x64, B: 0x70, A: 0xff}

//var colorText = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
var colorText = color.NRGBA{R: 0xdd, G: 0xDD, B: 0xDD, A: 0xff}
var colorDescriptionText = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x33}

const ItemHeight = 26
const SearchFontSize = 16
const ItemFontSize = 12

func (a *App) eventLoop(w *app.Window) error {
	centerOnce := true

	fonts := []text.FontFace{GetFont()}

	theme := material.NewTheme(fonts)
	theme.TextSize = unit.Sp(ItemFontSize)

	var ops op.Ops

	for {
		select {
		case evt := <-a.eventChannel:
			switch evt {
			case api.EventHotkey:
				a.doShow(w)

			case api.EventShow:
				a.doShow(w)

			case api.EventHide:
				a.doHide(w)

			case api.EventSuggestionsChanged:
				w.Invalidate()

			case api.EventQuit:
				return nil
			}

		case evt := <-w.Events():
			switch e := evt.(type) {
			case system.DestroyEvent:
				return e.Err

			case system.FrameEvent:
				var gtx = layout.NewContext(&ops, e)

				// Skip rendering if we're not supposed to be visible
				if !a.isVisible {
					e.Frame(gtx.Ops)
					continue
				}

				// React to events
				a.handleInputEvents(gtx)

				newItems := Min(len(a.suggestItems), 10)
				if a.lastVisibleItems != newItems {
					log.Printf("Resizing windows\n")
					a.lastVisibleItems = newItems
					w.Option(app.Size(unit.Dp(WindowWidth), unit.Dp(SearchBoxHeight+newItems*ItemHeight)))
				}

				// Background
				paint.Fill(&ops, colorBackground)

				root := layout.Flex{
					Axis: layout.Vertical,
				}

				root.Layout(gtx, layout.Rigid(
					func(gtx C) D {
						return drawInputBox(gtx, a, theme)
					},
				), layout.Flexed(1,
					func(gtx C) D {
						dim := material.List(theme, &a.listWidget).Layout(gtx, len(a.suggestItems),
							func(gtx C, index int) D {
								gtx.Constraints.Max.Y = ItemHeight
								gtx.Constraints.Min.Y = ItemHeight
								item := a.suggestItems[index].Item
								return drawItem(gtx, theme, item, index == a.suggestItemIndex)
							},
						)

						// Pass through mouse events so the list still receives scrolling events
						defer pointer.PassOp{}.Push(gtx.Ops).Pop()

						// Confine click listener to dimensions of the list
						defer clip.Rect{Max: dim.Size}.Push(gtx.Ops).Pop()

						pointer.InputOp{
							Tag:   &a.listWidget,
							Types: pointer.Press,
						}.Add(gtx.Ops)

						return dim
					}))

				if centerOnce {
					w.Perform(system.ActionCenter)
					centerOnce = false
				}

				e.Frame(gtx.Ops)
			}
		}
	}
}

func (a *App) handleInputEvents(gtx layout.Context) {
	// Keyboard events
	for _, event := range a.textInput.Events() {
		if _, ok := event.(widget.ChangeEvent); ok {
			a.Search(a.textInput.Text())
		}
	}

	// Mouse events
	for _, e := range gtx.Events(&a.listWidget) {
		if e, ok := e.(pointer.Event); ok {
			switch e.Type {
			// Handle single and double-clicks on list items
			case pointer.Press:
				lastIndex := a.suggestItemIndex
				a.suggestItemIndex = a.listWidget.Position.First + int(e.Position.Y/24)

				clickDelta := e.Time - a.lastClickTime
				a.lastClickTime = e.Time

				if lastIndex == a.suggestItemIndex && clickDelta < time.Millisecond*500 {
					// It's a double click, on the same item
					a.enter()
				}
			}
		}
	}
}

func drawInputBox(gtx C, a *App, th *material.Theme) D {
	border := widget.Border{
		Color:        colorBorder,
		CornerRadius: unit.Dp(0),
		Width:        unit.Dp(2),
	}

	return border.Layout(gtx,
		func(gtx C) D {
			if len(a.itemStack) > 0 {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						// Draw a stack
						return layout.Stack{Alignment: layout.W}.Layout(gtx,
							// First fill the background
							layout.Expanded(func(gtx layout.Context) layout.Dimensions {
								defer clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops).Pop()
								paint.Fill(gtx.Ops, colorBorder)
								return layout.Dimensions{Size: gtx.Constraints.Min}
							}),

							// Draw the label on top
							layout.Stacked(func(gtx layout.Context) layout.Dimensions {
								label := material.Label(th, unit.Sp(16), a.itemStack[len(a.itemStack)-1].item.DisplayName())
								label.Color = colorText

								return drawInset(gtx, func(gtx C) D {
									return label.Layout(gtx)
								})
							}),
						)
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return drawInputTextField(gtx, a, th)
					}),
				)
			} else {
				return drawInputTextField(gtx, a, th)
			}
		})
}

func drawInputTextField(gtx C, a *App, th *material.Theme) layout.Dimensions {
	return drawInset(gtx, func(gtx C) D {
		return a.layoutEditor(gtx, th)
	})
}

func drawInset(gtx C, f func(gtx C) D) layout.Dimensions {
	return layout.UniformInset(unit.Dp(5)).Layout(gtx, f)
}

func drawItem(gtx C, th *material.Theme, item InternalItem, isHighlighted bool) D {
	itemLabel := material.Label(th, unit.Sp(ItemFontSize), item.DisplayName())
	itemLabel.Alignment = text.Start
	itemLabel.Color = colorText
	//itemLabel.Font.Weight = text.Thin
	itemLabel.MaxLines = 1

	descriptionLabel := material.Label(th, unit.Sp(ItemFontSize), item.Description())
	descriptionLabel.MaxLines = 1
	descriptionLabel.Alignment = text.Start
	descriptionLabel.Color = colorDescriptionText
	//descriptionLabel.Font.Weight = text.Thin

	if isHighlighted {
		ColorBox(gtx, gtx.Constraints.Max, colorHighlightItem)
	} else {
		ColorBox(gtx, gtx.Constraints.Max, colorBackground)
	}

	dim := layout.UniformInset(unit.Dp(2)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Spacing:   layout.SpaceEnd,
			Alignment: layout.Middle,
		}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Draw the icon, if any
			if icon := item.Icon(); icon != nil {
				stack := op.Offset(image.Point{
					X: 3,
					Y: 0,
				}).Push(gtx.Ops)

				imageOp := paint.NewImageOp(*icon)
				imageOp.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)

				stack.Pop()
			}

			return layout.Dimensions{
				Size: image.Point{
					X: 16,
					Y: 16,
				},
			}
		}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// The label with some offset
			op.Offset(image.Point{X: 7, Y: 1}).Add(gtx.Ops)
			return itemLabel.Layout(gtx)
		}), layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			// The description label
			return layout.Inset{
				Top:    1,
				Bottom: 0,
				Left:   12,
				Right:  0,
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return descriptionLabel.Layout(gtx)
			})
		}))
	})

	return dim
}

func (a *App) layoutEditor(gtx C, th *material.Theme) D {
	editor := material.Editor(th, &a.textInput, "")
	editor.Color = colorText
	editor.TextSize = SearchFontSize

	key.InputOp{Tag: &a.eventKey, Keys: "⎋|↓|↑|⏎|⌤|Tab|⌫"}.Add(gtx.Ops)

	for _, e := range gtx.Events(&a.eventKey) {
		switch ev := e.(type) {
		case key.Event:
			if ev.State == key.Press {
				switch {
				case ev.Name == key.NameReturn:
					a.enter()
				case ev.Name == key.NameEscape:
					a.cancel()
				case ev.Name == key.NameDownArrow:
					a.itemDown()
				case ev.Name == key.NameUpArrow:
					a.itemUp()
				case ev.Name == key.NameTab:
					a.itemSuggest()
				case ev.Name == key.NameDeleteBackward:
					if len(a.itemStack) > 0 {
						a.cancel()
					}
				}
			}
		}
	}

	return editor.Layout(gtx)
}
