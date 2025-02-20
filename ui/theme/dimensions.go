package theme

// Dimensions определяет размеры и отступы для UI элементов
var Dimensions = struct {
	// Основные отступы
	PaddingX int
	PaddingY int

	// Размеры диалогов
	DialogMinWidth  int
	DialogMinHeight int

	// Размеры кнопок
	ButtonWidth  int
	ButtonHeight int

	// Отступы между элементами
	ElementSpacing int

	// Высота командной строки
	CommandHeight int
}{
	PaddingX:        1,
	PaddingY:        1,
	DialogMinWidth:  60,
	DialogMinHeight: 8,
	ButtonWidth:     10,
	ButtonHeight:    2,
	ElementSpacing:  2,
	CommandHeight:   3,
}
