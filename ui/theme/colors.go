package theme

import "github.com/jroimartin/gocui"

// Colors определяет цвета для различных элементов UI
var Colors = struct {
	// Основные цвета
	DefaultFg gocui.Attribute
	DefaultBg gocui.Attribute

	// Цвета выделения
	SelectionFg gocui.Attribute
	SelectionBg gocui.Attribute

	// Цвета кнопок
	ButtonConfirmFg gocui.Attribute
	ButtonCancelFg  gocui.Attribute

	// Цвета статусов
	SuccessFg gocui.Attribute
	ErrorFg   gocui.Attribute
	WarnFg    gocui.Attribute
}{
	DefaultFg:       gocui.ColorWhite,
	DefaultBg:       gocui.ColorBlack,
	SelectionFg:     gocui.ColorBlack,
	SelectionBg:     gocui.ColorGreen,
	ButtonConfirmFg: gocui.ColorGreen,
	ButtonCancelFg:  gocui.ColorRed,
	SuccessFg:       gocui.ColorGreen,
	ErrorFg:         gocui.ColorRed,
	WarnFg:          gocui.ColorYellow,
}
