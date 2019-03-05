package vctui

import (
	"regexp"

	"github.com/rivo/tview"
	"github.com/vmware/govmomi/object"
)

//SearchUI - this will provide a user interface for finding articles based upon a keyword entered in the UI
func SearchUI(v []*object.VirtualMachine) (string, []*object.VirtualMachine) {

	uiBugFix()

	title := "Search"
	label := "Search string (RegEx)"
	for {
		app := tview.NewApplication()

		form := tview.NewForm().
			AddInputField(label, "", 30, nil, nil).
			AddButton("Search", func() { app.Stop() })

		form.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignLeft)

		if err := app.SetRoot(form, true).SetFocus(form).Run(); err != nil {
			panic(err)
		}

		searchString := form.GetFormItemByLabel(label).(*tview.InputField).GetText()

		newVMs, err := searchVMS(searchString, v)
		if err == nil {
			return searchString, newVMs
		}
		title = err.Error()
	}
}

func searchVMS(searchString string, v []*object.VirtualMachine) ([]*object.VirtualMachine, error) {

	var newVMList []*object.VirtualMachine
	var err error
	for x := range v {
		matched, err := regexp.MatchString(searchString, v[x].Name())
		if err != nil {
			break
		}
		// If the regex matches then add it to the new subset
		if matched == true {
			newVMList = append(newVMList, v[x])
		}
	}
	if err == nil {
		return newVMList, nil
	}
	return nil, err
}
