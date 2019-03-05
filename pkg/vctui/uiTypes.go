package vctui

import (
	"github.com/vmware/govmomi/object"
)

type reference struct {
	objectType string
	vm         *object.VirtualMachine
}
