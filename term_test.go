package dux

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/golang/geo/r1"
	"github.com/golang/geo/r2"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/z2"
)

const (
	TermHeight = 24
	TermWidth  = 80
)

func getOutput(t *testing.T, events []StateUpdate) string {
	s := tcell.NewSimulationScreen("utf-8")
	s.SetSize(TermWidth, TermHeight)
	err := s.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan StateUpdate)
	go func() {
		for _, ev := range events {
			ch <- ev
		}
		close(ch)
	}()
	tui := NewTerminalView(s, ch, nil)
	for i := 0; i < len(events); i++ {
		tui.poll()
	}

	w := bytes.Buffer{}
	err = printScreen(&w, s)
	if err != nil {
		t.Fatal(err)
	}

	return w.String()
}

func printScreen(writer io.Writer, s tcell.SimulationScreen) error {
	cells, w, _ := s.GetContents()
	for i, cell := range cells {
		for _, char := range cell.Runes {
			_, err := fmt.Fprintf(writer, "%c", char)
			if err != nil {
				return err
			}
		}
		if (i+1)%w == 0 {
			_, err := fmt.Fprint(writer, "\n")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Test_OutputContainsNameAndSize(t *testing.T) {
	events :=
		[]StateUpdate{
			{
				State: State{

					StatusbarSpazeZ2: z2.Rect{
						X: z2.Interval{Lo: 0, Hi: TermWidth},
						Y: z2.Interval{Lo: 0, Hi: 1},
					},
					TreemapSpaceR2: r2.Rect{
						X: r1.Interval{Lo: 0, Hi: TermWidth},
						Y: r1.Interval{Lo: 0, Hi: TermHeight - 1},
					},
					Treemap: Treemap{
						File: files.File{Path: "root"},
						Children: []Treemap{
							{
								File: files.File{Path: "root/foo", Size: 99},
								Rect: r2.Rect{
									X: r1.Interval{Lo: 0, Hi: TermWidth / 3},
									Y: r1.Interval{Lo: 0, Hi: TermHeight},
								},
								Children: nil,
							},
							{
								File: files.File{Path: "root/bar", Size: 2},
								Rect: r2.Rect{
									X: r1.Interval{Lo: TermWidth / 3, Hi: TermWidth - 1},
									Y: r1.Interval{Lo: 0, Hi: TermHeight},
								},
								Children: nil,
							},
						},
					},
				},
			},
		}
	got := getOutput(t, events)
	fmt.Println("Console output start >>>>>")
	fmt.Printf("%v", got)
	fmt.Println("<<<<< End of console ouput")

	tests := []string{
		"foo 99",
		"bar 2",
	}
	for _, tt := range tests {
		if !strings.Contains(got, tt) {
			t.Errorf("'%v' not in output", tt)
		}
	}
}
