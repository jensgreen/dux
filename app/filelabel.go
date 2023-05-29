package app

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/files"
)

type FileLabel struct {
	file       files.File
	isRoot     bool
	isSelected bool

	view     views.View
	nameText *views.Text
	sizeText *views.Text

	views.WidgetWatchers
}

func (fl *FileLabel) Draw() {
	fl.nameText.Draw()
	// sizeText won't be drawn on insufficient width; its view will be nil
	fl.sizeText.Draw()
}

func (fl *FileLabel) Resize() {
	fl.nameText.Resize()
	fl.sizeText.Resize()
}

func (fl *FileLabel) HandleEvent(ev tcell.Event) bool {
	return false
}

func (fl *FileLabel) SetView(view views.View) {
	fl.view = view
	nameWidth, _ := fl.nameText.Size()
	nameView := views.NewViewPort(view, 0, 0, nameWidth, 1)
	fl.nameText.SetView(nameView)

	sizeTextWidth, _ := fl.sizeText.Size()
	sizeView := views.NewViewPort(view, nameWidth, 0, sizeTextWidth, 1)
	visibleWith, _ := sizeView.Size()
	if visibleWith >= sizeTextWidth {
		fl.sizeText.SetView(sizeView)
	} else {
		fl.sizeText.SetView(nil)
	}
}

func (fl *FileLabel) Size() (int, int) {
	w1, h1 := fl.nameText.Size()
	w2, h2 := fl.sizeText.Size()
	h := h1
	if h2 > h1 {
		h = h2
	}
	return w1 + w2, h
}

func (fl *FileLabel) SetFile(file files.File) {
	fl.file = file
	fl.update()
}

func (fl *FileLabel) Select(selected bool) {
	fl.isSelected = selected
	fl.update()
}

func (fl *FileLabel) SetIsRoot(isRoot bool) {
	fl.isRoot = isRoot
	fl.update()
}

func (fl *FileLabel) update() {
	label := fl.file.Name()
	if fl.isRoot {
		label = fl.file.Path
	}

	style := tcell.StyleDefault
	if fl.isSelected {
		label = string('●') + " " + label
		style = style.Italic(true).Foreground(tcell.ColorYellow).Bold(true)
	}

	if fl.file.IsDir {
		if !strings.HasSuffix(label, "/") {
			// avoid showing for example "/" as "//"
			label += "/"
		}
		if !fl.isSelected {
			style = style.Foreground(tcell.ColorBlue)
		}
	}

	fl.nameText.SetText(label)
	fl.sizeText.SetText(" " + files.HumanizeIEC(fl.file.Size))

	fl.nameText.SetStyle(style)
	if fl.isSelected {
		noItalics := style.Italic(false)
		fl.nameText.SetStyleAt(0, noItalics)
		fl.sizeText.SetStyle(noItalics)
	} else {
		fl.sizeText.SetStyle(tcell.StyleDefault)
	}
}

func NewFileLabel() *FileLabel {
	fl := &FileLabel{
		nameText: views.NewText(),
		sizeText: views.NewText(),
	}
	return fl
}
