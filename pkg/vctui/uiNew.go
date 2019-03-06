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

	// Get inventories from VMware vCenter
	h, err := hostInventory(c)
	n, _ := netInventory(c)
	d, err := storageInventory(c)

	// Arrays to hold the resolved names of the objects
	var hosts, networks, datastores []string

	for i := range h {
		hosts = append(hosts, h[i].Summary.Config.Name)
	}

	// Function used to ensure that only numbers are entered into a field where expected
	var intCheck func(string, rune) bool
	intCheck = func(textToCheck string, lastChar rune) bool {
		_, err := strconv.Atoi(textToCheck)
		if err != nil {
			return false
		}
		return true
	}

	form := tview.NewForm()
	form.AddCheckbox("Boot new VM", false, nil).
		AddInputField("VM Name", "", 20, nil, nil).
		AddDropDown("VM Type", vmTypes, 0, nil).
		AddDropDown("Hosts", hosts, 0, func(option string, optionIndex int) {
			// Clear original arrays
			networks = nil
			datastores = nil

			// Build new Network array, this matches the references of all networks to references the host is mapped to
			for y := range h[optionIndex].Network {
				for i := range n {
					if h[optionIndex].Network[y].String() == n[i].Summary.GetNetworkSummary().Network.String() {
						networks = append(networks, n[i].Name)
					}
				}
			}
			// Update the Network dropdown with the new array
			form.GetFormItemByLabel("Network").(*tview.DropDown).SetOptions(networks, nil)

			// Build new datastore array, this matches the references of all datastore to references the host is mapped to
			for y := range h[optionIndex].Datastore {
				for i := range d {
					if h[optionIndex].Datastore[y].String() == d[i].Summary.Datastore.String() {
						datastores = append(datastores, d[i].Summary.Name)
					}
				}
			}

			//Update the datastore with the new array
			form.GetFormItemByLabel("Datastore").(*tview.DropDown).SetOptions(datastores, nil)
		}).
		AddDropDown("Network", networks, 0, nil).
		AddDropDown("Datastore", datastores, 0, nil).
		AddInputField("vCPUs", "", 2, intCheck, nil).
		AddInputField("Memory (MB)", "", 6, intCheck, nil).
		AddInputField("Disk Size (GB)", "", 6, intCheck, nil).
		AddButton("Save Settings", func() { app.Stop() })

	form.SetBorder(true).
		SetTitle("New Virtual Machine").
		SetTitleAlign(tview.AlignCenter).
		SetRect(5, 1, 60, 23)

	if err := app.SetRoot(form, false).Run(); err != nil {
		panic(err)
	}

	// New Virtual Machine configuration
	var vCPU, mem, diskSize int
	var name, guestType, host, network, datastore string

	// Parse the form values
	vCPU, _ = strconv.Atoi(form.GetFormItemByLabel("vCPUs").(*tview.InputField).GetText())
	mem, _ = strconv.Atoi(form.GetFormItemByLabel("Memory (MB)").(*tview.InputField).GetText())
	diskSize, _ = strconv.Atoi(form.GetFormItemByLabel("Disk Size (GB)").(*tview.InputField).GetText())

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
		MemoryMB: int64(mem),
	}

	// TODO - ALL Code below needs moving to a seperate func()

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

	// Create the new Virtual Machine
	task, err := i.dcFolders.VmFolder.CreateVM(ctx, spec, i.resourcePool, i.hostSystem)

	if err != nil {
		log.Fatalln("Creating new VM failed, more detail can be found in vCenter tasks")
	}

	// Wait for the results of vCenter parsing
	info, err := task.WaitForResult(ctx, nil)
	if err != nil {
		log.Fatalf("Creating new VM failed\n%v", err)
	}

	// Retrieve the new VM
	vm := object.NewVirtualMachine(c.Client, info.Result.(types.ManagedObjectReference))

	// Modify it's configuration
	backing, err := i.network.EthernetCardBackingInfo(ctx)
	if err != nil {
		log.Fatalf("Unable to determine vCenter network backend\n%v", err)
	}

	netdev, err := object.EthernetCardTypes().CreateEthernetCard("vmxnet3", backing)
	if err != nil {
		log.Fatalf("Unable to create vmxnet3 network interface\n%v", err)
	}

	var add []types.BaseVirtualDevice
	add = append(add, netdev)

	devices, err := vm.Device(ctx)
	if err != nil {
		log.Fatalf("Unable to read devices from VM configuration\n%v", err)
	}

	controller, err := devices.FindDiskController("scsi")
	if err != nil {
		log.Fatalf("Unable to find SCSI device from VM configuration\n%v", err)
	}

	disk := devices.CreateDisk(controller, i.datastore.Reference(), i.datastore.Path(fmt.Sprintf("%s/%s.vmdk", name, name)))
	disk.CapacityInKB = int64(diskSize * 1024 * 1024)
	add = append(add, disk)

	if vm.AddDevice(ctx, add...); err != nil {
		log.Fatalf("Unable to add new storage device to VM configuration\n%v", err)
	}

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
