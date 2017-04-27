/*
 * File: node_manager.go
 *
 * Description: functions to handle nodes and node creation
 */

package main

import (
	"encoding/json"
	"path"
	"strconv"
	"strings"
	"time"

	clientv3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
)

// The node definition.
type Node struct {
	HostName string
	HostID   string
}

//! Update the TTL for a given reserved entry.
/*
 * @param    LeaseID    id of a given etcd key entry
 *
 * @return   none
 */
func (inst *EtcdInstance) updateTTL(leaseID clientv3.LeaseID) {

	// Keep the entry alive.
	_, err := inst.internal.KeepAliveOnce(context.TODO(), leaseID)

	// if an error occurs, dump it to stdout
	if err != nil {
		stdlog(err.Error())
		stdlog("updateTTL() --> could not update TTL")
	}
}

//! Function to keep etcd key values alive.
/*
 * @param    LeaseGrantResponse    lease response
 *
 * @return   none
 */
func (inst *EtcdInstance) keepKeyAlive(lease *clientv3.LeaseGrantResponse) {

	// Set a duration time based on the time-to-live, this will be a
        // sort of 'duration' time.
	sleep_duration := time.Duration(lease.TTL / 2)

	// infinite loop
	for {

		// update the time-to-live
		inst.updateTTL(lease.ID)

		// then go back to sleep
		time.Sleep(time.Second * sleep_duration)
	}
}

//! Prime a given node to make it ready for jobs.
/*
 * @param    chan    channel for listening who is primed
 *
 * @return   none
 */
func (inst *EtcdInstance) primedLock(primedNotificationChan chan bool) {

	// Grant the lease.
	lease, err := inst.internal.Grant(context.TODO(), primedTTL)

	// if an error occurs, print it out
	if err != nil {
		stdlog("primedLock(): failed to issue a TTL for the future primed lock: " + err.Error())
		return
	}

	// Setup a key-value store.
	kvc := clientv3.NewKV(inst.internal)

	// Grab the background context.
	ctx, cancel := context.WithTimeout(context.Background(), etcdGracePeriod*time.Second)

	// Check if the value exists, and insert it if it does not.
	response, err := kvc.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(primed), "=", "primed")).
		Then().
		Else(clientv3.OpPut(primed, "primed", clientv3.WithLease(lease.ID))).
		Commit()

	// since this is done, cancel the current context
	cancel()

	// if an error occurred during the, print it out
	if err != nil {
		stdlog("primedLock() --> error occured whilst priming lock")
		stdlog(err.Error())
		return
	}

	// if no other node response, this probably means that we are primed
	// node, so keep the lease entry alive
	if !response.Succeeded {
		go inst.keepKeyAlive(lease)
		primedNotificationChan <- true
	}

	// if a node *did* response, then this node is not the primed node; go
	// ahead and clean up the lease ref whilst waiting for a turn...
	if response.Succeeded {
		inst.internal.Revoke(context.TODO(), lease.ID)
		go inst.watchUntilPrimed(primedNotificationChan)
	}

	// if the node got this far, close the notification channel
	primedNotificationChan <- false
}

//! Listens on the notification channel until the current primed node
//! has completed it's job.
/*
 * @param    chan    notification channel
 *
 * @return   none
 */
func (inst *EtcdInstance) watchUntilPrimed(primedNotificationChan chan bool) {

	// Check for a channel response.
	responsing_chan := inst.internal.Watch(context.Background(), primed)

	// infinite, keep running until this node gets to be primed
	for {

		// Grab the response from the channel.
		wresponse := <-responsing_chan

		// Cycle thru the events.
		for _, ev := range wresponse.Events {

			// if not a deleted prime event, skip it...
			if ev.Type != mvccpb.DELETE {
				continue
			}

			// otherwise the primed node is complete, so this node needs to
			// vie for prime lock
			stdlog("watchUntilPrimed(): main node no longer primed, " +
				"attempting to prime this node...")
			inst.primedLock(primedNotificationChan)
			return
		}
	}
}

