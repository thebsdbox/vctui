package vctui

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

type vcInternal struct {
	findDataStore string
	findNetwork   string
	findHost      string
	datastore     *object.Datastore
	dcFolders     *object.DatacenterFolders
	hostSystem    *object.HostSystem
	network       object.NetworkReference
	resourcePool  *object.ResourcePool
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
	i.datastore, err = f.DatastoreOrDefault(ctx, i.findDataStore)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	i.dcFolders, err = dc.Folders(ctx)
	if err != nil {
		return fmt.Errorf("Error locating default datacenter folder")
	}

	// Set the host that the VM will be created on
	i.hostSystem, err = f.HostSystemOrDefault(ctx, i.findHost)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	// Find the resource pool attached to this host
	i.resourcePool, err = i.hostSystem.ResourcePool(ctx)
	if err != nil {
		return fmt.Errorf("Error locating default resource pool")
	}

	i.network, err = f.NetworkOrDefault(ctx, i.findNetwork)
	if err != nil {
		return fmt.Errorf("Network could not be found")
	}

	return nil
}

func netInventory(c *govmomi.Client) (networks []mo.Network, err error) {
	// Create a view of Network types
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := view.NewManager(c.Client)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Network"}, true)
	if err != nil {
		log.Fatal(err)
	}

	defer v.Destroy(ctx)

	err = v.Retrieve(ctx, []string{"Network"}, nil, &networks)
	return
}

func storageInventory(c *govmomi.Client) (dss []mo.Datastore, err error) {
	// Create a view of Datastore objects
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := view.NewManager(c.Client)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datastore"}, true)
	if err != nil {
		log.Fatal(err)
	}

	defer v.Destroy(ctx)

	err = v.Retrieve(ctx, []string{"Datastore"}, []string{"summary"}, &dss)
	return
}
func hostInventory(c *govmomi.Client) (hss []mo.HostSystem, err error) {
	// Create a view of Datastore objects
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := view.NewManager(c.Client)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		log.Fatal(err)
	}

	defer v.Destroy(ctx)

	err = v.Retrieve(ctx, []string{"HostSystem"}, nil, &hss)
	return
}

//VMInventory will create an inventory
func VMInventory(c *govmomi.Client, sortVMs bool) ([]*object.VirtualMachine, error) {

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

	if sortVMs == true {
		// Sort function to sort by name
		sort.Slice(vms, func(i, j int) bool {
			switch strings.Compare(vms[i].Name(), vms[j].Name()) {
			case -1:
				return true
			case 1:
				return false
			}
			return vms[i].Name() > vms[j].Name()
		})
	}

	return vms, nil
}
