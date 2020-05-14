package vctui

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const (
	powerOn = iota
	powerOff
	guestPowerOff
	guestReboot
	suspend
	reset
	netPowerOn
	diskPowerOn
)

var cancel bool
var app *tview.Application

// RadioButtons implements a simple primitive for radio button selections.
type RadioButtons struct {
	*tview.Box
	options       []string
	currentOption int
}

// NewRadioButtons returns a new radio button primitive.
func NewRadioButtons(options []string) *RadioButtons {
	return &RadioButtons{
		Box:     tview.NewBox(),
		options: options,
	}
}

// Draw draws this primitive onto the screen.
func (r *RadioButtons) Draw(screen tcell.Screen) {
	r.Box.Draw(screen)
	x, y, width, height := r.GetInnerRect()

	for index, option := range r.options {
		if index >= height {
			break
		}
		radioButton := "\u25ef" // Unchecked.
		if index == r.currentOption {
			radioButton = "\u25c9" // Checked.
		}
		line := fmt.Sprintf(`%s[white]  %s`, radioButton, option)
		tview.Print(screen, line, x, y+index, width, tview.AlignLeft, tcell.ColorYellow)
	}
}

// InputHandler returns the handler for this primitive.
func (r *RadioButtons) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return r.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			r.currentOption--
			if r.currentOption < 0 {
				r.currentOption = 0
			}
		case tcell.KeyDown:
			r.currentOption++
			if r.currentOption >= len(r.options) {
				r.currentOption = len(r.options) - 1
			}
		case tcell.KeyEnter:
			app.Stop()
		case tcell.KeyCtrlC:
			app.Stop()
			cancel = true
		}

	})
}

func powerui() int {

	uiBugFix()

	app = tview.NewApplication()
	cancel = false
	radioButtons := NewRadioButtons([]string{"Power On", "Power Off", "Guest Power Off (tools required)", "Guest Reboot (tools required)", "Suspend", "Reset", "PXE Boot (one-shot)", "Power On (Disk Boot)"})

	radioButtons.SetBorder(true).
		SetTitle("Set the power state for this VM").
		SetRect(20, 5, 40, 10)

	if err := app.SetRoot(radioButtons, false).Run(); err != nil {
		panic(err)
	}
	// Return a value outside of the constants
	if cancel == true {
		return -1
	}

	return radioButtons.currentOption
}