//! Prime this node for any up-in-coming job that is currently in the queue.
/*
 * @param     chan    notification channel
 *
 * @return    none
 */
func (inst *EtcdInstance) primeThisNode(notify chan bool) {

	// infinite loop
	for {

		// Wait for the channel to notify this.
		isPrimed := <-notify

		// not yet primed? do nothing...
		if !isPrimed {
			continue
		}

		// Assemble the nodes via the list.
		nodes, err := inst.obtainListOfNodes()

		// if an error occurs, print it out
		if err != nil {
			stdlog("primeThisNode() --> node gathering failed\n " +
				err.Error())
			continue
		}

		// Generate the job queue
		err = inst.initializeJobQueue()

		// if an error occurs, print it
		if err != nil {
			stdlog("primeThisNode() --> unable to create job queue\n" +
				err.Error())
			continue
		}

		// reserve a global id for jobs
		err = inst.initializeGlobalJobID()

		// if an error occurs, print it
		if err != nil {
			stdlog("primeThisNode() --> unable to set global id\n" +
				err.Error())
			continue
		}

		// setup the directory for processes to be ran on nodes
		err = inst.initializeProcessStorage()

		// if an error occurred, print it out
		if err != nil {
			stdlog("primeThisNode() --> unable to set aside storage for " +
				"processes\n" + err.Error())
			continue
		}

		// with all of the pieces in place, this syncs with the scheduler,
		// and the entire setup is primed for new jobs
		syncScheduler(inst.node.HostID, nodes)
		go inst.watchGeneralJobQueue()
	}
}

//! Setup a job queue so that the end-user can add jobs to it.
/*
 * @return   error   error message, if any
 */
func (inst *EtcdInstance) initializeJobQueue() error {

	// Setup a key-value store for the queue elements.
	kvc := clientv3.NewKV(inst.internal)

	// Grab the context.
	ctx, cancel := context.WithTimeout(context.Background(),
		etcdGracePeriod*time.Second)

	// Ensure the queue dirs exist, and if not, create them.
	_, err := kvc.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(queue_dir), "=", "queue")).
		Then().
		Else(clientv3.OpPut(queue_dir, "queue")).
		Commit()

	// cancel the context since it is no longer needed
	cancel()

	// if an error occurred, print it out
	if err != nil {
		stdlog("initializeJobQueue() --> error occured while trying to " +
			"initialize job queue:\n" + err.Error())
		return err
	}

	// if everything worked, pass back nil
	return nil
}

//! Setup a global id for potential jobs.
/*
 * @return    error    error message, if any
 */
func (inst *EtcdInstance) initializeGlobalJobID() error {

	// Setup a key value storage.
	kvc := clientv3.NewKV(inst.internal)

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		etcdGracePeriod*time.Second)

	// Check if a jobs dir exists, else create one.
	_, err := kvc.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(jobs_dir), "=", "0")).
		Then().
		Else(clientv3.OpPut(jobs_dir, "0")).
		Commit()

	// cancel the current context if we no longer need it
	cancel()

	// if an error occurred, print it out
	if err != nil {
		stdlog("initializeGlobalJobID() --> error occured while trying to " +
			"setup the global job id: " + err.Error())
		return err
	}

	// if everything turned out alright, pass back nil
	return nil
}

//! Setup the processes storage.
/*
 * @param    error
 */
func (inst *EtcdInstance) initializeProcessStorage() error {

	// setup the key-value store
	kvc := clientv3.NewKV(inst.internal)

	// grab the current context
	ctx, cancel := context.WithTimeout(context.Background(),
		etcdGracePeriod*time.Second)

	// check if the processes location exists, and if not, creates it
	_, err := kvc.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(processes_dir), "=", "processes")).
		Then().
		Else(clientv3.OpPut(processes_dir, "processes")).
		Commit()

	// cancel the current context as this no longer needs it
	cancel()

	// if an error occurs, print it out
	if err != nil {
		stdlog("initializeProcessStorage() --> unable to make process " +
			"storage\n" + err.Error())
		return err
	}

	// if no error occurred, then pass back nil
	return nil
}

