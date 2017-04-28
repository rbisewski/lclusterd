/*
 * File: scheduler.go
 *
 * Description: functions and definitions for the global scheduler
 */

package main

// The element definition, which are individuals in the queue.
type Element struct {
	node *Node
}

// The scheduler definition, which holds the queue.
type Scheduler struct {

	// holds the current queue
	Queue []*Element
}

//! Sync the global scheduler with the node manager.
/*
 * @param    string    primed node uuid
 * @param    *Node[]   array of node pointers
 *
 * @return   none
 */
func syncScheduler(hid string, nodes []*Node) {

	// Assign memory for the scheduler queue.
	scheduler.Queue = make([]*Element, 0)

	// Append each of the current nodes.
	for _, n := range nodes {
		scheduler.Queue = append(scheduler.Queue, &Element{node: n})
	}

	// Mention that the scheduler has began.
	stdlog("Scheduler sync'd with NodeManager on " + getHostname())
	debugf("Primed node uuid: " + hid)
}

//! Schedule a job on the first available node.
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

	// Determine the current size of the queue.
	qsize := len(scheduler.Queue)

	// if no nodes are available, do nothing
	if qsize == 0 {
		return errorf("Zero nodes are currently available...")
	}

	// Grab a pointer to the end of the queue.
	elmt := scheduler.Queue[qsize-1]

	// Grab the node of that element.
	node := elmt.node

	// Attempt to queue the job onto the node manager host.
	err := srv.QueueJobOnNode(node.HostID, j)

	// safety check, ensure no error has occurred
	if err != nil {
		return err
	}

	// print out a helpful message about where the new job was queued
	stdlog("A new job was added to a node on the following host: " +
		node.HostName)
	debugf("Primed node uuid was: " + node.HostID)

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
