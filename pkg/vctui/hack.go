package vctui

import (
	"github.com/micmonay/keybd_event"
)

// uiBugFix - This is used to fix the issue with tcell dropping a keystroke between new tcell.screens being created
// TODO (thebsdbox) remove this when tcell issue #194 is fixed
func uiBugFix() {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		return
	}

	//set keys
	kb.SetKeys(keybd_event.VK_SPACE)

	//launch
	kb.Launching()
	//fmt.Printf("\033[2J")
}
