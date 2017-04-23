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

//! Creates the directories needed by the rootfs
/*
 * @return    error   message of the error, nil if none
 */
func (inst *EtcdInstance) generateRootfsDirs() error {

    /*
    // have the server start a new key-value storage
    kvc := client.NewKV(inst.internal)

    // grab the context
    ctx, cancel := context.WithTimeout(context.Background(),
      etcdGracePeriod * time.Second)

    // check if the /nodes dir exists
    _, err := kvc.Txn(ctx).
        If(client.Compare(client.Value(nodes_dir), "=", "nodes_dir"))
        Then().
        Else(client.OpPut(nodes_dir, "nodes_dir"))
        Commit()

    // Cancel the current context as it is no longer needed.
    cancel()

    // if an error message was found, return it
    if err != nil {
        printf("CreateEtcdInstance() --> unable to generate rootfs dirs")
        return err
    }
*/
    // otherwise everything could be created properly, so go ahead and
    // return no error
    return nil
}

//! Adds a job to the scheduler queue
/*
 * @param    string   command to be executed
 *
 * @return   uuid     newly generated job uuid
 * @return   error    error message, if any
 */
func (inst *EtcdInstance) addJobToQueue(cmd string) (string, error) {

    // ensure the job command has a valid string length
    if len(cmd) < 1 {
        return "1",errorf("addJobToQueue() --> improper command given")
    }

    // attempt to add the job to the back of the scheduler queue
    //err := scheduler.addJob(j.Command)

    // if there was an error, go ahead and pass it back
    return "1", nil
}

//! Grab the process details of a given pid
/*
 * TODO: complete this function
 *
 * @return   string     job uuid
 *
 * @return   Job        details of a given process
 * @return   error      error message, if any
 */
func (inst *EtcdInstance) getProcess(uuid string) (*Job, error) {
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
