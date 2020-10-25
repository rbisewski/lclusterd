package libetcd

import (
	"fmt"
	"log"
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
	if err != nil {
		return err
	}

	log.Println("Scheduler sync'd with NodeManager on " + hostname)
	debugf("Primed node uuid: " + hid)

	return nil
}

//! Schedule a job on the first available node.
/*
 * @param    *Job            given job
 *
 * @return   error           error message, if any
 */
func (inst *EtcdInstance) scheduleJob(j *Job) error {

	if j == nil {
		return fmt.Errorf("scheduleJob() --> invalid input")
	}

	// Determine the current size of the queue.
	qsize := len(inst.scheduler.Queue)
	if qsize == 0 {
		return fmt.Errorf("Zero nodes are currently available...")
	}

	// Attempt to queue the job onto the node manager host.
	elmt := inst.scheduler.Queue[qsize-1]
	node := elmt.node
	err := inst.QueueJobOnNode(node.HostID, j)
	if err != nil {
		return err
	}

	log.Println("A new job was added to a node on the following host: " +
		node.HostName)

	debugf("Primed node uuid was: " + node.HostID)

	// grab the first available node
	node, err = inst.getNode(node.HostID)
	if err != nil {
		return err
	}

	elmt.node = node
	return nil
}
