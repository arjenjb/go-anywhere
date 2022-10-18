package main

import (
	_ "embed"

	"gioui.org/font/opentype"
	"gioui.org/text"
)

//go:embed assets/fonts/OpenSans-Regular.ttf
var fontOpenSans []byte

func GetFont() text.FontFace {
	face, _ := opentype.Parse(fontOpenSans)

	fnt := text.Font{}
	fnt.Typeface = "Open Sans"

	return text.FontFace{Font: fnt, Face: face}
}
