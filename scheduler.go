/*
 * File: scheduler.go
 *
 * Description: definition of the Scheduler structure 
 */

package main

import (
    "strconv"
    pb "./lclusterpb"
)

//
// Scheduler structure
//
type Scheduler struct {

    // Array of pointers to the currently queued jobs
    Queue []*Job
}


//! Adds a job to the queue of the scheduler
/*
 * @param    string    command line argument to execute; i.e. 'ls'
 *
 * @return   error     error message, if any
 */
func (sch *Scheduler) addJob(cmd string) error {

    // input validation
    if len(cmd) < 1 {
        return errorf("addJob() --> invalid input")
    }

    // attempt to assemble a job object
    j := &Job{Command: cmd}

    // append the job to the back of the queue
    sch.Queue = append(sch.Queue, j)

    // Announce that the job was added to the scheduler since
    stdlog("                                 ")
    stdlog("---------------------------------")
    stdlog("A new job was added to the queue.")
    stdlog("                                 ")
    stdlog("Command: " + cmd)
    stdlog("---------------------------------")
    stdlog("                                 ")

    // if the program got to this point, assume everything worked as
    // expected...
    return nil
}


//! Removes a job to the queue of the scheduler
/*
 * @param    int64    uuid of the job in question
 *
 * @return   error    error message, if any
 */
func (sch *Scheduler) removeJob(uuid int64) error {

    // input validation
    if uuid < 1 {
        return errorf("removeJob() --> invalid or expired process")
    }

    // Cycle thru all of the jobs...
    for index, j := range sch.Queue {

        // If a job with that uuid is discovered...
        if j.Uuid == strconv.FormatInt(uuid, 10) {

            // attempt to remove it from the jobs queue
            sch.Queue = append(sch.Queue[:index], sch.Queue[index+1:]...)

            // as the job was removed successful, return nil
            return nil
        }
    }

    // otherwise request to remove a job of uuid failed, so pass back an
    // error that alerts the end-user
    return errorf("removeJob() --> failed to remove non-existing job " +
                  "process w/ uuid: " + strconv.FormatInt(uuid, 10))
}


//! Pops the first job off the front of the queue, then hands off to a node
/*
 * @return    error    error message, if any
 */
func (sch *Scheduler) runJob() error {

    // if the queue is empty, do nothing
    if len(sch.Queue) < 1 {
        return nil
    }

    // Variable declaration
    var j *Job
    var thereAreZeroNodes bool = false

    // pop the first element off of the queue
    j, sch.Queue = sch.Queue[len(sch.Queue)-1], sch.Queue[:len(sch.Queue)-1]

    // safety check, ensure the global node manager is still active
    if nodeManager == nil {
        return errorf("Error: fatal response as global node manager inactive!")
    }

    // safety check, ensure the node manager actually has nodes
    if len(nodeManager.Nodelist) < 1 {
        thereAreZeroNodes = true
    }

    // in the case where the node manager does *not* actually have nodes...
    if thereAreZeroNodes {

        // assemble a start job request
        sjr := pb.StartJobRequest{Command: j.Command, Machine: j.Machine}

        // tell the node manager to start a container
        tmpNode, errDuringNodeStartup := nodeManager.startNewNode(sjr, rootfs)

        // check if an error occurred during node generation
        if errDuringNodeStartup != nil {

            // if it did, past it back
            return errDuringNodeStartup
        }

        // append it to the node manager
        nodeManager.Nodelist = append(nodeManager.Nodelist, tmpNode)

        // TODO: reset the grace period to prevent sitations involving too
        // many nodes being generated at once

        // and leave the function since the process was passed along to
        // this newly generated node
        return nil
    }

    // since the node manager is active, attempt to grab an available node
    // from the cluster
    n := nodeManager.grabNode()

    // start the given job on the first available node
    err := n.startJob(j)

    // if an error occurs during job execution, go ahead and pass it back
    if err != nil {
        return err
    }

    // else having successfully passed the job to the node manager, consider
    // this completed successfully
    return nil
}