//! Initialize the nodes.
/*
 * @return    chan    newly initialized node comm channel
 * @return    error   error message, if any
 */
func (inst *EtcdInstance) InitNode() (chan bool, error) {

	// Assign a chunk of memory for any potential processes
	processesList = make(map[int64]*Process)

	// Connect this with the nodes list.
	err := inst.addToNodesList()

	// if any error, pass it back
	if err != nil {
		return nil, err
	}

	// Assign memory for the channel.
	notify := make(chan bool, 1)

	// wait around until this node can be primed
	go inst.primeThisNode(notify)
	inst.primedLock(notify)

	// pass back a ref to the notif
	return notify, nil
}

//! Setup this node and attach it to the list.
/*
 * @return    error   error message, if any
 */
func (inst *EtcdInstance) addToNodesList() error {

	// Marshal a given node.
	mresult, err := json.Marshal(inst.node)

	// if an error occurred, print it out
	if err != nil {
		stdlog(err.Error())
		return err
	}

	// Grant a lease.
	lease, err := inst.internal.Grant(context.TODO(), nlistTTL)

	// if an error occurred, print it out
	if err != nil {
		stdlog("addToNodesList() --> unable to give TTL for node " +
			inst.node.HostID + "\n" +
			err.Error())
		return err
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		etcdGracePeriod*time.Second)

	// insert the key into etcd
	_, err = inst.internal.Put(ctx, path.Join(nodes_dir, inst.node.HostID),
		string(mresult), clientv3.WithLease(lease.ID))

	// cancel the current context as this no longer needs it
	cancel()

	// revoke the lease if an error occurred
	if err != nil {
		inst.internal.Revoke(context.TODO(), lease.ID)
		return err
	}

	// Setup a local jobs queue for the node.
	jobQueue := path.Join(nodes_dir, inst.node.HostID, "jobs")

	// grab the current context
	ctx, cancel = context.WithTimeout(context.Background(),
		etcdGracePeriod*time.Second)

	// add the context into the k-v storage as well
	_, err = inst.internal.Put(ctx, jobQueue, "")

	// cancel the current context as it is no long needed
	cancel()

	// if an error occurs, pass it back
	if err != nil {
		return err
	}

	// keep the key entry alive
	go inst.keepKeyAlive(lease)

	// keep tabs on the job queue
	go inst.watchInternalJobQueue()

	// everything works, so pass back nil
	return nil
}

//! Append a process to the list of processes.
/*
 * @param    Process    requested process
 *
 * @result   error      error message, if any
 */
func (inst *EtcdInstance) storeProcess(p *Process) error {

	// utilize the uuid as the hash key; add the process to the list of
	// processes
	processesList[p.uuid] = p

	// Attempt to marshell a node.
	mresult, err := json.Marshal(inst.node)

	// if an error occurs, print it out
	if err != nil {
		stdlog(err.Error())
		return err
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		etcdGracePeriod*time.Second)

	// insert a key-value into etcd
	_, err = inst.internal.Put(ctx, path.Join(processes_dir,
		strconv.FormatInt(p.uuid, 10)), string(mresult))

	// cancel the current context, as it is no longer needed
	cancel()

	// print out the resulting error, if any
	if err != nil {
		stdlog(err.Error())
		return err
	}

	// success, so pass back nil
	return nil
}

//! Grab the process from the global list of processes.
/*
 * @param    int64    process uuid
 *
 * @return   Process  the process information
 * @return   error    error message, if any
 */
func (inst *EtcdInstance) obtainProcess(uuid int64) (p *Process, err error) {

	// ensure the uuid is something reasonable
	if uuid < 0 {
		return nil, errorf("obtainProcess() --> invalid input")
	}

	// otherwise return the process in question
	return processesList[uuid], nil
}

//! Allows the node to watch internal queue, in case there are more jobs.
/*
 * @return    none
 */
