/*
 * File: scheduler.go
 *
 * Description: functions and definitions for the global scheduler
 */

package libetcd

import (
    "fmt"
    "os"
)

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
 * @return   error     error message, if any
 */
func (inst *EtcdInstance) syncScheduler(hid string, nodes []*Node) error {

	// Assign memory for the scheduler queue.
	inst.scheduler.Queue = make([]*Element, 0, len(nodes))

	// Append each of the current nodes.
	for _, n := range nodes {
		inst.scheduler.Queue = append(inst.scheduler.Queue,
                  &Element{node: n})
	}

        // Attempt to grab the host name.
        hostname, err := os.Hostname()

        // if an error occurred, pass it back
        if err != nil {
            return err
        }

	// Mention that the scheduler has began.
	stdlog("Scheduler sync'd with NodeManager on " + hostname)
	debugf("Primed node uuid: " + hid)

        // pass back a nil since this was fine
        return nil
}

//! Schedule a job on the first available node.
/*
 * @param    *Job            given job
 *
 * @return   error           error message, if any
 */
func (inst *EtcdInstance) scheduleJob(j *Job) error {

	// input validation
	if j == nil {
		return fmt.Errorf("scheduleJob() --> invalid input")
	}

	// Determine the current size of the queue.
	qsize := len(inst.scheduler.Queue)

	// if no nodes are available, do nothing
	if qsize == 0 {
		return fmt.Errorf("Zero nodes are currently available...")
	}

	// Grab a pointer to the end of the queue.
	elmt := inst.scheduler.Queue[qsize-1]

	// Grab the node of that element.
	node := elmt.node

	// Attempt to queue the job onto the node manager host.
	err := inst.QueueJobOnNode(node.HostID, j)

	// safety check, ensure no error has occurred
	if err != nil {
		return err
	}

	// print out a helpful message about where the new job was queued
	stdlog("A new job was added to a node on the following host: " +
		node.HostName)
	debugf("Primed node uuid was: " + node.HostID)

	// grab the first available node
	node, err = inst.getNode(node.HostID)

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
