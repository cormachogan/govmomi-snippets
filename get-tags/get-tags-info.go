
package main

import (
	"context"
	"fmt"
	"os"
	"net/url"

	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/property"
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
//    The vim25 client requires a URL object, u, not just a string representation of the vCenter URL.

        u, err := soap.ParseURL(vc)

        if u == nil {
                fmt.Println("could not parse URL (environment variables set?)")
        }

        if err != nil {
                fmt.Println("URL parsing not successful, error %v", err)
        }

        u.User = url.UserPassword(user, pwd)
//
// Share session cache
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
                fmt.Println("Log in (vim25) not successful- could not get vCenter client: %v", err)
        } else {
                fmt.Println("Log in (vim25) successful")
	}

//
// -- Just get tags of VMs -- e.g. https://github.com/vmware/govmomi/blob/master/vapi/tags/example_test.go
//

	v, err := view.NewManager(c).CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
	}

// List all VMs in the inventory
	vms, err := v.Find(ctx, nil, property.Filter{})
	if err != nil {
                fmt.Println("Error in find", err)
	}

// Convert list type
	refs := make([]mo.Reference, len(vms))
	for i := range vms {
		refs[i] = vms[i]
	}

//
// -- This section of code will print tags associated with VMs only, which will have a list of refs for
//
        fmt.Println("\n-- First mechanism -- VM Objects\n")


//
// -- using the rest client for tags (rest client uses vim25 client)
//
//    rc - Return the client object rc
//    err - Return the error object err
//    ctx - Pass in the shared context
//

	rc := rest.NewClient(c)

	err = s.Login(ctx, rc, nil)
        if err != nil {
                fmt.Println("Log in (rest) not successful- could not get vCenter client: %v", err)
        } else {
                fmt.Println("Log in (rest) successful")
	}
	m := tags.NewManager(rc)
	attached, err := m.GetAttachedTagsOnObjects(ctx, refs)

       	for _, id := range attached {
		fmt.Println("\tFound Tags :", id.Tags)
	}

        fmt.Println("\n")
//
// -- https://pkg.go.dev/github.com/vmware/govmomi/vapi/tags#Tag
// -- https://pkg.go.dev/github.com/vmware/govmomi/vapi/tags#AttachedTags.Tags
//

        for _, vm := range attached {
                fmt.Println("Found Tags on VM:", vm.ObjectID.Reference().Value)

//
// Convert from a reference back to a VM - somehow ...
//

//
// Now print the tags for the VM
//
                for _, found_tag := range vm.Tags {
                        fmt.Println ("Found Tag Name: ", found_tag.Name)
                }
                fmt.Println("\n")
        }

//
// Try another way -- for all Objects, not just VMs
//
        fmt.Println("\n-- Alternate mechanism -- All Objects\n")

//
// Get a list of all tags
//
        tagList, err := m.ListTags(ctx)
        if err != nil {
                fmt.Println("Could not get list of tags, error %v\n", err)
        }

        for _, tag := range tagList {
                fmt.Println("Found a Tag in list of Tags:", tag)
//
// Get Tag details
//
                taginfo, err := m.GetTag(ctx, tag)
		if err != nil {
			fmt.Println("Could not get tag info, error %v\n", err)
		}

		fmt.Println("Found Tag Name:", taginfo.Name)
//
// Get Inventory Items that are using the Tags
//
		attached2, err := m.GetAttachedObjectsOnTags(ctx, []string{tag})
		if err != nil {
			fmt.Println("Could not get list of objects with tags, error %v\n", err)
		}

		for _, item := range attached2 {
			fmt.Println("Found Inventory Item(s) with Tag:", item)
			fmt.Println("\n")
		}
	}
	fmt.Println("\n")

}
