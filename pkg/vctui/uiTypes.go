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
	"ubuntu64Guest",
	"rhel7_64Guest",
	"vmkernel65Guest",
	"windows9_64Guest",
	"vmwarePhoton64Guest"}

var deplopyTypes = []string{"reboot",
	"preseed",
	"kickstart",
	"pull",
	"push"}