func (inst *EtcdInstance) watchInternalJobQueue() {

	// Assemble the job queue path.
	jobQueue := path.Join(nodes_dir, inst.node.HostID, "jobs")

	// Watch the channel for a response.
	rchan := inst.internal.Watch(context.Background(), jobQueue,
		clientv3.WithPrefix())

	// infinite loop
	for {

		// Grab responses from the channel.
		wresponse := <-rchan

		// Cycle thru all given events.
		for _, ev := range wresponse.Events {

			// this function is only interested in put events
			if ev.Type != mvccpb.PUT {
				continue
			}

			// Unmarshal the data, if any.
			var j Job
			err := json.Unmarshal(ev.Kv.Value, &j)

			// print out the error message, if any
			if err != nil {
				stdlog("watchInternalJobQueue() --> failed to read job data")
				stdlog(err.Error())
				return
			}

			// Using the job data, attempt to start a process
                        // on the node.
			p, err := StartProcess(j)

			// print out the error message, if any
			if err != nil {
				stdlog("watchInternalJobQueue() --> unable to start process")
				stdlog(err.Error())
				return
			}

			// store a ref to the process into etcd
			err = inst.storeProcess(p)

			// print out the error message, if any
			if err != nil {
				stdlog("watchInternalJobQueue() --> could not store process " +
					"data into etcd")
				stdlog(err.Error())
				return
			}

			// otherwise the job started successful, so go ahead and print
			// out a helpful message
			stdlog("New job started (" + j.Path + ") in a node on cluster " +
				"host: " + inst.node.HostName)
			debugf("Primed node that received job has uuid: " +
				inst.node.HostID)
		}
	}
}

//! The primed node needs to watch the queue for jobs to handle.
/*
 * @return    none
 */
func (inst *EtcdInstance) watchGeneralJobQueue() {

	// Watch the queue in case a new job appears.
	rch := inst.internal.Watch(context.Background(), queue_dir,
		clientv3.WithPrefix())

	// infinite loop
	for {

		// Look for responses in the channel.
		wresponse := <-rch

		// Cycle thru all of the events.
		for _, ev := range wresponse.Events {

			// this function is only interested in put events
			if ev.Type != mvccpb.PUT {
				continue
			}

			// Attempt to unmarshal the job data.
			var j Job
			err := json.Unmarshal(ev.Kv.Value, &j)

			// if an error occurred, print it out
			if err != nil {
				stdlog("watchGeneralJobQueue() --> unable to read job data")
				stdlog(err.Error())
			}

			// attempt to schedule the job
			err = scheduleJob(inst, &j)

			// if an error occurred, print it out
			if err != nil {
				stdlog("watchGeneralJobQueue() --> unable to schedule job")
				stdlog(err.Error())
			}

			// print out a helpful message
			stdlog("Node manager has detected a scheduled job.")
		}
	}
}

//! Add a job to the scheduler queue.
/*
 * @param    Job      given job
 *
 * @return   int64    uuid
 * @return   error    error message, if any
 */
