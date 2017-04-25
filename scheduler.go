/*
 * File: scheduler.go
 *
 * Description: functions and definitions for the global scheduler
 */

package main

//
// Element definition
//
type Element struct {
    node     *Node
}

//
// Scheduler definition
//
type Scheduler struct {

    // holds the current queue
    Queue []*Element
}

//! Sync the global scheduler with the node manager
/*
 * @param    string    hostname
 * @param    *Node[]   array of node pointers
 *
 * @return   none
 */
func syncScheduler(hostname string, nodes []*Node) {

    // assign memory for the scheduler queue
    scheduler.Queue = make([]*Element, 0)

    // append each of the current nodes
    for _, n := range nodes {
        scheduler.Queue = append(scheduler.Queue, &Element{ node: n })
    }

    // Mention that the scheduler has began.
    stdlog("Scheduler sync'd with NodeManager on " + hostname)
}

//! Schedule a job on the first available node
/*
 * @param    EtcdInstance    etcd server
 * @param    *Job            given job
 *
 * @return   error           error message, if any
 */
func scheduleJob(srv *EtcdInstance, j *Job) error {

    // input validation
    if j == nil {
        return errorf("scheduleJob() --> invalid input")
    }

    // determine the current size of the queue
    qsize := len(scheduler.Queue)

    // if no nodes are available, do nothing
    if qsize == 0 {
        return errorf("Zero nodes are currently available...")
    }

    // grab a pointer to the end of the queue
    elmt := scheduler.Queue[qsize-1]

    // grab the node of that element
    node := elmt.node

    // attempt to queue the job onto the node manager host
    err := srv.QueueJobOnNode(node.HostID, j)

    // safety check, ensure no error has occurred
    if err != nil {
        return err
    }

    // grab the first available node
    node, err = srv.getNode(node.HostID)

    // if an error occurs, go ahead and pass it back
    if err != nil {
        return err
    }

    // since the job is scheduled on that node, go ahead and make the
    // element point to it
    elmt.node = node

    // if everything turned out alright, then this is good
    return nil
}
