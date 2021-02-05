/*
This example program shows how the `view` and `property` packages can
be used to navigate a vSphere inventory structure using govmomi.
*/

package main

import (
	"context"
	"fmt"
//	"reflect"
	"os"
	"text/tabwriter"
	"net/url"
	"sort"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/units"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
        "github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/find"
        "github.com/vmware/govmomi/vim25/soap"
        "github.com/vmware/govmomi/session/cache"

)

//
//-- sort hosts by model
//

type hostByModel []mo.HostSystem
func (n hostByModel) Len() int           { return len(n) }
func (n hostByModel) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n hostByModel) Less(i, j int) bool { return n[i].Hardware.SystemInfo.Model < n[j].Hardware.SystemInfo.Model }



func main() {

// We need to get 3 environment variables:
//
// GOVMOMI_URL
// GOVMOMI_USERNAME
// GOVMOMI_PASSWORD

        vc := os.Getenv ("GOVMOMI_URL")
        user := os.Getenv ("GOVMOMI_USERNAME")
        pwd := os.Getenv ("GOVMOMI_PASSWORD")


//      fmt.Printf ("DEBUG: vc is %s\n", vc)
//      fmt.Printf ("DEBUG: user is %s\n", user)
//      fmt.Printf ("DEBUG: password is %s\n", pwd)

//
// Imagine that there were multiple operations taking place such as processing some data, logging into vCenter, etc.
// If one of the operations failed, the context would be used to share the fact that all of the other operations sharing that context needs cancelling.
//

        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()

//
// Create a vSphere/vCenter client
//
//    The govmomi client requires a URL object, u, not just a string representation of the vCenter URL.

        u, err := soap.ParseURL(vc)

        if u == nil {
                fmt.Println("could not parse URL (environment variables set?)")
        }

        if err != nil {
                fmt.Println("URL parsing not successful, error %v", err)
        }

        u.User = url.UserPassword(user, pwd)
//
// Share govc's session cache
//
        s := &cache.Session{
                URL:      u,
                Insecure: true,
        }

//
//    c - Return the client object c
//    err - Return the error object err
//    ctx - Pass in the shared context
//
        c := new (vim25.Client)

        err = s.Login(ctx, c, nil)

        if err != nil {
                fmt.Println("Log in not successful- could not get vCenter client: %v", err)
        } else {
                fmt.Println("Log in successful")

//
// Create a view of HostSystem objects
//

		m := view.NewManager(c)

		v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
		if err != nil {
       	         fmt.Println("error 1")
		}
	
		defer v.Destroy(ctx)

//
// Retrieve summary property for all hosts
//
// Ref: https://vdc-download.vmware.com/vmwb-repository/dcr-public/b50dcbbf-051d-4204-a3e7-e1b618c1e384/538cf2ec-b34f-4bae-a332-3820ef9e7773/vim.host.PciDevice.html
//

		var hss []mo.HostSystem
		err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
		if err != nil {
       	        	fmt.Println("error 2")
		}
	
//
// Print summary per host (see also: govc/host/info.go)
//

		tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "Name:\tUsed CPU:\tTotal CPU:\tFree CPU:\tUsed Memory:\tTotal Memory:\tFree Memory:\t\n")

		for _, hs := range hss {
			totalCPU := int64(hs.Summary.Hardware.CpuMhz) * int64(hs.Summary.Hardware.NumCpuCores)
			freeCPU := int64(totalCPU) - int64(hs.Summary.QuickStats.OverallCpuUsage)
			freeMemory := int64(hs.Summary.Hardware.MemorySize) - (int64(hs.Summary.QuickStats.OverallMemoryUsage) * 1024 * 1024)
			fmt.Fprintf(tw, "%s\t", hs.Summary.Config.Name)
			fmt.Fprintf(tw, "%d\t", hs.Summary.QuickStats.OverallCpuUsage)
			fmt.Fprintf(tw, "%d\t", totalCPU)
			fmt.Fprintf(tw, "%d\t", freeCPU)
			fmt.Fprintf(tw, "%s\t", (units.ByteSize(hs.Summary.QuickStats.OverallMemoryUsage))*1024*1024)
			fmt.Fprintf(tw, "%s\t", units.ByteSize(hs.Summary.Hardware.MemorySize))
			fmt.Fprintf(tw, "%d\t", freeMemory)
			fmt.Fprintf(tw, "\n")
		}
		_ = tw.Flush()


//
// -- To use Property collector, I need to govmomi client. It won't work with the vim25 client
//

		ctx2, cancel := context.WithCancel(context.Background())
    		defer cancel()

//
// Connect and log in to vCenter
//
   		c2, err := govmomi.NewClient(ctx2, u, true)
    		if err != nil {
               		fmt.Println("error 3")
    		}

		f := find.NewFinder(c2.Client, true)

//
// Find one and only datacenter
//

		dc, err := f.DefaultDatacenter(ctx2)
    
		if err != nil {
               	fmt.Println("error 4")
		}

//
// Make future calls local to this datacenter
//

		f.SetDatacenter(dc)

    		pc := property.DefaultCollector(c2.Client)

//
// Convert hosts into list of references
//

		var refs []types.ManagedObjectReference

		for _, hs := range hss {
    			refs = append(refs, hs.Reference())
		}

//
// Retrieve all property for all hosts -- note there that the focus is on host name, not hostsummary
//
// Ref: https://vdc-download.vmware.com/vmwb-repository/dcr-public/b50dcbbf-051d-4204-a3e7-e1b618c1e384/538cf2ec-b34f-4bae-a332-3820ef9e7773/vim.HostSystem.html
//
  		var allhosts []mo.HostSystem

    		err = pc.Retrieve(ctx, refs, []string{"hardware"}, &allhosts)
    		if err != nil {
               		fmt.Println("error 5")
    		}

    		fmt.Fprintf(tw, "\nHosts found: %v\n\n", len(allhosts))
//
// Let's sort them
//
		sort.Sort(hostByModel(allhosts))

//
// Note usefulness of reflect to tell you what *type* has been returned
//
//       	dumpinfo := reflect.ValueOf(fullhost.Hardware.PciDevice)    //.Elem()
//        	fmt.Fprintf(tw, "%v\n", dumpinfo)
//        	fmt.Fprintf(tw, "%v\n", dumpinfo.Type())


//
// For each host, get a list of its PCI Devices --- not working!!!!
//
		var pciDevices []types.HostPciDevice

		fmt.Fprintf(tw, "UUid:\tModel:\tVendor:\tNumber of PCI Devices\t\n")
		fmt.Fprintf(tw, "-----\t------\t-------\t---------------------\t\n")

		for _, fullhost := range allhosts {

//    	 	dumpinfo := reflect.ValueOf(fullhost)    //.Elem()
//        	fmt.Fprintf(tw, "%v\n", dumpinfo)
//        	fmt.Fprintf(tw, "%v\n", dumpinfo.Type())

//
// Note that the ...  at the end of the append call unpacks the slice. You can now work on the individual entries again 
//
		pciDevices = append(pciDevices, fullhost.Hardware.PciDevice...)


		if len(pciDevices) == 0 {
            		fmt.Fprintf(tw, "%v\t%v\t%v\t0\n", fullhost.Hardware.SystemInfo.Uuid,  fullhost.Hardware.SystemInfo.Vendor,  fullhost.Hardware.SystemInfo. Model)
        	} else {
	        	fmt.Fprintf(tw, "%v\t%v\t%v\t%v\n", fullhost.Hardware.SystemInfo.Uuid,  fullhost.Hardware.SystemInfo.Vendor,  fullhost.Hardware.SystemInfo. Model, len(pciDevices) )

//
// Ref: https://vdc-download.vmware.com/vmwb-repository/dcr-public/b50dcbbf-051d-4204-a3e7-e1b618c1e384/538cf2ec-b34f-4bae-a332-3820ef9e7773/vim.host.PciDevice.html
//

//
// Use this to get individual PCI device information, such as the VendorName
//

    			//for _, obj := range pciDevices {
       	         	//	fmt.Fprintf(tw, "\tFound PCI devices %s.\n", obj.VendorName)
			//	}
   			}
		}
		_ = tw.Flush()
	}
}