func (inst *EtcdInstance) addToGlobalQueue(j *Job) (int64, error) {

	// input validation
	if j == nil {
		return -1, errorf("addToGlobalQueue() --> invalid input")
	}

	// further check, ensure the instance is safe
	if inst.internal == nil {
		return -1, errorf("addToGlobalQueue() --> malformed etcd instance")
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		etcdGracePeriod*time.Second)

	// Grab the list of queued jobs.
	response, err := inst.internal.Get(ctx, queue_dir)

	// if debug mode...
	if debugMode {

		// cycle thru all of the current jobs for the benefit of the developer
		debugf("Current queued jobs are as follows:")
		for i, ent := range response.Kvs {
			debugf(strconv.Itoa(i+1) + ") " + string(ent.Value))
		}
	}

	// cancel the context as it is no longer needed
	cancel()

	// if an error occurs, pass it back; note that if an error occurs at
	// this point it is likely a connection issue exists, hence the need to
	// check for a proper contextual response
	if err != nil {
		stdlog(err.Error())
		return -1, err
	}

	// Use the nanosecond timestamp as an Uuid.
	//
	// TODO: change this to something better
	//
	nextUuid := time.Now().UnixNano()
	nextUuidAsStr := strconv.FormatInt(nextUuid, 10)

	// if an error occurs, pass it back
	if err != nil {
		stdlog(err.Error())
		return -1, err
	}

	// set the job process id to the recovered value
	j.Pid = nextUuid

	// Attempt to marshal the job.
	mresult, err := json.Marshal(j)

	// if an error occurs, pass it back
	if err != nil {
		stdlog(err.Error())
		return -1, err
	}

	// Add task to the general queue
	ctx, cancel = context.WithTimeout(context.Background(),
		etcdGracePeriod*time.Second)

	// attempt to insert it into etcd
	_, err = inst.internal.Put(ctx, path.Join(queue_dir,
		nextUuidAsStr), string(mresult))

	// if debug mode...
	if debugMode {

		// Grab the newly inserted job entry.
		debug_resp, err := inst.internal.Get(ctx, path.Join(queue_dir,
			nextUuidAsStr))

		// if an error occurs, this failed to insert the new job
		if err != nil {
			cancel()
			stdlog(err.Error())
			return -1, err
		}

		// Cycle thru the values of the newly queued job.
		debugf("The newly queued job was as follows:")
		for _, ent := range debug_resp.Kvs {
			debugf(string(ent.Value))
		}
	}

	// cancel the current context as it is no longer needed
	cancel()

	// if an error occurred, pass it back
	if err != nil {
		stdlog(err.Error())
		return -1, err
	}

	// if all was successful, go ahead and pass back the id
	return nextUuid, nil
}

//! Hand the job off to the job list of a node.
/*
 * @param    string    host id
 * @param    Job       given job to add to node
 */
func (inst *EtcdInstance) QueueJobOnNode(hid string, j *Job) error {

	// input validation
	if j == nil {
		return errorf("QueueJobOnNode() --> invalid input")
	}

	// define a node husk; it'll get used later on when this hands the job off to it
	var node Node

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		etcdGracePeriod*time.Second)

	// Get a response from the primed node.
	response, err := inst.internal.Get(ctx, path.Join(nodes_dir, hid),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	// if debug mode...
	if debugMode && err == nil {

		// cycle thru all of the current nodes for the benefit of the developer
		debugf("The primed node contains the following:")
		for _, ent := range response.Kvs {
			debugf(string(ent.Key) + " => " + string(ent.Value))
		}
	}

	// cancel the current context as it is no longer needed
	cancel()

	// if an error occurs, pass it back
	if err != nil {
		stdlog("QueueJobOnNode() --> unable to get response from primed node")
		return err
	}

	// further safety check, ensure the primed node is not null
	if len(response.Kvs) < 1 {
		stdlog("QueueJobOnNode() --> primed node appears to be null")
		return errorf("QueueJobOnNode() --> primed node appears to be null")
	}

	// unmarshal the recovered response data
	err = json.Unmarshal(response.Kvs[0].Value, &node)

	// if an error occurs, pass it back
	if err != nil {
		return err
	}

	// Job ID is the next index in the array.
	uuid := strconv.FormatInt(j.Pid, 10)

	// Marshal the job data.
	mresult, err := json.Marshal(j)

	// if an error occurred, pass it back
	if err != nil {
		return err
	}

	// Since the primed node gave back a valid response, this needs to
	// place the job on the queue of the node; so firstly assemble the
	// queue path on etcd as per the following:
	//
	// key => /nodes/jobs/uuid
	//
	// Developer note: the queued job of uuid is waiting, and its details
	// can be found in the global 'processesList'.
	//
	jobQueue := path.Join(nodes_dir, hid, "jobs")

	// grab the current context, need it to hand off the job to the node
	ctx, cancel = context.WithTimeout(context.Background(),
		etcdGracePeriod*time.Second)

	// Insert the job into the job queue of the node.
	responseToJobAddToNode, err := inst.internal.Put(ctx,
		path.Join(jobQueue, uuid), string(mresult))

	// if debug mode...
	if debugMode && err == nil && responseToJobAddToNode != nil {

		// tell the developer what happened at this stage...
		debugf("Node manager has responsed with valid put response.")
	}

	// cancel the current context as this no longer needs it
	cancel()

	// if an error occurs, pass it back
	if err != nil {
		return err
	}

	// update etcd with the new node queue info
	err = inst.putNode(&node)

	// if an error occurred, pass it back
	if err != nil {
		stdlog(err.Error())
		return err
	}

	// print out a helpful message about where the new job was queued
	stdlog("A new job was added to a node on the following host: " +
		node.HostName)
	debugf("Primed node uuid was: " + node.HostID)

	// everything was a success, so return nil here
	return nil
}

