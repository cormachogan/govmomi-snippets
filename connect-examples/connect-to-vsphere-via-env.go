//
// Go code to connect to vSphere via environment
// variables and retrieve the defautl datacenter
//
// -- Cormac J. Hogan (VMware)
//
// -- 25 Jan 2021
//
//------------------------------------------------------------------------------------------------------------------------------------
//
// client information from Doug MacEachern:
//
// govmomi.Client extends vim25.Client
// govmomi.Client does nothing extra aside from automatic login
//
// In the early days (2015), govmomi.Client did much more, but we moved most of it to vim25.Client.
// govmomi.Client remained for compatibility and minor convenience.
//
// Using soap.Client and vim25.Client directly allows apps to use other authentication methods,
// session caching, session keepalive, retries, fine grained TLS configuration, etc.
//
// For the inventory, ContainerView is a vSphere primitive.
// Compared to Finder, ContainerView tends to use less round trip calls to vCenter.
// It may generate more response data however.
//
// Finder was written for govc, where we treat the vSphere inventory as a virtual filesystem.
// The inventory path as input to `govc` behaves similar to the `ls` command, with support for relative paths, wildcard matching, etc.
//
// Use govc commands as a reference, and "godoc" for examples that can be run against `vcsim`:
// See: https://godoc.org/github.com/vmware/govmomi/view#pkg-examples
//
//------------------------------------------------------------------------------------------------------------------------------------
//
// functionality comes from the following packages
//
//    context        - https://golang.org/pkg/context/
//    flag           - https://golang.org/pkg/flag/
//    fmt            - https://golang.org/pkg/fmt/
//    net/url        - https://golang.org/pkg/net/url/
//    os             - https://golang.org/pkg/os/
//    text/tabwriter - https://golang.org/pkg/text/tabwriter/
//
//    govmomi        - https://github.com/vmware/govmomi

package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/soap"
)

func main() {

	// We need to get 3 environment variables in order to connect to the vSphere infra
	//
	// Change these to reflect your vSphere infra:
	//
	// GOVMOMI_URL=vcsa-06.rainpole.com/sdk
	// GOVMOMI_USERNAME=administrator@vsphere.local
	// GOVMOMI_PASSWORD=VMware123!
	// GOVMOMI_INSECURE=true

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
		c.Logout(ctx)
	}
}
