////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//
// Description: Go code which will return the list of Kubernetes nodes in the current context, which is then used to
//              find the ESXi host on which the VM/node is running
//
//		There are also 2 pieces of simulation, one which calculated when the next maintenance schedule is 
//		due to take place on each host, and another which reports whether or not a host has a GPU
//
//		The first part is not available in vSphere (a maintenance mode scheudle which can be queried).
//
//		The second part  can be implemented but we would need to know the PCI identifiers to correctly
//		identify GPUs. Perhaps if the customer wanted a partiocular GPU for their workload, we could implement.
//		Now we just randomly state if the host has a GPU or not.
//
// Author: 	Cormac Hogan
//
// Date: 	4 Feb 2021
//
// Version:	v0.1
//
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//
// Note on running "go build"
//
// -- https://github.com/kubernetes/client-go/blob/master/INSTALL.md#for-the-casual-user
//
// -- If you want to write a simple script, don't care about a reproducible client library install, then simply:
//
//
// export GO111MODULE=on
// go get k8s.io/client-go@master
// go mod init 
//
//
// -- Also see README regarding go client dependencies
//
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

package main

import (
        "context"
        "flag"
        "fmt"
        "path/filepath"
        "net/url"
	"os"
	"text/tabwriter"
	"reflect"
	"math/rand"
	"time"

	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/session/cache"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/client-go/kubernetes"
        "k8s.io/client-go/tools/clientcmd"
        "k8s.io/client-go/util/homedir"

)

type CandidateList struct {
	hostName string
	availAccTime int
	hasGPU bool
	nodeMemoryUsage int32
	nodeCpuUsage int32
	nodeName string
}



func main() {

	var candidate []CandidateList
	var bestCandidates []CandidateList
	var winnerCandidate CandidateList

// Randomly selected time for the long running job, in hours

	desiredAcceleratorTime := 300

	suitableCandidates := 0

//
// Ref: https://golang.org/pkg/text/tabwriter/#NewWriter 
//
	tw := tabwriter.NewWriter(os.Stdout, 4, 0, 4, ' ', 0)

	
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
    		fmt.Fprintf(tw, "Error occurred: ", err)
	}

	fmt.Fprintf(tw, "There are %d nodes in the cluster\n", len(nodes.Items))

//
// Simulation Code for generating next maintenance slot, in hours and if GPU exists
//
	rand.Seed(time.Now().UnixNano())
    	mmMin := 200
    	mmMax := 400

//
// We need to get 3 environment variables to connect to vSphere
//
// GOVMOMI_URL
// GOVMOMI_USERNAME
// GOVMOMI_PASSWORD


	vc := os.Getenv ("GOVMOMI_URL")
	user := os.Getenv ("GOVMOMI_USERNAME")
	pwd := os.Getenv ("GOVMOMI_PASSWORD")


//	fmt.Fprintf (tw, "DEBUG: vc is %s\n", vc)	
//	fmt.Fprintf (tw, "DEBUG: user is %s\n", user)	
//	fmt.Fprintf (tw, "DEBUG: password is %s\n", pwd)	

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
		fmt.Fprintf(tw, "Could not parse vCenter URL (Are environment variables set?)")
	}

	if err != nil {
		fmt.Fprintf(tw, "vCenter URL parsing not successful, error %v", err)
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
	c := new (vim25.Client)

	err = s.Login(ctx, c, nil)

	if err != nil {
		fmt.Fprintf(tw, "Log in not successful- could not get vCenter client: %v\n", err)
               	return
        } else {
        	fmt.Fprintf(tw, "Log in successful\n")

//
// Create a view manager
//

		m := view.NewManager(c)

//
// Create a container view of VM objects
//

		v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
		if err != nil {
			fmt.Fprintf(tw, "Unable to create Virtual Machine Container View: error %s", err)
			return
		}

		defer v.Destroy(ctx)

//
// Retrieve summary property for all virtual machines - descriptions of objects are available at the following links
//
// Ref: https://vdc-download.vmware.com/vmwb-repository/dcr-public/b50dcbbf-051d-4204-a3e7-e1b618c1e384/538cf2ec-b34f-4bae-a332-3820ef9e7773/vim.vm.Summary.html
//

		var vms []mo.VirtualMachine
		err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)

		if err != nil {
			fmt.Fprintf(tw, "Unable to retrieve VM information: error %s", err)
			return
		}

//
// Create a container view of HostSystem objects
//

                h, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)

                if err != nil {
                        fmt.Fprintf(tw, "Unable to create Host Container View: error %s", err)
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
			fmt.Fprintf(tw, "Unable to retrieve Host information: error %s", err)
			return
		}
//
// Print summary per vm
//

		for i:= 0; i < len(nodes.Items); i++ {

			for _, vm := range vms {

				if vm.Summary.Config.Name == nodes.Items[i].ObjectMeta.Name {

//
// Find ESXi Hypervisor/Host where VM/Node runs, and display relevant info  (currently just the name of the host)
//


			        	for _, hs := range hss {
						if reflect.DeepEqual(hs.Summary.Host, vm.Summary.Runtime.Host) {
							candidate = append(candidate, CandidateList{
								hs.Summary.Config.Name,
//
// Simulation Code for generating next maintenance slot, in hours
//
								rand.Intn(mmMax - mmMin + 1) + mmMin,

//
// Simulation Code for randomly selecting if host has GPU or not
//
								(bool)(rand.Float32() < 0.5),
//
// Get some CPU and Memory usage stats from the node - we will use this to decide the best node in the case of multiple node candidate being available
//
								vm.Summary.QuickStats.GuestMemoryUsage,
								vm.Summary.QuickStats.OverallCpuDemand,

//
// VM Name - usually long in TKG clusters
//
								vm.Summary.Config.Name})
						}
					}
				}

			}
		}
