/*
 * File: main.go
 *
 * Description: Contains the main.go routine.
 */

package main

import (
	libetcd "./lib/etcd"
	"flag"
	"fmt"
	"log"
	"os"
)

/*
 * Globals
 */

// The global network namespace.
//
// TODO: right now this program treats `namespace` as a hostname or IPv4
// address, rather than an actual network namespace; this ought to be fixed
//
// TODO: this has odd issues with non-localhost values, something to
// consider for future versions
//
var namespace string

// The global etcd server.
var etcdServer *libetcd.EtcdInstance

// The global rootfs location; it gets defined via commandline argument.
var rootfs string

//! The program main function.
/*
 * @return    none
 */
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
		fmt.Printf("Error: the following directory does not exist: " + rootfs)
		return
	}

	// Attempt to start the etcd server in the background so that the
	// instances are able to store and obtain key-values.
	err := libetcd.StartEtcdServerBackgroundProcess(namespace)

	// safety check, ensure that the background etcd service has actually
	// started
	if err != nil {
		fmt.Printf(err.Error())
		fmt.Printf("The background etcd service could not be started!")
		fmt.Printf("Ensure that some other instance of the service is not " +
			"already running and in use by another process")
		return
	}

	// Otherwise the etcd service started correctly.
	fmt.Printf(" ")
	log.Println("Background etcd service started successfully.")

	// Having confirmed that the namespace and rootfs location exists,
	// background a checker loop to determine if a signal flag to terminate
	// the program is ever raised.
	go loopUtilSIGINT()

	// Give end-user a message stating that the lclusterd server started
	log.Println(" ")
	log.Println("-----------------------------")
	log.Println(" lcluster Server has started ")
	log.Println("-----------------------------")
	log.Println(" ")

	// Print out some informative information about how the namespace.
	log.Println("Network Namespace: " + namespace)

	// Print out some informative information about how the rootfs dir.
	log.Println("Rootfs Location: " + rootfs)

	// Grab the hostname, if an error occurs, end this here.
	hostname, err := os.Hostname()
	if err != nil {
		log.Println(err.Error())
		return
	}

	// Go ahead and start an etcd server instance.
	etcd_server_inst, err := libetcd.CreateEtcdInstance(namespace, rootfs)

	// Safety check, ensure that no errors have occurred during startup of
	// the EtcdServer. If it fails to start, go ahead and terminate the
	// program.
	if err != nil {
		log.Println(" ")
		log.Println("The following error has occurred: ")
		log.Println(err.Error())
		log.Println(" ")
		log.Println("Warning: Unable to start etcd server!")
		return
	}

	// Escalate the etcd server instance to become the global etcd server.
	etcdServer = etcd_server_inst

	// Mention that the etcd server has now started.
	log.Println("Etcd server startup on " + hostname)

	// Have the etcd server initialize the nodes, with a "primed" node that
	// functions as a sort of "manager" for the other nodes
	etcdServer.InitNode()

	// Mention that the node manager has now started.
	log.Println("Node manager startup successful.")

	// In order to register all of the elements in the cluster, this grpc
	// server needs to exist to have something they can return back to.
	err = startGRPCServer()

	// if an error, print it out
	if err != nil {
		fmt.Printf(err.Error())
		fmt.Printf("Error: Unable to start gRPC server on the requested port!")
	}
}

// Initialize the arg flags.
func init() {

	// Set a definition for the namespace flag.
	flag.StringVar(&namespace, "namespace", "localhost",
		"Hostname or IP address")

	// Location of the intended rootfs to run the runc libcontainer instances.
	flag.StringVar(&rootfs, "rootfs", "",
		"Rootfs POSIX directory location for runc")
}
