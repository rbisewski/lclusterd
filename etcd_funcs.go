/*
 * File: etcd_funcs.go
 *
 * Description: holds the etcd instance structure and funcs
 */

package main

import (
        "./lcfg"
	clientv3 "github.com/coreos/etcd/clientv3"
	"os/exec"
	"time"
)

// Etcd instance structure, which holds a pointer to the client and nodes.
type EtcdInstance struct {

	// Internal pointer to the server itself.
	internal *clientv3.Client

	// Pointer to nodes.
	node *Node
}

//! Function to start etcd in the background.
/*
 * @return   bool      whether or not this succeeded
 */
func StartEtcdServerBackgroundProcess() bool {

	// input validation, ensure that the global network namespace value
	// gets set to something safe; note that this value is set in the main
	// routine of main.go
	if len(namespace) < 1 {
		printf("Error: Improper network namespace length!")
		return false
	}

	// Current protocol being used.
	protocol := "http://"

	// Assemble a command based on the client / server ports and given
	// global etcd server address.
	//
	// Note: this appends a 32 digit number to etcd data dir ensure the
	// given etcd session is unique.
	initial_advertise_peer_urls := protocol + namespace + lcfg.EtcdServerPort
	listen_peer_urls := protocol + namespace + lcfg.EtcdServerPort
	listen_client_urls := protocol + namespace + lcfg.EtcdClientPort
	advertise_client_urls := protocol + namespace + lcfg.EtcdClientPort
	data_dir_w_unique_cryptonum := lcfg.EtcdDataDir + spawnPseudorandomString(32)

	// Create a string array to hold all of the necessary arguments.
	var etcdArgs = []string{
		"--name",
		namespace,
		"--initial-advertise-peer-urls",
		initial_advertise_peer_urls,
		"--listen-peer-urls",
		listen_peer_urls,
		"--listen-client-urls",
		listen_client_urls,
		"--advertise-client-urls",
		advertise_client_urls,
		"--data-dir",
		data_dir_w_unique_cryptonum,
	}

	// Attempt to exec the command.
	err := exec.Command(lcfg.EtcdBinaryPath, etcdArgs...).Start()

	// if an error occurred, print it out and pass back a false
	if err != nil {
		printf(err.Error())
		printf("Error: Unable to start background etcd service!")
		return false
	}

	// otherwise if everything turned out fine, pass back a true
	printf(" ")
	stdlog("Background etcd service started successfully.")
	return true
}

//! Creates a new EtcdInstance and returns a pointer to it.
/*
 * @param   string          address to listen on; usually etcd likes to
 *                          listen on 2379; this can be adjusted in the
 *                          config for low-level testing
 *
 * @return  EtcdInstance    pointer to the obj
 *          error           message of the error, nil if none
 */
func CreateEtcdInstance(socket string) (inst *EtcdInstance, err error) {

	// Input validation, make sure this actually got a string
	if len(socket) < 1 {
		return nil, errorf("CreateEtcdInstance() --> invalid input")
	}

	// Make a client configuration for use with generating the etcd client
	// instance later on...
	etcdClientConfiguration := clientv3.Config{
		Endpoints:   []string{namespace + lcfg.EtcdClientPort},
		DialTimeout: 5 * time.Second,
	}

	// Use the above configuration to set the new client.
	newlyGeneratedClient, err := clientv3.New(etcdClientConfiguration)

	// Safety check, ensure the above config is not nil.
	if err != nil {
		return nil, errorf("CreateEtcdInstance() --> improperly " +
			"generated client due to...\n" +
			err.Error())
	}

	// Assign the details of the new node as per...
	//
	// HostName: hostname of the server
	// HostID:   n_XYZ
	//
	// where XYZ is just a random crypto num of base64 w/ roughly 16
	// digits.
	//
	node := &Node{
		HostName: getHostname(),
		HostID:   "n_" + spawnPseudorandomString(16),
	}

	// Create an etcdInstance using the new internal client.
	inst = &EtcdInstance{internal: newlyGeneratedClient, node: node}

	// Return the completed etcdInstance.
	return inst, nil
}
