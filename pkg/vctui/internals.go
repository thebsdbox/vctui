package vctui

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

type vcInternal struct {
	datastore    *object.Datastore
	dcFolders    *object.DatacenterFolders
	hostSystem   *object.HostSystem
	network      object.NetworkReference
	resourcePool *object.ResourcePool
}

func (i *vcInternal) parseInternals(c *govmomi.Client) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new finder that will discover the defaults and are looked for Networks/Datastores
	f := find.NewFinder(c.Client, true)

	// Find one and only datacenter, not sure how VMware linked mode will work
	dc, err := f.DatacenterOrDefault(ctx, "")
	if err != nil {
		return fmt.Errorf("No Datacenter instance could be found inside of vCenter %v", err)
	}

	// Make future calls local to this datacenter
	f.SetDatacenter(dc)

	// Find Datastore/Network
	i.datastore, err = f.DatastoreOrDefault(ctx, "")
	if err != nil {
		return fmt.Errorf("Datastore [%s], could not be found", "")
	}

	i.dcFolders, err = dc.Folders(ctx)
	if err != nil {
		return fmt.Errorf("Error locating default datacenter folder")
	}

	// Set the host that the VM will be created on
	i.hostSystem, err = f.HostSystemOrDefault(ctx, "")
	if err != nil {
		return fmt.Errorf("No vSphere hosts could be found")
	}

	// Find the resource pool attached to this host
	i.resourcePool, err = i.hostSystem.ResourcePool(ctx)
	if err != nil {
		return fmt.Errorf("Error locating default resource pool")
	}
	return nil
}

//VMInventory will create an inventory
func VMInventory(c *govmomi.Client) ([]*object.VirtualMachine, error) {

	ctx := context.Background()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new finder that will discover the defaults and are looked for Networks/Datastores
	f := find.NewFinder(c.Client, true)

	// Find one and only datacenter, not sure how VMware linked mode will work
	dc, err := f.DatacenterOrDefault(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("No Datacenter instance could be found inside of vCenter %v", err)
	}

	// Make future calls local to this datacenter
	f.SetDatacenter(dc)

	vms, err := f.VirtualMachineList(ctx, "*")

	return vms, nil
}
