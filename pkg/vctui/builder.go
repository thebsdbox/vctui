package vctui

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// This function will take the full article set and build a tree from any search parameters
func buildTree(v []*object.VirtualMachine) *tview.TreeNode {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// reference is used to label the type of tree Node
	var r reference

	// Begin the UI Tree
	rootDir := "VMware vCenter"
	root := tview.NewTreeNode(rootDir).
		SetColor(tcell.ColorWhite).SetReference(r)

	// Add Github articles to the tree
	vmNode := tview.NewTreeNode("VMs").SetReference(r).SetSelectable(true)
	vmNode.SetColor(tcell.ColorYellow)

	// Add Github articles to the tree
	templateNode := tview.NewTreeNode("Templates").SetReference(r).SetSelectable(true)
	templateNode.SetColor(tcell.ColorYellow)

	for x := range v {
		// Set the reference to point back to the VM object
		r.vm = v[x]
		// Create the Virtual Machine child node
		vmChildNode := tview.NewTreeNode(v[x].Name()).SetSelectable(true).SetExpanded(false)

		// Retrieve the managed object (using the summary string)
		var o mo.VirtualMachine

		err := v[x].Properties(ctx, v[x].Reference(), []string{"summary", "snapshot"}, &o)
		if err != nil {
			break
		}

		powerstate, err := v[x].PowerState(ctx)
		if err != nil {
			vmChildNode.SetColor(tcell.ColorGray)
		}
		switch powerstate {
		case types.VirtualMachinePowerStatePoweredOff:
			vmChildNode.SetColor(tcell.ColorRed)

		case types.VirtualMachinePowerStatePoweredOn:
			vmChildNode.SetColor(tcell.ColorGreen)

		case types.VirtualMachinePowerStateSuspended:
			vmChildNode.SetColor(tcell.ColorGray)

		}

		// Build out the details of the virtual machine
		vmDetails := buildDetails(ctx, v[x], o)
		vmSnapshots := buildSnapshots(ctx, v[x], o)
		// Add as child nodes to the virtual machine node
		vmChildNode.AddChild(vmDetails)
		vmChildNode.AddChild(vmSnapshots)

		// Set the object type and pin the reference to the new node
		if o.Summary.Config.Template == true {
			r.objectType = "template"
			vmChildNode.SetReference(r)
			templateNode.AddChild(vmChildNode)
		} else {
			r.objectType = "Virtual Machines"
			vmChildNode.SetReference(r)
			vmNode.AddChild(vmChildNode)
		}
	}

	root.AddChild(vmNode)
	root.AddChild(templateNode)

	return root
}

func buildDetails(ctx context.Context, vm *object.VirtualMachine, vmo mo.VirtualMachine) *tview.TreeNode {

	// Add Details subtree information
	vmDetails := tview.NewTreeNode("Details").SetReference("Details").SetSelectable(true)

	vmDetail := tview.NewTreeNode(fmt.Sprintf("CPUs: %d", vmo.Summary.Config.NumCpu)).SetReference("Cpu").SetSelectable(true)
	vmDetails.AddChild(vmDetail)

	vmDetail = tview.NewTreeNode(fmt.Sprintf("Memory: %d", vmo.Summary.Config.MemorySizeMB)).SetReference("memory").SetSelectable(true)
	vmDetails.AddChild(vmDetail)

	vmDetail = tview.NewTreeNode(fmt.Sprintf("Type: %s", vmo.Summary.Config.GuestFullName)).SetReference("memory").SetSelectable(true)
	vmDetails.AddChild(vmDetail)

	vmDetail = tview.NewTreeNode(fmt.Sprintf("VMware Tools: %s", vmo.Summary.Guest.ToolsStatus)).SetReference("toolsStatus").SetSelectable(true)
	vmDetails.AddChild(vmDetail)

	vmDetail = tview.NewTreeNode(fmt.Sprintf("VM IP Address: %s", vmo.Summary.Guest.IpAddress)).SetReference("toolsStatus").SetSelectable(true)
	vmDetails.AddChild(vmDetail)

	devices, _ := vm.Device(ctx)

	vmDetail = tview.NewTreeNode(fmt.Sprintf("MAC ADDRESS: %s", devices.PrimaryMacAddress())).SetReference("toolsStatus").SetSelectable(true)
	vmDetails.AddChild(vmDetail)

	return vmDetails
}

func buildSnapshots(ctx context.Context, vm *object.VirtualMachine, vmo mo.VirtualMachine) *tview.TreeNode {
	// Add Snapshots subtree information
	vmSnapshots := tview.NewTreeNode("snapshots").SetReference("snapshots").SetSelectable(true)
	var r reference
	r.objectType = "snapshot"
	r.vm = vm
	if vmo.Snapshot != nil {
		if len(vmo.Snapshot.RootSnapshotList) != 0 {
			for i := range vmo.Snapshot.RootSnapshotList {
				vmSnapshot := tview.NewTreeNode(vmo.Snapshot.RootSnapshotList[i].Name).SetReference(r).SetSelectable(true)
				vmSnapshots.AddChild(vmSnapshot)
			}
		}
	}
	return vmSnapshots
}
