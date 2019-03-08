package vctui

import "github.com/rivo/tview"

func errorUI(err error) {
	uiBugFix()
	app := tview.NewApplication()
	button := tview.NewButton(err.Error()).SetSelectedFunc(func() {
		app.Stop()
	})
	button.SetBorder(true).SetRect(5, 5, 75, 3)
	if err := app.SetRoot(button, false).Run(); err != nil {
		panic(err)
	}
}
