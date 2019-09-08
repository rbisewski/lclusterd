/*
 * File: main.go
 *
 * Description: Contains the main.go routine.
 *
 * TODO: right now this program treats `namespace` as a hostname or IPv4
 * address, rather than an actual network namespace; this ought to be fixed
 *
 * TODO: this has odd issues with non-localhost values, something to
 * consider for future versions
 */

package main

import (
	libetcd "./lib/etcd"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	namespace  string
	etcdServer *libetcd.EtcdInstance
	rootfs     string
)

func init() {
	flag.StringVar(&namespace, "namespace", "localhost",
		"Hostname or IP address")

	flag.StringVar(&rootfs, "rootfs", "",
		"Rootfs POSIX directory location for runc")
}

func main() {

	flag.Parse()

	if namespace == "" {
		namespace = "localhost"
	}

	if rootfs == "" {
		flag.Usage()
		return
	}

	// TODO: consider checking to ensure the network namespace actually
	// exists on this system.

	// Do a safety check to ensure that the rootfs is actually a
	// valid POSIX directory location, and that it actually exists.
	if !directoryExists(rootfs) {
		fmt.Printf("Error: the following directory does not exist: " + rootfs)
		return
	}

	// Attempt to start the etcd server in the background so that the
	// instances are able to store and obtain key-values.
	err := libetcd.StartEtcdServerBackgroundProcess(namespace)
	if err != nil {
		fmt.Printf(err.Error())
		fmt.Printf("The background etcd service could not be started!")
		fmt.Printf("Ensure that some other instance of the service is not " +
			"already running and in use by another process")
		return
	}

	log.Println("Background etcd service started successfully.")

	// Having confirmed that the namespace and rootfs location exists,
	// background a checker loop to determine if a signal flag to terminate
	// the program is ever raised.
	go loopUtilSIGINT()

	log.Println("lcluster Server has started")
	log.Println("Network Namespace: " + namespace)
	log.Println("Rootfs Location: " + rootfs)

	// start an Etcd server
	etcdServer, err = libetcd.CreateEtcdInstance(namespace, rootfs)
	if err != nil {
		log.Println(" ")
		log.Println("The following error has occurred: ")
		log.Println(err.Error())
		log.Println(" ")
		log.Println("Warning: Unable to start etcd server!")
		return
	}

	// register and init the "node"; i.e. a machine that runs containers
	hostname, err := os.Hostname()
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("Etcd server startup on " + hostname)
	etcdServer.InitNode()
	log.Println("Node manager startup successful.")

	// In order to register all of the elements in the cluster, this grpc
	// server needs to exist to have something they can return back to.
	err = startGRPCServer()
	if err != nil {
		fmt.Printf(err.Error())
		fmt.Printf("Error: Unable to start gRPC server on the requested port!")
	}
}