//
// More simulator code:
//
// First step is to just return suitable candidates for the long running job
// Once the list of candidates is found, search through them for the winning candidate
// We decided to use the node/virtual machine that had the least amount of CPU used
//
		fmt.Fprintf(tw, "\n--\n")
		
		for _, entry := range candidate {

			if ((entry.availAccTime >= desiredAcceleratorTime) && (entry.hasGPU == true)) {

				fmt.Fprintf(tw, "\tSuitable candidate is node %s on ESXi host %s\n", entry.nodeName, entry.hostName)
				
				suitableCandidates++
				bestCandidates = append(bestCandidates, entry)

			} else {
				fmt.Fprintf(tw, "\tNode %s on ESXi host %s is not a suitable candidate for the long running job\n", entry.nodeName, entry.hostName)

				if (entry.hasGPU == false) {
					fmt.Fprintf(tw, "\t\tIt does not have a GPU: status is %v\n", entry.hasGPU)
				}

				if (entry.availAccTime < desiredAcceleratorTime) {
					fmt.Fprintf(tw, "\t\tDesired access time %v is greater than Available Accelerator Time %v\n", desiredAcceleratorTime, entry.availAccTime)
				}
				fmt.Fprintf(tw, "---\n")
			}
		}

//
// -- initialize winnerCandidate
//
		winnerCandidate.hostName = ""
		winnerCandidate.availAccTime = 0
		winnerCandidate.hasGPU = false
		winnerCandidate.nodeMemoryUsage = 0
		winnerCandidate.nodeCpuUsage = 999999 	// values returned from node statistics should be less than this value
		winnerCandidate.nodeName = ""


		if (suitableCandidates == 0) {
			fmt.Fprintf(tw, "Found *** NO*** suitable candidates for the long running job\n")
		} else if (suitableCandidates == 1) {
			fmt.Fprintf(tw, "\n\nFound a total of *** 1 *** suitable candidates for the long running job\n")
			fmt.Fprintf(tw, "\n--\n")
			fmt.Fprintf(tw, "Best Candidates:\n")
		
			for _, newentry := range bestCandidates {
				fmt.Fprintf(tw, "\t\t%s does have a GPU: status is %v\n", newentry.hostName, newentry.hasGPU)
				fmt.Fprintf(tw, "\t\tDesired access time %v is less than Available Accelerator Time %v\n", desiredAcceleratorTime, newentry.availAccTime)
				fmt.Fprintf(tw, "\t\tNode %s CPU Usage is %v MHz\n", newentry.nodeName, newentry.nodeCpuUsage)
				fmt.Fprintf(tw, "\t\tNode %s Memory Usage is %v MB\n", newentry.nodeName, newentry.nodeMemoryUsage)
				fmt.Fprintf(tw, "---\n")

				winnerCandidate.hostName = newentry.hostName
				winnerCandidate.nodeMemoryUsage = newentry.nodeMemoryUsage
				winnerCandidate.nodeCpuUsage = newentry.nodeCpuUsage
				winnerCandidate.nodeName = newentry.hostName
			}
		} else {
			fmt.Fprintf(tw, "\n\nFound a total of *** %v *** suitable candidates for the long running job\n", suitableCandidates)
			fmt.Fprintf(tw, "\n--\n")
			fmt.Fprintf(tw, "Winning Candidates:\n")

			for _, newentry := range bestCandidates {
				fmt.Fprintf(tw, "\t\t%s does have a GPU: status is %v\n", newentry.hostName, newentry.hasGPU)
				fmt.Fprintf(tw, "\t\tDesired access time %v is less than Available Accelerator Time %v\n", desiredAcceleratorTime, newentry.availAccTime)
				fmt.Fprintf(tw, "\t\tNode %s CPU Usage is %v MHz\n", newentry.nodeName, newentry.nodeCpuUsage)
				fmt.Fprintf(tw, "\t\tNode %s Memory Usage is %v MB\n", newentry.nodeName, newentry.nodeMemoryUsage)
				fmt.Fprintf(tw, "---\n")
//
// Pick best candidate based on lowest CPU Usags, it will be large on the first iteration
//
				if (newentry.nodeCpuUsage < winnerCandidate.nodeCpuUsage) {
					winnerCandidate.hostName = newentry.hostName
					winnerCandidate.nodeMemoryUsage = newentry.nodeMemoryUsage
					winnerCandidate.nodeCpuUsage = newentry.nodeCpuUsage
					winnerCandidate.nodeName = newentry.hostName
				}

			}
		}
		fmt.Fprintf(tw, "\n--\n")
		fmt.Fprintf(tw, "Winner:\n")
		fmt.Fprintf(tw, "\t\tWinning node is %s \n", winnerCandidate.nodeName)
		fmt.Fprintf(tw, "\t\tWinning host is is %v\n", winnerCandidate.hostName)
		fmt.Fprintf(tw, "\t\tNode %s CPU Usage is %v MHz\n", winnerCandidate.nodeName, winnerCandidate.nodeCpuUsage)
		fmt.Fprintf(tw, "\t\tNode %s Memory Usage is %v MB\n", winnerCandidate.nodeName, winnerCandidate.nodeMemoryUsage)
		fmt.Fprintf(tw, "---\n")
		
		_ = tw.Flush()
	}
}
