package vctui

import (
	"fmt"

	"github.com/rivo/tview"
)

func newVM() {
	uiBugFix()
	app := tview.NewApplication()

	form := tview.NewForm().
		AddCheckbox("Update on starting katbox", false, nil).
		AddDropDown("Editor", vmTypes, 0, nil).
		AddInputField("(optional) Custom editor Path", "", 30, nil, nil).
		AddInputField("Git clone path", "", 30, nil, nil).
		AddCheckbox("Open URLs in Browser", false, nil).
		AddButton("Save Settings", func() { app.Stop() })

	form.SetBorder(true).SetTitle("New Virtual Machine").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).Run(); err != nil {
		panic(err)
	}
}

func newVMFromTemplate(template string) {
	uiBugFix()
	app := tview.NewApplication()

	form := tview.NewForm().
		AddCheckbox("Update on starting katbox", false, nil).
		AddDropDown("Editor", vmTypes, 0, nil).
		AddInputField("(optional) Custom editor Path", "", 30, nil, nil).
		AddInputField("Git clone path", "", 30, nil, nil).
		AddCheckbox("Open URLs in Browser", false, nil).
		AddButton("Save Settings", func() { app.Stop() })

	form.SetBorder(true).SetTitle(fmt.Sprintf("New Virtual Machine from template: %s", template)).SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).Run(); err != nil {
		panic(err)
	}
}
