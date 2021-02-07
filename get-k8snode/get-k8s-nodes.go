////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//
// Description: Go code which will return the list of Kubernetes nodes in the current context, which is then used to
//              find the ESXi host on which the VM/node is running
//
// Author: Cormac Hogan
//
// Date: 1 Feb 2021
//
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"text/tabwriter"
	"time"

	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {

	//
	// Find the KUBECONFIG, which is most likely $HOME/.kube/config
	//

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// BuildConfigFromFlags is a helper function that builds configs from a master
	// url or a kubeconfig filepath. These are passed in as command line flags for cluster
	// components. Warnings should reflect this usage. If neither masterUrl or kubeconfigPath
	// are passed in we fallback to inClusterConfig. If inClusterConfig fails, we fallback
	// to the default config.

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	nodeList := clientSet.CoreV1().Nodes()
	nodes, err := nodeList.List(context.TODO(), v1.ListOptions{})

	if err != nil {
		fmt.Println("Error occurred: ", err)
	}

	fmt.Printf("There are %d nodes in the cluster\n", len(nodes.Items))

	//
	// Simulation Code for generating next maintenance slot, in hours
	//
	rand.Seed(time.Now().UnixNano())
	min := 200
	max := 400

	// We need to get 3 environment variables:
	//
	// GOVMOMI_URL
	// GOVMOMI_USERNAME
	// GOVMOMI_PASSWORD

	vc := os.Getenv("GOVMOMI_URL")
	user := os.Getenv("GOVMOMI_USERNAME")
	pwd := os.Getenv("GOVMOMI_PASSWORD")

	//	fmt.Printf ("DEBUG: vc is %s\n", vc)
	//	fmt.Printf ("DEBUG: user is %s\n", user)
	//	fmt.Printf ("DEBUG: password is %s\n", pwd)

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
		return
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
	c := new(vim25.Client)

	err = s.Login(ctx, c, nil)

	if err != nil {
		fmt.Println("Log in not successful- could not get vCenter client: %v", err)
		return
	} else {
		fmt.Println("Log in successful")

		//
		// Create a view manager - a mechanism that supports selection of objects on the server and subsequently, access to those objects.
		//
		// Ref: https://vdc-download.vmware.com/vmwb-repository/dcr-public/b50dcbbf-051d-4204-a3e7-e1b618c1e384/538cf2ec-b34f-4bae-a332-3820ef9e7773/vim.view.ViewManager.html
		//

		m := view.NewManager(c)

		//
		// Create a container view (a means of monitoring the contents of a single container) of VM objects
		//
		// Ref: https://vdc-download.vmware.com/vmwb-repository/dcr-public/b50dcbbf-051d-4204-a3e7-e1b618c1e384/538cf2ec-b34f-4bae-a332-3820ef9e7773/vim.view.ContainerView.html
		//
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
		// Ref: https://vdc-download.vmware.com/vmwb-repository/dcr-public/b50dcbbf-051d-4204-a3e7-e1b618c1e384/538cf2ec-b34f-4bae-a332-3820ef9e7773/vim.vm.Summary.html
		// Ref: https://vdc-download.vmware.com/vmwb-repository/dcr-public/b50dcbbf-051d-4204-a3e7-e1b618c1e384/538cf2ec-b34f-4bae-a332-3820ef9e7773/vim.vm.Summary.GuestSummary.html
		// Ref: https://vdc-download.vmware.com/vmwb-repository/dcr-public/b50dcbbf-051d-4204-a3e7-e1b618c1e384/538cf2ec-b34f-4bae-a332-3820ef9e7773/vim.vm.Summary.ConfigSummary.html
		// Ref: https://vdc-download.vmware.com/vmwb-repository/dcr-public/b50dcbbf-051d-4204-a3e7-e1b618c1e384/538cf2ec-b34f-4bae-a332-3820ef9e7773/vim.vm.RuntimeInfo.html
		//

		var vms []mo.VirtualMachine
		err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)

		if err != nil {
			fmt.Printf("Unable to retrieve VM information: error %s", err)
			return
		}

		//
		// Create a container view (a means of monitoring the contents of a single container) of Hostsystem objects
		//
		// Ref: https://vdc-download.vmware.com/vmwb-repository/dcr-public/b50dcbbf-051d-4204-a3e7-e1b618c1e384/538cf2ec-b34f-4bae-a332-3820ef9e7773/vim.view.ContainerView.html
		//

		h, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)

		if err != nil {
			fmt.Printf("Unable to create Host Container View: error %s", err)
			return
		}

		defer h.Destroy(ctx)

		//
		// Retrieve summary property for all ESXi hosts
		//
		// Ref: https://vdc-download.vmware.com/vmwb-repository/dcr-public/b50dcbbf-051d-4204-a3e7-e1b618c1e384/538cf2ec-b34f-4bae-a332-3820ef9e7773/vim.HostSystem.html
		//

		var hss []mo.HostSystem
		err = h.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)

		if err != nil {
			fmt.Printf("Unable to retrieve Host information: error %s", err)
			return
		}

		//
		// Print summary per vm
		//
		// -- https://golang.org/pkg/text/tabwriter/#NewWriter
		//

		tw := tabwriter.NewWriter(os.Stdout, 4, 0, 4, ' ', 0)

		for i := 0; i < len(nodes.Items); i++ {

			fmt.Fprintf(tw, "Guest\tHW Version\tIP Address\tESXi Hypervisor Hostname\tHours to Maintenance\tVirtual Machine/Nodename\n")
			fmt.Fprintf(tw, "-----\t-- -------\t-- -------\t---- ---------- --------\t----- -- -----------\t------------------------\n")

			for _, vm := range vms {

				if vm.Summary.Config.Name == nodes.Items[i].ObjectMeta.Name {
					fmt.Fprintf(tw, "%s\t", vm.Summary.Guest.GuestId)
					fmt.Fprintf(tw, "%s\t", vm.Summary.Guest.HwVersion)
					fmt.Fprintf(tw, "%s\t", vm.Summary.Guest.IpAddress)

					//
					// Find Host where VM/Node runs, and display relevant info  (currently just the name of the host)
					//

					for _, hs := range hss {
						if reflect.DeepEqual(hs.Summary.Host, vm.Summary.Runtime.Host) {
							fmt.Fprintf(tw, "%s\t", hs.Summary.Config.Name)
						}
					}

					//
					// Simulation Code for generating next maintenance slot, in hours
					//

					fmt.Fprintf(tw, "%v\t", rand.Intn(max-min+1)+min)

					fmt.Fprintf(tw, "%s\n", vm.Summary.Config.Name)
				}
			}
			fmt.Fprintf(tw, "\n")
			_ = tw.Flush()
		}
	}
}
