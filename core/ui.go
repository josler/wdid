package core

import (
	"context"
	"fmt"
	"log"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func RenderUI(ctx context.Context) {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	store := ctx.Value("store").(Store)
	items, _ := listFromFilters(store, "time=week", false)

	// builder := bytes.Buffer{}
	rows := []string{}
	for _, item := range items {
		rows = append(rows, fmt.Sprintf("[%s](fg:blue) - %s", item.Status(), item.Data()))
		// fmt.Fprintf(&builder, "[%s](fg:blue)", item.Data())
	}

	// p := widgets.NewParagraph()
	// p.Text = builder.String()
	// p.SetRect(0, 0, 25, 5)

	l := widgets.NewList()
	l.Title = "List"
	l.Rows = rows
	// l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	// l.SelectedRowStyle.Fg = ui.ColorClear
	l.SelectedRowStyle = ui.StyleClear
	l.SetRect(0, 0, 120, 20)

	ui.Render(l)
	// for e := range ui.PollEvents() {
	// 	if e.Type == ui.KeyboardEvent {
	// 		break
	// 	}
	// }

	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "j", "<Down>":
			l.ScrollDown()
		case "k", "<Up>":
			l.ScrollUp()
		case "<C-d>":
			l.ScrollHalfPageDown()
		case "<C-u>":
			l.ScrollHalfPageUp()
		case "<C-f>":
			l.ScrollPageDown()
		case "<C-b>":
			l.ScrollPageUp()
		case "g":
			if previousKey == "g" {
				l.ScrollTop()
			}
		case "<Home>":
			l.ScrollTop()
		case "G", "<End>":
			l.ScrollBottom()
		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		ui.Render(l)
	}
}
