/*
 * File: main.go
 *
 * Description: Contains the main.go routine.
 */

package main

import "flag"

//
// Globals
//

// Network Namespace
//
// TODO: right now this program treats `namespace` as a hostname or IPv4
// address, rather than an actual network namespace; this ought to be fixed
//
var namespace string

// Etcd Server
var etcdServer *EtcdInstance

// Scheduler
var scheduler *Scheduler

// List of current processes
var processesList map[int64]*Process

// Rootfs Location
var rootfs string

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

    // attempt to start the etcd server in the background so that the
    // instances are able to store and obtain key-values.
    etcdBgSuccessfullyStarted := StartEtcdServerBackgroundProcess()

    // safety check, ensure that the background etcd service has actually
    // started
    if !etcdBgSuccessfullyStarted {
        printf("The background etcd service could not be started!")
        printf("Ensure that some other instance of the service is not " +
               "already running and in use by another process")
        return
    }

    // Having confirmed that the namespace and rootfs location exists,
    // background a checker loop to determine if a signal flag to terminate
    // the program is ever raised.
    go loopUtilSIGINT()

    // Give end-user a message stating that the lclusterd server started
    stdlog(" ")
    stdlog("-----------------------------")
    stdlog(" lcluster Server has started ")
    stdlog("-----------------------------")
    stdlog(" ")

    // Print out some informative information about how the namespace.
    stdlog("Network Namespace: " + namespace)

    // Print out some informative information about how the rootfs dir.
    stdlog("Rootfs Location: " + rootfs)

    // Go ahead and start the Scheduler.
    scheduler = &Scheduler{}

    // Mention that the global scheduler has now started.
    stdlog("Scheduler startup on " + getHostname())

    // Go ahead and start the etcd server instance.
    etcd_server_inst, err := CreateEtcdInstance(namespace+etcdClientPort)

    // Safety check, ensure that no errors have occurred during startup of
    // the EtcdServer. If it fails to start, go ahead and terminate the
    // program.
    if err != nil {
        stdlog(" ")
        stdlog("The following error has occurred: ")
        stdlog(err.Error())
        stdlog(" ")
        stdlog("Warning: Unable to start etcd server!")
        return
    }

    // Escalate the etcd server instance to become the global etcd server.
    etcdServer = etcd_server_inst

    // Mention that the etcd server has now started.
    stdlog("Etcd server startup on " + getHostname())

    // Have the etcd server initialize the nodes, with a "primed" node that
    // functions as a sort of "manager" for the other nodes
    etcdServer.InitNode()

    // Mention that the node manager has now started.
    stdlog("Node manager startup successful.")

    // In order to register all of the elements in the cluster, this grpc
    // server needs to exist to have something they can return back to.
    startServerInstanceOfGRPC()
}
