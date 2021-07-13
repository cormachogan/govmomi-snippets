You may need to run go mod init <name> to begin, if the working directory is not part of a module.
You may also need an "export GO111MODULE=on" to pull the appropriate packages.

If you get errors similar to:

get-gpu-candidates.go:45:2: module k8s.io/client-go@latest found (v1.5.2), but does not contain package k8s.io/client-go/kubernetes
get-gpu-candidates.go:46:2: module k8s.io/client-go@latest found (v1.5.2), but does not contain package k8s.io/client-go/tools/clientcmd
get-gpu-candidates.go:47:2: module k8s.io/client-go@latest found (v1.5.2), but does not contain package k8s.io/client-go/util/homedir

... you may need to edit the go.mod file to add versions of k8s.io/api and k8s.io/client-go

require (
	github.com/vmware/govmomi v0.26.0 // indirect
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2 // indirect
	k8s.io/client-go v0.21.2
)
