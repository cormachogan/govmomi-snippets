////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//
// Description: 		Go code to connect to vSphere via environment
// 						variables and retrieve vSphere hosts, datastores
// 						and virtual machines
//
// Author:				Cormac J. Hogan (VMware)
//
// Date:				25 Jan 2021
//
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/soap"
)

func main() {

	// We need to get 3 environment variables:
	//
	// GOVMOMI_URL
	// GOVMOMI_USERNAME
	// GOVMOMI_PASSWORD

	var insecure bool

	flag.BoolVar(&insecure, "insecure", true, "ignore any vCenter TLS cert validation error")

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
	// Create a vSphere/vCenter client
	//
	//    The govmomi client requires a URL object not just a string representation of the vCenter URL.
	//    c, err - Return the client object c and an error object err
	//    govmomi.NewClient - Call the function from the govmomi package
	//    ctx - Pass in the shared context

	u, err := soap.ParseURL(vc)

	if u == nil {
		fmt.Println("could not parse URL (environment variables set?)")
	}

	if err != nil {
		fmt.Println("URL parsing not successful, error %v", err)
		return
	}

	u.User = url.UserPassword(user, pwd)

	c, err := govmomi.NewClient(ctx, u, insecure)

	if err != nil {
		fmt.Println("Log in not successful- could not get vCenter client: %v", err)
		return
	} else {
		fmt.Println("Log in successful")

		finder := find.NewFinder(c.Client)

		//
		// Find the Default Datacenter
		//

		dc, err := finder.DefaultDatacenter(ctx)

		if err != nil {
			fmt.Printf("Could not get default datacenter : error %s\n", err)
		} else {
			fmt.Printf("Default Datacenter %s found\n", dc)
			finder.SetDatacenter(dc)
		}

		//
		// Find the Default Cluster
		//

		cl, err := finder.DefaultClusterComputeResource(ctx)

		if err != nil {
			fmt.Printf("Could not get default cluster : error %s\n", err)
		} else {
			fmt.Printf("Default cluster %s found\n", cl)
		}

		c.Logout(ctx)
	}
}
