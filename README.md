# govmomi-examples #

A few example govmomi scripts, mostly involving how to connect to vSphere and retrieve some infrastructure information

For first time use, run the following:

```shell
 % go build <Filename>.go 
 % go run <Filename>.go
```

For subsequent runs, once the imports are local, simply run:

```shell
 % go run <Filename>.go
```

Two of the scripts show the different ways to connect to vSphere
- via URL
- via environment variables

The other scripts show how to connect and retrieve various vSphere information, e.g.
- Datacenter
- Cluster
- Hosts
- Datastores
- Virtual Machines (VMs)
- First Class Disks (FCDs) - used to back Kubernetes Persistent Volumes

Here are some example outputs, assuming the environment variables have been set appropriately.

```shell
$ go run conn-to-v-via-e-find-ho-ds-vm.go
DEBUG: vc is vcsa-06.rainpole.com/sdk
DEBUG: user is administrator@vsphere.local
DEBUG: password is *********
Log in successful

*** Host Information ***
------------------------

Name:                           Used CPU:  Total CPU:  Free CPU:  Used Memory:  Total Memory:  Free Memory:
esxi-dell-h.rainpole.com        1615       43980       42365      48.9GB        127.9GB        84821704704
esxi-dell-f.rainpole.com        3832       43980       40148      58.8GB        127.9GB        74196484096
esxi-dell-e.rainpole.com        4521       43980       39459      53.5GB        127.9GB        79893397504
esxi-dell-g.rainpole.com        1496       43980       42484      43.3GB        127.9GB        90875133952
vcsa06-witness-01.rainpole.com  19         4400        4381       1.3GB         16.0GB         15764742144

*** Datastore Information ***
------------------------------

Name:              Type:  Capacity:  Free:
PureVMFSDatastore  VMFS   500.0TB    499.9TB
isilon-01          NFS    50.5TB     46.3TB
vsanDatastore      vsan   5.8TB      5.0TB

*** VM Information ***
-----------------------

Name:                                               Guest Full Name:
SupervisorControlPlaneVM (1):                       Other 3.x Linux (64-bit)
vSAN File Service Node (8):                         Other 3.x or later Linux (64-bit)
haproxy:                                            Other 3.x or later Linux (64-bit)
vCLS (8):                                           Other 3.x or later Linux (64-bit)
pfsense-virt-route-70-51:                           Other Linux (64-bit)
photon-3-haproxy-v1.2.4+vmware.1:                   VMware Photon OS (64-bit)
Ubuntu1804Template:                                 Ubuntu Linux (64-bit)
photon-3-kube-v1.18.6+vmware.1:                     VMware Photon OS (64-bit)
pfsense-virt-route-70-32:                           Other Linux (64-bit)
vCLS (10):                                          Other 3.x or later Linux (64-bit)
tkg-cluster-1-18-5b-workers-kc5xn-dd68c4685-c24l7:  VMware Photon OS (64-bit)
tkg-cluster-1-18-5b-control-plane-4f8rq:            VMware Photon OS (64-bit)
SupervisorControlPlaneVM (3):                       Other 3.x Linux (64-bit)
vSAN File Service Node (1):                         Other 3.x or later Linux (64-bit)
ubuntu-desktop:                                     Ubuntu Linux (64-bit)
SupervisorControlPlaneVM (2):                       Other 3.x Linux (64-bit)
tkg-cluster-1-18-5b-workers-kc5xn-dd68c4685-7rgkk:  VMware Photon OS (64-bit)
vSAN File Service Node (3):                         Other 3.x or later Linux (64-bit)
tkg-cluster-1-18-5b-control-plane-t59gq:            VMware Photon OS (64-bit)
haproxy-62:                                         Other 3.x or later Linux (64-bit)
vCLS (9):                                           Other 3.x or later Linux (64-bit)
tkg-cluster-1-18-5b-control-plane-gh2kt:            VMware Photon OS (64-bit)
elaine-vm:                                          Ubuntu Linux (64-bit)
vSAN File Service Node (5):                         Other 3.x or later Linux (64-bit)
tkg-cluster-1-18-5b-workers-kc5xn-dd68c4685-5v298:  VMware Photon OS (64-bit)
```