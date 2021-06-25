package main

import (
	"context"
	"fmt"
//	"reflect"
	"os"
	"text/tabwriter"
	"net/url"

	"github.com/vmware/govmomi/view"
        "github.com/vmware/govmomi/vim25"
        "github.com/vmware/govmomi/vim25/mo"
        "github.com/vmware/govmomi/vim25/soap"
        "github.com/vmware/govmomi/session/cache"

)


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

		m := view.NewManager(c)

//--- Get Host Info. Create a view of HostSystem objects from the RootFolder

		v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
		if err != nil {
       	        	fmt.Println("error 1")
		}

		defer v.Destroy(ctx)

		var hss []mo.HostSystem
//
// Retrieve the host list summary propery
//
		err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
		if err != nil {
       	        	fmt.Println("error 2")
		}

		tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "\n- List of ALL hosts:\n")

		for _, hs := range hss {
			fmt.Fprintf(tw, "- %s:\t%s\n", hs.Summary.Config.Name, hs.Reference())
		}

//--- Get Datastore Info. Create a view of Datastore objects from the RootFolder

		v2, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datastore"}, true)
		if err != nil {
       	        	fmt.Println("error 3")
		}

		defer v2.Destroy(ctx)

		var dss []mo.Datastore
//
// Retrieve the datastore list summary propery
//

		err = v2.Retrieve(ctx, []string{"Datastore"}, []string{"summary"}, &dss)
		if err != nil {
       	        	fmt.Println("error 4")
		}

		fmt.Fprintf(tw, "\n-- List of ALL datastores:\n")

		tw = tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
		for _, ds := range dss {
			fmt.Fprintf(tw, "-- %s:\t%s\n", ds.Summary.Name, ds.Reference())
		}

//--- Get Network Info. Create a view of Network objects from the RootFolder

		v3, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Network"}, true)
		if err != nil {
       	        	fmt.Println("error 5")
		}

		defer v3.Destroy(ctx)

		var nws []mo.Network
//
// Retrieve the network list -- there is no "summary" property for network which is why you can use either name or nil here
//

		//err = v3.Retrieve(ctx, []string{"Network"}, nil, &nws)
		err = v3.Retrieve(ctx, []string{"Network"}, []string{"name"}, &nws)
		if err != nil {
       	        	fmt.Println("error 6")
		}

		fmt.Fprintf(tw, "\n--- List of ALL Networks:\n")

		tw = tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
		for _, nw := range nws {
			fmt.Fprintf(tw, "--- %s:\t%s\n", nw.Name, nw.Reference())
		}
		_ = tw.Flush()
	}
}
