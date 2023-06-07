package app

import (
	"strings"
	"testing"

	"github.com/jensgreen/dux/app/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_BoxDraw(t *testing.T) {
	screen := testutil.InitSimScreen(t, 4, 4)

	box := NewBox()
	box.SetView(screen)
	box.Draw()

	got := testutil.ScreenToString(screen)
	want := strings.TrimSpace(`
┌──┐
│  │
│  │
└──┘
`)

	assert.Equal(t, want, got, "output differs")
}

func Test_BoxDrawHeight1(t *testing.T) {
	screen := testutil.InitSimScreen(t, 4, 1)

	box := NewBox()
	box.SetView(screen)
	box.Draw()

	got := testutil.ScreenToString(screen)
	want := "┌──┐"

	assert.Equal(t, want, got, "output differs")
}
