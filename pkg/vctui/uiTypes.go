package vctui

import (
	"github.com/vmware/govmomi/object"
)

type reference struct {
	objectType    string
	objectDetails string
	vm            *object.VirtualMachine
}

var vmTypes = []string{"otherLinux64Guest",
	"Red Hat Enterprise Linux 6 (64-bit)",
	"Red Hat Enterprise Linux 6 (32-bit)",
	"Red Hat Enterprise Linux 7 (64-bit)",
	"Red Hat Enterprise Linux 7 (32-bit)"}

var deplopyTypes = []string{"reboot",
	"preseed",
	"kickstart"}
