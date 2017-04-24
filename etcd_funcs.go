/*
 * File: etcd_funcs.go
 *
 * Description: holds the etcd instance structure and funcs
 */

package main

import (
    "time"
    clientv3 "github.com/coreos/etcd/clientv3"
)

// Etcd Instance Structure
type EtcdInstance struct {

    // Internal pointer to the server itself
    internal *clientv3.Client

    // Pointer to nodes
    node *Node
}

//! Creates a new EtcdInstance and returns a pointer to it
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
                                              Endpoints: []string{etcdSocket},
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

    // define the node
    node := &Node{
        HostName: getHostname(),
        HostID:   getHostname(),
        CPUPercentRoom: 100 - hostCPURoomReserve,
        MemoryMBRoom:   0,
    }

    // Create an etcdInstance using the new internal client
    inst = &EtcdInstance{internal: newlyGeneratedClient, node: node}

    // Return the completed etcdInstance
    return inst, nil
}
