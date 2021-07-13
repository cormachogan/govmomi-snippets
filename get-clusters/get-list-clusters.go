//
// Sample GOVMOMI snippet to list all clusters in a vSphere environment
//

package main

import (
        "context"
        "fmt"
	      "flag"
        "net/url"
	      "os"
        "github.com/vmware/govmomi"
        "github.com/vmware/govmomi/view"
	      "github.com/vmware/govmomi/find"
	      "github.com/vmware/govmomi/vim25"
	      "github.com/vmware/govmomi/vim25/soap"
	      "github.com/vmware/govmomi/vim25/mo"
        "github.com/vmware/govmomi/session/cache"
)

func main() {


// We need to get 3 environment variables:
//
// GOVMOMI_URL=vcsa-06.rainpole.com/sdk
// GOVMOMI_USERNAME=administrator@vsphere.local
// GOVMOMI_PASSWORD=VMware123!
// GOVMOMI_INSECURE=true

	var insecure bool

	flag.BoolVar(&insecure, "insecure", true, "ignore any vCenter TLS cert validation error")

	vc := os.Getenv ("GOVMOMI_URL")
	user := os.Getenv ("GOVMOMI_USERNAME")
	pwd := os.Getenv ("GOVMOMI_PASSWORD")


	fmt.Printf ("DEBUG: vc is %s\n", vc)
	fmt.Printf ("DEBUG: user is %s\n", user)
	fmt.Printf ("DEBUG: password is %s\n", pwd)

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
	}
       	c.Logout(ctx)

//
// Display list of all clusters, but switch to vim25.client
//

        s := &cache.Session{
        	URL:      u,
               	Insecure: true,
        }

        vimc := new(vim25.Client)

        err = s.Login(ctx, vimc, nil)

        if err != nil {
          fmt.Println("")
          fmt.Println("Log in not successful (vim25) - could not get vCenter client: ", err)
          fmt.Println("")
        } else {
          fmt.Println("")
          fmt.Println("Log in successful (vim25)")
          fmt.Println("")
        }

	      m := view.NewManager(vimc)
	      v, err := m.CreateContainerView(ctx, vimc.ServiceContent.RootFolder, []string{"ClusterComputeResource"}, true)
	      var clusters []mo.ClusterComputeResource
	      err = v.Retrieve(ctx, []string{"ClusterComputeResource"}, []string{"name"}, &clusters)

        if err != nil {
        	fmt.Printf("Could not get list of clusters : error %s\n", err)
        } else {
		        for _, cluster := range clusters {
			         fmt.Println("Found a cluster:", cluster.Name)
		        }
        }

	defer v.Destroy(ctx)
}
