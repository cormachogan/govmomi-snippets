////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//
// Description:		Go code to connect to vSphere via environment
// 					variables and retrieve VM information
//
// 					Login moved to function in this example
//
// Author:			Cormac J. Hogan (VMware)
//
// Date:			 25 Jan 2021
//
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"text/tabwriter"

	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
)

func vlogin(ctx context.Context, vc, user, pwd string) (*vim25.Client, error) {

	//
	// Create a vSphere/vCenter client
	//
	//    The govmomi client requires a URL object, u, not just a string representation of the vCenter URL.
	//

	u, err := soap.ParseURL(vc)

	if u == nil {
		fmt.Printf("could not parse URL (environment variables set?)")
	}

	if err != nil {
		fmt.Printf("URL parsing not successful, error %v", err)
		return nil, err
	}

	u.User = url.UserPassword(user, pwd)

	//
	// Ripped from https://github.com/vmware/govmomi/blob/master/examples/examples.go
	//

	// Share govc's session cache
	s := &cache.Session{
		URL:      u,
		Insecure: true,
	}

	c := new(vim25.Client)

	err = s.Login(ctx, c, nil)

	if err != nil {
		fmt.Printf("Log in not successful- could not get vCenter client: %v", err)
		return nil, err
	} else {
		fmt.Printf("Log in successful")

		return c, nil
	}
}

func main() {

	// We need to get 3 environment variables:
	//
	//-- GOVMOMI_URL
	//-- GOVMOMI_USERNAME
	//-- GOVMOMI_PASSWORD

	vc := os.Getenv("GOVMOMI_URL")
	user := os.Getenv("GOVMOMI_USERNAME")
	pwd := os.Getenv("GOVMOMI_PASSWORD")

	fmt.Printf("DEBUG: vc is %s\n", vc)
	fmt.Printf("DEBUG: user is %s\n", user)
	fmt.Printf("DEBUG: password is %s\n", pwd)

	//
	// Imagine that there were multiple operations taking place such as processing some data, logging into vCenter, etc.
	// If one of the operations failed, the context would be used to share the fact that all of the other operations sharing that context needs cancelling.
	//

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//
	// Call the login function
	//

	c, err := vlogin(ctx, vc, user, pwd)

	//
	// Create a view manager
	//

	m := view.NewManager(c)

	//
	// Create a container view of VM objects
	//

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		fmt.Printf("Unable to create Virtual Machine Container View: error %s", err)
		return
	}

	defer v.Destroy(ctx)

	//
	// Retrieve summary property for all machines
	//

	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)

	if err != nil {
		fmt.Printf("Unable to retrieve VM information: error %s", err)
		return
	}

	//
	// Print summary per vm
	//
	// -- https://golang.org/pkg/text/tabwriter/#NewWriter
	//

	tw := tabwriter.NewWriter(os.Stdout, 4, 0, 4, ' ', 0)
	fmt.Printf("\n*** VM Information ***\n")
	fmt.Printf("-----------------------\n\n")
	fmt.Fprintf(tw, "Name\tGuest\tCPU\tCPU Rsv\tMem(MB)\tMem Rsv\tState\tHW Version\tIP Address\tVM Path\n")
	fmt.Fprintf(tw, "----\t-----\t---\t--- ---\t-------\t--- ---\t-----\t-- -------\t-- -------\t-- ----\n")

	for _, vm := range vms {
		fmt.Fprintf(tw, "%s:\t", vm.Summary.Config.Name)
		fmt.Fprintf(tw, "%s\t", vm.Summary.Guest.GuestId)
		fmt.Fprintf(tw, "%v\t", vm.Summary.Config.NumCpu)
		fmt.Fprintf(tw, "%v\t", vm.Summary.Config.CpuReservation)
		fmt.Fprintf(tw, "%v\t", vm.Summary.Config.MemorySizeMB)
		fmt.Fprintf(tw, "%v\t", vm.Summary.Config.MemoryReservation)
		fmt.Fprintf(tw, "%s\t", vm.Summary.Runtime.PowerState)
		fmt.Fprintf(tw, "%s\t", vm.Summary.Guest.HwVersion)
		fmt.Fprintf(tw, "%s\t", vm.Summary.Guest.IpAddress)
		fmt.Fprintf(tw, "%s\t\n", vm.Summary.Config.VmPathName)
	}

	fmt.Fprintf(tw, "\n")

	_ = tw.Flush()
}
