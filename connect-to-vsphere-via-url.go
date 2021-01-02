//
// functionality comes from the following packages
//
//    context - https://golang.org/pkg/context/
//    flag - https://golang.org/pkg/flag/
//    fmt - https://golang.org/pkg/fmt/
//    net/url - https://golang.org/pkg/net/url/
//    govmomi - https://github.com/vmware/govmomi

package main
  
import (
        "context"
        "flag"
        "fmt"
        "net/url"
        "github.com/vmware/govmomi"
)

func main() {

// The govmomi client requires a URL object not just a string representation of the vCenter URL.

// Declare a variable vURL that will be assigned from the flag (command line argument) named url.
        vURL := flag.String("url", "", "The URL of a vCenter server")

// Parse all of the created flags (command line arguments) to ensure that they’re read correctly or assigned their default values.
// This will catch any formatting errors with special characters, etc.

        flag.Parse()

        u, err := url.Parse(*vURL)
        if err != nil {
                fmt.Printf("Error parsing url %s\n", vURL)
                return
        }
//
// Imagine that there were multiple operations taking place such as processing some data, logging into vCenter, etc. 
// If one of the operations failed, the context would be used to share the fact that all of the other operations sharing that context needs cancelling. 
//

        ctx, cancel := context.WithCancel(context.Background())
        defer cancel()

//
// Create a vSphere/vCenter client
//
//    c, err - Return the client object c and an error object err
//    govmomi.NewClient - Call the function from the govmomi package
//    ctx - Pass in the shared context
//    u - Pass in the “parsed” url (ultimately taken from the -url string)
//    true - A Boolean value true/false that dictates if the client will tolerate an insecure certificate (self-signed)

        c, err := govmomi.NewClient(ctx, u, true)

//
//  A lot of functions in GO will typically return more than one variable/object and the majority of them will return an object of type error.
//
//  In the event of a function being successful then the function will return nil in the place of an error object. 
//  However when things go wrong then a function should create a new error object with the appropriate error details/messaging.
//
        if err != nil {
                fmt.Printf("Logging in error: %s\n", err.Error())
                return
        }

        fmt.Println("Log in successful")
        c.Logout(ctx)
}
