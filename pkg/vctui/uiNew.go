package vctui

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/vmware/govmomi"

	"github.com/rivo/tview"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

func newVM(c *govmomi.Client) *types.VirtualMachineConfigSpec {
	uiBugFix()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := tview.NewApplication()

	n, _ := netInventory(c)
	var networks []string
	for i := range n {
		networks = append(networks, n[i].Name)
	}

	d, err := storageInventory(c)
	if err != nil {
		log.Fatalf("%v", err)
	}
	var datastores []string
	for i := range d {
		datastores = append(datastores, d[i].Summary.Name)
	}

	h, err := hostInventory(c)
	var hosts []string
	for i := range h {
		hosts = append(hosts, h[i].Summary.Config.Name)
	}

	form := tview.NewForm().
		AddCheckbox("Boot new VM", false, nil).
		AddInputField("VM Name", "", 20, nil, nil).
		AddDropDown("VM Type", vmTypes, 0, nil).
		AddDropDown("Hosts", hosts, 0, nil).
		AddDropDown("Network", networks, 0, nil).
		AddDropDown("Datastore", datastores, 0, nil).
		AddInputField("vCPUs", "", 2, (func(textToCheck string, lastChar rune) bool {
			_, err := strconv.Atoi(textToCheck)
			if err != nil {
				return false
			}
			return true
		}), nil).
		AddInputField("Memory (MB)", "", 6, (func(textToCheck string, lastChar rune) bool {
			_, err := strconv.Atoi(textToCheck)
			if err != nil {
				return false
			}
			return true
		}), nil).
		AddCheckbox("Disk Size (GB)", false, nil).
		AddButton("Save Settings", func() { app.Stop() })

	form.SetBorder(true).
		SetTitle("New Virtual Machine").
		SetTitleAlign(tview.AlignCenter).
		SetRect(2, 2, 60, 22)

	if err := app.SetRoot(form, false).Run(); err != nil {
		panic(err)
	}

	// New Virtual Machine configuration
	var vCPU, Mem int
	var name, guestType, host, network, datastore string

	// Parse the form values
	vCPU, _ = strconv.Atoi(form.GetFormItemByLabel("vCPUs").(*tview.InputField).GetText())
	Mem, _ = strconv.Atoi(form.GetFormItemByLabel("Memory (MB)").(*tview.InputField).GetText())

	name = form.GetFormItemByLabel("VM Name").(*tview.InputField).GetText()
	_, guestType = form.GetFormItemByLabel("VM Type").(*tview.DropDown).GetCurrentOption()
	_, network = form.GetFormItemByLabel("Network").(*tview.DropDown).GetCurrentOption()
	_, datastore = form.GetFormItemByLabel("Datastore").(*tview.DropDown).GetCurrentOption()
	_, host = form.GetFormItemByLabel("Hosts").(*tview.DropDown).GetCurrentOption()

	spec := types.VirtualMachineConfigSpec{
		Name:     name,
		GuestId:  guestType,
		Files:    &types.VirtualMachineFileInfo{VmPathName: fmt.Sprintf("[%s]", datastore)},
		NumCPUs:  int32(vCPU),
		MemoryMB: int64(Mem),
	}

	// Add SCSI controller to the new specification
	scsi, err := object.SCSIControllerTypes().CreateSCSIController("pvscsi")
	if err != nil {
		log.Fatalln("Error creating pvscsi controller as part of new VM")
	}

	spec.DeviceChange = append(spec.DeviceChange, &types.VirtualDeviceConfigSpec{
		Operation: types.VirtualDeviceConfigSpecOperationAdd,
		Device:    scsi,
	})

	i := vcInternal{
		findNetwork:   network,
		findDataStore: datastore,
		findHost:      host,
	}
	err = i.parseInternals(c)
	if err != nil {
		log.Fatalf("%v", err)
	}

	i.dcFolders.VmFolder.CreateVM(ctx, spec, i.resourcePool, i.hostSystem)
	return nil
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
