# govmomi-examples #

A few example govmomi scripts, mostly involving how to connect to vSphere and retrieve some infrastructure information. I've tried to add as many comment to the code as possible as there is not a lot of documentation on how to code govmomi. There are (some other examples here)<https://pkg.go.dev/github.com/vmware/govmomi/view#pkg-examples>. 

## Getting started ##

For first time use, run the following:

```shell
% export GO111MODULE=on
% go get k8s.io/client-go@master
% go mod init 
% go build <Filename>.go 
% go run <Filename>.go
```

For subsequent runs, once the imports are local, simply run:

```shell
 % go run <Filename>.go
```

Notes on running "go build": <https://github.com/kubernetes/client-go/blob/master/INSTALL.md#for-the-casual-user>

## About the scripts ##

Two of the modules show the different ways to connect to vSphere

- via URL
- via environment variables

The other modules show how to connect and retrieve various vSphere information, e.g.

- Datacenter
- Cluster
- Hosts
- Datastores
- Virtual Machines (VMs)
- First Class Disks (FCDs) - used to back Kubernetes Persistent Volumes

Finally we have two modules that use a combinatation of vSphere and Kubernetes Code modules:

- Return K8s nodes running on a vSphere infrastructure
- Return PCI devices on an ESXi device host where a Kubernetes node/VM runs

## Sample outputs ##

Here are some example outputs, assuming the required vSphere `environment variables` have been set appropriately in the shell.

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

```shell
$ go run conn-to-v-via-e-find-fcd.go
DEBUG: vc is vcsa-06.rainpole.com/sdk
DEBUG: user is administrator@vsphere.local
DEBUG: password is ************

Log in successful (govmomi)


Found default datacenter:  Datacenter:datacenter-3 @ /OCTO-Datacenter


Got datastore, number of datastore(s) is :  3

Found host: Datastore:datastore-11 @ /OCTO-Datacenter/datastore/PureVMFSDatastore
Found host: Datastore:datastore-38 @ /OCTO-Datacenter/datastore/isilon-01
Found host: Datastore:datastore-33 @ /OCTO-Datacenter/datastore/vsanDatastore

Datastores found: 3

Found Datastore: PureVMFSDatastore
Found Datastore: isilon-01
Found Datastore: vsanDatastore

Log in successful (vim25)

List of FCDs on datastore: PureVMFSDatastore

  Found FCD Id: b7698784-1f52-4ae0-a8b6-ba88838e3513
  FCD Name              : pvc-65a6b9c4-844e-4552-90dc-495168d745fc
  FCD Creation Time     : 2020-06-11 12:48:05.575642 +0000 UTC
  FCD Size (MB)         : 4769
  FCD Consumption Type  : [disk]
  FCD Datastore Type    : Datastore
  FCD Datastore Info    : datastore-11
  FCD FilePath          : [PureVMFSDatastore] fcd/96c4df4d2f9a429ebb062f94b4ac3d21.vmdk
  FCD Backing Object Id :
  FCD Delta Size (MB)   : 0


List of FCDs on datastore: isilon-01


List of FCDs on datastore: vsanDatastore

  Found FCD Id: 19db43c4-1713-453c-a1f0-e1d6482b60d4
  FCD Name              : pvc-e3f6dd59-cbc0-49a7-97c8-d92a26732c43
  FCD Creation Time     : 2020-10-23 13:08:53.401297 +0000 UTC
  FCD Size (MB)         : 1024
  FCD Consumption Type  : [disk]
  FCD Datastore Type    : Datastore
  FCD Datastore Info    : datastore-33
  FCD FilePath          : [vsanDatastore] 038f6b5f-8122-d3af-eabe-246e962c240c/b39bcacc6ff143439f9cd6b7454999e4.vmdk
  FCD Backing Object Id : e5d5925f-fff6-9e09-0b30-246e962f4854
  FCD Delta Size (MB)   : 0

  Found FCD Id: 19e07b27-e02c-4366-bb1d-772fe3c9a4f3
  FCD Name              : pvc-27197aab-9c6b-4cb7-b4a6-dba4a3b3429d
  FCD Creation Time     : 2020-10-23 13:10:15.922917 +0000 UTC
  FCD Size (MB)         : 1024
  FCD Consumption Type  : [disk]
  FCD Datastore Type    : Datastore
  FCD Datastore Info    : datastore-33
  FCD FilePath          : [vsanDatastore] 038f6b5f-8122-d3af-eabe-246e962c240c/784d461b95bf43bd9f69177ba9813ac5.vmdk
  FCD Backing Object Id : 37d6925f-765c-7317-e507-246e962f5274
  FCD Delta Size (MB)   : 0

  Found FCD Id: 499eee6a-ee1a-4793-9ffa-3d9a1c2518a9
  FCD Name              : pvc-b2cdfd8f-24bc-487b-ad02-a749d985c19b
  FCD Creation Time     : 2020-10-13 14:14:05.747654 +0000 UTC
  FCD Size (MB)         : 1024
  FCD Consumption Type  : [disk]
  FCD Datastore Type    : Datastore
  FCD Datastore Info    : datastore-33
  FCD FilePath          : [vsanDatastore] 038f6b5f-8122-d3af-eabe-246e962c240c/efbb39583d25442b8d964216adbcbf2a.vmdk
  FCD Backing Object Id : 2db6855f-13f0-ffec-1912-246e962f4854
  FCD Delta Size (MB)   : 0

  Found FCD Id: c8fbb21f-c380-4bf5-af24-699b0ef4665c
  FCD Name              : pvc-73752334-c3c0-4be2-9eb8-2192c1197a6b
  FCD Creation Time     : 2020-12-14 14:38:53.601105 +0000 UTC
  FCD Size (MB)         : 1024
  FCD Consumption Type  : [disk]
  FCD Datastore Type    : Datastore
  FCD Datastore Info    : datastore-33
  FCD FilePath          : [vsanDatastore] fc78d75f-dd14-9bce-9e2f-246e962f4854/a3277e06b1094ddf959515e7835345a6.vmdk
  FCD Backing Object Id : fd78d75f-38bf-154d-1bfd-246e962f4854
  FCD Delta Size (MB)   : 0

  Found FCD Id: c9e2c645-0117-4ac0-9a18-8f3fdd4e838b
  FCD Name              : pvc-57a4d167-a5aa-47a0-9852-108d65338ad1
  FCD Creation Time     : 2020-10-23 13:07:33.602157 +0000 UTC
  FCD Size (MB)         : 1024
  FCD Consumption Type  : [disk]
  FCD Datastore Type    : Datastore
  FCD Datastore Info    : datastore-33
  FCD FilePath          : [vsanDatastore] 038f6b5f-8122-d3af-eabe-246e962c240c/fc8b344c69d244988368fa9a173de44e.vmdk
  FCD Backing Object Id : 95d5925f-b256-186a-cd15-246e962c240c
  FCD Delta Size (MB)   : 0
```
