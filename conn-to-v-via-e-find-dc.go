
//
// functionality comes from the following packages
//
//    context - https://golang.org/pkg/context/
//    flag - https://golang.org/pkg/flag/
//    fmt - https://golang.org/pkg/fmt/
//    net/url - https://golang.org/pkg/net/url/
//    govmomi - https://github.com/vmware/govmomi
//    os - TBD

package main

import (
        "context"
        "fmt"
        "flag"
        "net/url"
        "os"
        "text/tabwriter"
        "github.com/vmware/govmomi"
        "github.com/vmware/govmomi/view"
        "github.com/vmware/govmomi/units"
        "github.com/vmware/govmomi/find"
        "github.com/vmware/govmomi/vim25"
        "github.com/vmware/govmomi/vim25/mo"
        "github.com/vmware/govmomi/vim25/soap"
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
//     govmomi.NewClient - Call the function from the govmomi package
//     ctx - Pass in the shared context
//
//-------------------------------------------------------------------

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
// -- find implements inventory listing and searching.
//

        finder := find.NewFinder(c.Client, true)

//
// -- find and set the datacenter
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
// -- display the hosts in the datacenter
//

        hosts, err := finder.HostSystemList(ctx, "*")

        if err != nil {
                        fmt.Println("")
                        fmt.Println("Could not get host list, error: ", err)
                        fmt.Println("")
                } else {
                        fmt.Println("")
                        fmt.Println("Got host list, number of host(s) is : ", len(hosts))
                        fmt.Println("")

                        for i := 0; i < len(hosts); i++ {
                        fmt.Printf("Found host: %s\n", hosts[i])
                        }
        }
//
// -- logout
//
        c.Logout(ctx)

//-------------------------------------------------------------------
//
//     vimc, err - Return the client object vimc and an error object err
//     vim25.NewClient - Call the function from the vim25 package
//     ctx - Pass in the shared context
//
//-------------------------------------------------------------------

//
// Ripped from https://github.com/vmware/govmomi/blob/master/examples/examples.go
//

//
// Share govc's session cache
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

//
// -- Create a view of HostSystem objects
//

        m := view.NewManager(vimc)

        v, err := m.CreateContainerView(ctx, vimc.ServiceContent.RootFolder, []string{"HostSystem"}, true)
        if err != nil {
                fmt.Println("")
                fmt.Println("Container View creation not successful: ", err)
                fmt.Println("")
        }

        defer v.Destroy(ctx)

//
// Retrieve summary property for all hosts
// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.HostSystem.html
//

        var hss []mo.HostSystem
        err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
        if err != nil {
                fmt.Println("")
                fmt.Println("Unable to retrieve HostSystem information: ", err)
                fmt.Println("")
        }


//
// -- Print summary per host (see also: govc/host/info.go)
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
// -- display the datastores in the datacenter
//

//-------------------------------------------------------------------
//
//     c, err - Return the client object c and an error object err
//     govmomi.NewClient - Call the function from the govmomi package
//     ctx - Pass in the shared context
//
//-------------------------------------------------------------------

        c2, err := govmomi.NewClient(ctx, u, insecure)

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
// -- find implements inventory listing and searching.
//

        finder2 := find.NewFinder(c2.Client, true)

//
// -- find and set the datacenter
//

        dc2, err := finder2.DefaultDatacenter(ctx)

        if err != nil {
                fmt.Println("")
                fmt.Println("Could not get default datacenter, error: ", err)
                fmt.Println("")
                c2.Logout(ctx)
        } else {
                fmt.Println("")
                fmt.Println("Found default datacenter: ", dc2)
                fmt.Println("")

                finder2.SetDatacenter(dc2)
        }

        dss, err := finder2.DatastoreList(ctx, "*")

        if err != nil {
                        fmt.Println("Could not get datastore list, error: ", err)
                } else {
                        fmt.Println("")
                        fmt.Println("Got datastore list, number of datastore(s) is : ", len(dss))
                        fmt.Println("")

                        for i := 0; i < len(dss); i++ {
                        fmt.Printf("Found datastore: %s\n", dss[i])
                        }
        }
        fmt.Println("")
//
// -- logout
//
        c2.Logout(ctx)
}