//! Grab the node of a given host.
/*
 * @param    string    host id
 *
 * @return   Node*     pointer to a given node
 * @return   error     error message, if any
 */
func (inst *EtcdInstance) getNode(hid string) (*Node, error) {

	// input validation
	if len(hid) < 1 {
		return nil, errorf("getNode() --> invalid input")
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		etcdGracePeriod*time.Second)

	// Use the host id to determine which node.
	response, err := inst.internal.Get(ctx, path.Join(nodes_dir, hid),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	// cancel the current context
	cancel()

	// if an error occurs, pass it back
	if err != nil {
		return nil, err
	}

	// safety check, ensure the value is something proper
	if len(response.Kvs) < 1 {
		return nil, errorf("No node with a Host id " + hid +
			" detected in the store")
	}

	// Since the value appears safe, attempt to unmarshal.
	node := &Node{}
	err = json.Unmarshal(response.Kvs[0].Value, node)

	// if an error occurs, pass it back
	if err != nil {
		return nil, err
	}

	// if the program made it here, pass back the node pointer
	return node, nil
}

//! Insert the node ref into etcd.
/*
 * @param    Node     pointer to given node
 *
 * @param    error    error message, if any
 */
func (inst *EtcdInstance) putNode(node *Node) error {

	// input validation, make sure this actually got a node
	if node == nil {
		return errorf("putNode() --> invalid input")
	}

	// Marshal the node data.
	mresult, err := json.Marshal(node)

	// if an error occurs, print it out
	if err != nil {
		return err
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		etcdGracePeriod*time.Second)

	// attempt to insert the value into etcd
	_, err = inst.internal.Put(ctx, path.Join(nodes_dir, node.HostID),
		string(mresult))

	// cancel the current context
	cancel()

	// if an error occurred, pass it back
	if err != nil {
		return err
	}

	// otherwise this was successful so pass back nil
	return nil
}

//! Grab the entire list of every node in the cluster.
/*
 * @return    Node*[]    array of node pointers
 * @return    error      error message
 */
func (inst *EtcdInstance) obtainListOfNodes() (nodes []*Node, err error) {

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		etcdGracePeriod*time.Second)

	// Obtain the nodes dir contents.
	response, err := inst.internal.Get(ctx, nodes_dir, clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	// cancel the current context
	cancel()

	// if an error occurs, pass it back
	if err != nil {
		return nil, err
	}

	// For every key entry in the nodes location.
	debugf("The following nodes are present on the list:")
	for _, entry := range response.Kvs {

		// Parse the entry into pieces, useful for checking the
                // hostname.
		pieces := strings.Split(string(entry.Key), "/")

		// node data are stored in the form...
		//
		//  /nodes/hostname/
		//
		// so a couple of quick safety checks might be warranted.
		//
		if len(pieces) != 3 || len(pieces[2]) < 1 {
			continue
		}

		// Attempt to unmarshal the node data.
		node := Node{}
		err = json.Unmarshal(entry.Value, &node)

		// if an error occurs, pass it back; also a partial list of nodes
		// could be useful as a fallback
		if err != nil {
			return nodes, err
		}

		// having obtained the data correctly, append it to the list of
		// nodes array that eventually gets passed back
		nodes = append(nodes, &node)
		debugf(string(entry.Key) + " => hostname: " + node.HostName +
			", host id: " + node.HostID)
	}

	// pass back the completed list of nodes
	return nodes, nil
}
