/*
 * File: etcd_funcs.go
 *
 * Description: holds the etcd instance structure and funcs
 */

package main

import (
    "fmt"
    client "github.com/coreos/etcd/client"
)

// Etcd Instance Structure
type EtcdInstance struct {

    // Internal pointer to the server itself
    internal *client.Client
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
        return nil, fmt.Errorf("CreateEtcdInstance() --> invalid input")
    }

    // Make a client configuration for use with generating the etcd client
    // instance later on...
    etcdClientConfiguration := client.Config{
                                              Endpoints: []string{etcdSocket},
                                            }

    // Use the above configuration to set the new client.
    newlyGeneratedClient, err := client.New(etcdClientConfiguration)

    // Safety check, ensure the above config is not nil.
    if err != nil {
        return nil, fmt.Errorf("CreateEtcdInstance() --> improperly " +
                               "generated client due to...\n" +
                               err.Error())
    }

    // Create an etcdInstance using the new internal client
    inst = &EtcdInstance{internal: &newlyGeneratedClient}

    // Return the completed etcdInstance
    return inst, nil
}

//! Adds a job to the scheduler queue
/*
 * TODO: complete this function
 *
 * @param    Job      job to add to queue
 *
 * @return   int64    process id
 * @return   error    error message, if any
 */
func (inst *EtcdInstance) addJobToQueue(j *Job) (int64, error) {
    return -1, nil
}

//! Grab the process details of a given pid
/*
 * TODO: complete this function
 *
 * @return   int64      process id
 *
 * @return   Process    details of a given process
 * @return   error      error message, if any
 */
func (inst *EtcdInstance) getProcess(pid int64) (*Job, error) {
    return &Job{}, nil
}

//! Halts a job that is currently running
/*
 * TODO: complete this function
 *
 * @param    Job      job to halt
 *
 * @return   error    error message, if any
 */
func (inst *EtcdInstance) haltJob(j *Job) (error) {
    return nil
}
