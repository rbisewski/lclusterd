/*
 * File: main.go
 *
 * Description: Contains the main.go routine.
 */

package main

import (
    "flag"
)

//
// Globals
//

// Network Namespace
var namespace string

// Etcd Server
var etcdServer *EtcdInstance

// Node Manager
var nodeManager *NodeManager

// Scheduler
var scheduler *Scheduler

// Rootfs Location
var rootfs string

// Incremental Node number
var incrementalNodeNumber int64 = 0

// Initialize the flags beforehand.
func init() {

    // Set a definition for the namespace flag.
    flag.StringVar(&namespace, "namespace", "localhost",
      "Hostname or IP address")

    // Location of the intended rootfs to run the runc libcontainer instances.
    flag.StringVar(&rootfs, "rootfs", "",
      "Rootfs POSIX directory location for runc")
}

//
// Main
//
func main() {

    // Ensure that no more than 2 arguments are given, else print the usage
    // information.
    if flag.NArg() > 2 {
        flag.Usage()
        return
    }

    // Parse the given argument flags.
    flag.Parse()

    // Network namespace? Use that then, else default to localhost.
    if namespace == "" {
        namespace = "localhost"
    }

    // TODO: consider checking to ensure the network namespace actually
    // exists on this system.

    // Ensure that the end-user has provided a rootfs directory location.
    if rootfs == "" {
        flag.Usage()
        return
    }

    // Do a safety check to ensure that the rootfs is actually a
    // valid POSIX directory location, and that it actually exists.
    if !directoryExists(rootfs) {
        printf("Error: the following directory does not exist: " + rootfs)
        return
    }

    // Having confirmed that the namespace and rootfs location exists,
    // background a checker loop to determine if a signal flag to terminate
    // the program is ever raised.
    go loopUtilSIGINT()

    // Give end-user a message stating that the lclusterd server started
    printf("\n")
    stdlog("-----------------------------")
    stdlog(" lcluster Server has started ")
    stdlog("-----------------------------")
    stdlog(" ")

    // Print out some informative information about how the namespace.
    stdlog("Network Namespace: " + namespace)

    // Print out some informative information about how the rootfs dir.
    stdlog("Rootfs Location: " + rootfs)

    // Go ahead and start the etcd server instance.
    etcd_server_inst, err := CreateEtcdInstance(etcdSocket)

    // Safety check, ensure that no errors have occurred during startup of
    // the EtcdServer. If it fails to start, go ahead and terminate the
    // program.
    if err != nil {
        stdlog(err.Error())
        stdlog("Warning: Unable to start etcd server!")
        return
    }

    // Esclate the etcd server instance to become the global etcd server.
    etcdServer = etcd_server_inst

    // Mention that the etcd server has now started.
    stdlog("Etcd server startup successful.")

    // Go ahead and start the Scheduler.
    queue       := make([]*Job, 0)
    scheduler    = &Scheduler{Queue: queue}

    // Go ahead and start the Node Manager.
    nl          := make([]*Node,0)
    nodeManager  = &NodeManager{Nodelist: nl}

    // Mention that the node manager has now started.
    stdlog("Node manager startup successful.")

    // In order to register all of the elements in the cluster, this grpc
    // server needs to exist to have something they can return back to.
    startServerInstanceOfGRPC()

    // Mention that the grpc server has now started.
    stdlog("gRPC server startup successful.")
}
