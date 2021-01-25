package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"text/tabwriter"
	"sort"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
        "github.com/vmware/govmomi/vim25"
        "github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vslm"
)


//-- sort datastores

type dsByName []mo.Datastore
func (n dsByName) Len() int           { return len(n) }
func (n dsByName) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n dsByName) Less(i, j int) bool { return n[i].Name < n[j].Name }



func main() {
	//
	// 3 environment variables are required in order to connect to the vSphere infra
	//
	// Set these in your shell to reflect your vSphere infra:
	//
	// GOVMOMI_URL
	// GOVMOMI_USERNAME
	// GOVMOMI_PASSWORD
	//

	vc := os.Getenv("GOVMOMI_URL")

	if len(vc) > 0 {
		fmt.Printf("DEBUG: vc is %s\n", vc)
	} else {
		fmt.Printf("Unable to find env var GOVMOMI_URL, has it been set?\n", vc)
		return
	}

	user := os.Getenv("GOVMOMI_USERNAME")

	if len(user) > 0 {
		fmt.Printf("DEBUG: user is %s\n", user)
	} else {
		fmt.Printf("Unable to find env var GOVMOMI_USERNAME, has it been set?\n", vc)
		return
	}
	pwd := os.Getenv("GOVMOMI_PASSWORD")

	if len(pwd) > 0 {
		fmt.Printf("DEBUG: password is %s\n", pwd)
	} else {
		fmt.Printf("Unable to find env GOVMOMI_PASSWORD, has it been set?\n", vc)
		return
	}

	//
	// This section allows for insecure vSphere logins
	//

	var insecure bool
	flag.BoolVar(&insecure, "insecure", true, "ignore any vCenter TLS cert validation error")

	//
	// Imagine that there were multiple operations taking place such as processing some data, logging into vCenter, etc.
	// If one of the operations failed, the context would be used to share the fact that all of the other operations sharing that context needs cancelling.
	//

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//
	// Create a vSphere/vCenter client
	//
	//    The govmomi client requires a URL object not just a string representation of the vCenter URL
	//

	u, err := soap.ParseURL(vc)

	if u == nil {
		fmt.Println("could not parse URL (environment variables set?)")
	}

	if err != nil {
		fmt.Println("URL parsing not successful, error ", err)
		return
	}

	u.User = url.UserPassword(user, pwd)

	//-------------------------------------------------------------------
	//
	//     c, err - Return the client object c and an error object err
	//
	//     govmomi.NewClient - Call the function from the govmomi package
	//
	//     ctx - Pass in the shared context
	//
	//-------------------------------------------------------------------

	//
	//  A lot of GO functions return more than one variable/object
	//  The majority also return an object of type error.
	//
	//  If the function call is successful it returns nil in the place of an error object.
	//
	//  If something goes wrong the function should create a new error object with the appropriate messaging.
	//

	c, err := govmomi.NewClient(ctx, u, insecure)

	if err != nil {
		fmt.Println("")
		fmt.Println("Log in not successful (govmomi) - could not get vCenter client: ", err)
		fmt.Println("")
		return
	} else {
		fmt.Println("")
		fmt.Println("Log in successful (govmomi)")
		fmt.Println("")
	}

	//
	// -- "find" implements inventory listing and searching.
	// -- https://gowalker.org/github.com/vmware/govmomi/find
	//

	finder := find.NewFinder(c.Client, true)

	//
	// -- find and set the default datacenter
	//

	dc, err := finder.DefaultDatacenter(ctx)

	if err != nil {
		fmt.Println("")
		fmt.Println("Could not get default datacenter, error: ", err)
		fmt.Println("")
		c.Logout(ctx)
	} else {
		fmt.Println("")
		fmt.Println("Found default datacenter: ", dc)
		fmt.Println("")
		finder.SetDatacenter(dc)
	}

//
// Find the datastores available on this vSphere Infrastructure
//

//
// Retrieve summary property for all datastores
//
// -- http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.Datastore.html
//

	dss, err := finder.DatastoreList(ctx, "*")
        if err != nil {
                return
        }

	if err != nil {
		fmt.Println("")
		fmt.Println("Could not get datastore list, error: ", err)
		fmt.Println("")
	} else {
		fmt.Println("")
		fmt.Println("Got datastore, number of datastore(s) is : ", len(dss))
		fmt.Println("")

		for i := 0; i < len(dss); i++ {
			fmt.Printf("Found host: %s\n", dss[i])
		}

		pc := property.DefaultCollector(c.Client)


//
// "finder" only lists - to get really detailed info,
// Convert datastores into list of references
//

		var refs []types.ManagedObjectReference
		for _, ds := range dss {
			refs = append(refs, ds.Reference())
		}

//
// Retrieve name property for all datastore
//

		var dst []mo.Datastore
		err = pc.Retrieve(ctx, refs, []string{"name"}, &dst)
		if err != nil {
			return
		}

		fmt.Printf("\n")

//
// Print name of each datastore
//
		tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
        	fmt.Println("Datastores found:", len(dst))
		fmt.Printf("\n")
        	sort.Sort(dsByName(dst))

		for _, newds := range dst {
			fmt.Fprintf(tw, "Found Datastore: %s\n", newds.Name)
		}

		fmt.Printf("\n")

//
// Login using vim25 for vslm -  this is so that 
// we can get the First Class Disk (FCD/IVD) listings
//

	        s := &cache.Session{
       	        	URL:      u,
       	        	Insecure: true,
       	 	}

       	 	c2 := new (vim25.Client)

       	 	err = s.Login(ctx, c2, nil)

       	 	if err != nil {
       	        	fmt.Println("Log in not successful (vim25) - could not get vCenter client: %v", err)
       	        	return
       	 	} else {
                	fmt.Println("Log in successful (vim25)")
			fmt.Printf("\n")
		}

//
// -- More information about vslm 
//
// -- https://pkg.golangclub.com/github.com/vmware/govmomi/vslm?tab=doc
// -- https://github.com/vmware/govmomi/blob/v0.20.0/vslm/object_manager.go#L190
//

		m := vslm.NewObjectManager(c2)

//
// -- Display the FCDs on each datastore (held in array dst)
//
// -- https://pkg.golangclub.com/github.com/vmware/govmomi/vim25/types?tab=doc#VStorageObject
//


		var objids []types.ID
		var idinfo *types.VStorageObject

		for _, newds := range dst {
			fmt.Fprintf(tw, "List of FCDs on datastore: %s\n", newds.Name)
			fmt.Fprintf(tw, "\n")
			objids, err = m.List(ctx, newds)
//
// - With the list of FCD Ids, we can get further information about the FCD retrievec in VStorageObject
//
			for _, id := range objids {
				fmt.Fprintf(tw, "\tFound FCD Id: %s\n", id.Id)
				idinfo, err = m.Retrieve(ctx, newds, id.Id)
//
// -- More info:
// -- https://pkg.golangclub.com/github.com/vmware/govmomi/vim25/types?tab=doc#BaseConfigInfo
//
				fmt.Fprintf(tw, "\tFCD Name              : %s\n", idinfo.Config.BaseConfigInfo.Name)
				fmt.Fprintf(tw, "\tFCD Creation Time     : %s\n", idinfo.Config.BaseConfigInfo.CreateTime)
				fmt.Fprintf(tw, "\tFCD Size (MB)         : %v\n", idinfo.Config.CapacityInMB)
				fmt.Fprintf(tw, "\tFCD Consumption Type  : %s\n", idinfo.Config.ConsumptionType)
//
// -- More info:
// -- https://pkg.golangclub.com/github.com/vmware/govmomi/vim25/types?tab=doc#BaseBaseConfigInfoBackingInfo
//
				ds := idinfo.Config.BaseConfigInfo.Backing.GetBaseConfigInfoBackingInfo()
				fmt.Fprintf(tw, "\tFCD Datastore Type    : %v\n", ds.Datastore.Type)
				fmt.Fprintf(tw, "\tFCD Datastore Info    : %v\n", ds.Datastore.Value)
//
// -- More info:
// -- https://pkg.go.dev/github.com/vmware/govmomi/vim25/types#BaseConfigInfoFileBackingInfo
//
				backing := idinfo.Config.BaseConfigInfo.Backing.(*types.BaseConfigInfoDiskFileBackingInfo)
				fmt.Fprintf(tw, "\tFCD FilePath          : %s\n", backing.FilePath) 
				fmt.Fprintf(tw, "\tFCD Backing Object Id : %s\n", backing.BackingObjectId) 
				fmt.Fprintf(tw, "\tFCD Delta Size (MB)   : %v\n", backing.DeltaSizeInMB) 
				fmt.Fprintf(tw, "\tFCD Provisioning Type : %v\n", backing.ProvisioningType) 

				fmt.Fprintf(tw, "\n")
			}
			fmt.Fprintf(tw, "\n")
		}
	}
}
