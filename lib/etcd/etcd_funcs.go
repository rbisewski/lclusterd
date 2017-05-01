/*
 * File: etcd_funcs.go
 *
 * Description: holds the etcd instance structure and funcs
 */

package libetcd

import (
	"../../lcfg"
	pb "../../lclusterpb"
	"encoding/json"
	"fmt"
	clientv3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

// Jobs are merely gRPC obj refs.
type Job pb.StartJobRequest

// Etcd instance structure, which holds a pointer to the client and nodes.
type EtcdInstance struct {

	// Client pointer to the server itself.
	Client *clientv3.Client

	// Rootfs path, as string
	rootfs string

	// Pointer to nodes.
	node *Node

	scheduler *Scheduler

	// List of current processes.
	ProcessesList map[int64]*Process
}

// The node definition.
type Node struct {
	HostName string
	HostID   string
}

//! Function to start etcd in the background.
/*
 * @param    string   namespace
 *
 * @return   error    error message, if any
 */
func StartEtcdServerBackgroundProcess(namespace string) error {

	// input validation, ensure that the global network namespace value
	// gets set to something safe; note that this value is set in the main
	// routine of main.go
	if len(namespace) < 1 {
		return fmt.Errorf("Error: Improper network namespace length!")
	}

	// Current protocol being used.
	protocol := "http://"

	// Assemble a command based on the client / server ports and given
	// global etcd server address.
	//
	// Note: this appends a 32 digit number to etcd data dir ensure the
	// given etcd session is unique.
	initialAdvertisePeerUrls := protocol + namespace + lcfg.EtcdServerPort
	listenPeerUrls := protocol + namespace + lcfg.EtcdServerPort
	listenClientUrls := protocol + namespace + lcfg.EtcdClientPort
	advertiseClientUrls := protocol + namespace + lcfg.EtcdClientPort
	dataDirWithUniqueCryptonum := lcfg.EtcdDataDir + spawnUuid(32)

	// Create a string array to hold all of the necessary arguments.
	var etcdArgs = []string{
		"--name",
		namespace,
		"--initial-advertise-peer-urls",
		initialAdvertisePeerUrls,
		"--listen-peer-urls",
		listenPeerUrls,
		"--listen-client-urls",
		listenClientUrls,
		"--advertise-client-urls",
		advertiseClientUrls,
		"--data-dir",
		dataDirWithUniqueCryptonum,
	}

	// Attempt to exec the command.
	err := exec.Command(lcfg.EtcdBinaryPath, etcdArgs...).Start()

	// Pass back the error or nil.
	return err
}

//! Creates a new EtcdInstance and returns a pointer to it.
/*
 * @param   string          network namespace
 * @param   string          rootfs path
 *
 * @return  EtcdInstance    pointer to the obj
 *          error           message of the error, nil if none
 */
func CreateEtcdInstance(namespace string, rootfs string) (inst *EtcdInstance,
	err error) {

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
		return nil, fmt.Errorf("CreateEtcdInstance() --> improperly " +
			"generated client due to...\n" +
			err.Error())
	}

	// grab the hostname, if an error occurs pass it back
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
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
		HostName: hostname,
		HostID:   "n_" + spawnUuid(16),
	}

	// Assign a scheduler object.
	newScheduler := &Scheduler{}

	// Create an etcdInstance using the new Client client.
	inst = &EtcdInstance{Client: newlyGeneratedClient, rootfs: rootfs,
		node: node, scheduler: newScheduler}

	// Return the completed etcdInstance.
	return inst, nil
}

//! Update the TTL for a given reserved entry.
/*
 * @param    LeaseID    id of a given etcd key entry
 *
 * @return   error      error message, if any
 */
func (inst *EtcdInstance) updateTTL(leaseID clientv3.LeaseID) error {

	// Keep the entry alive.
	_, err := inst.Client.KeepAliveOnce(context.TODO(), leaseID)

	// return the error, if any
	return err
}

//! Function to keep etcd key values alive.
/*
 * @param    LeaseGrantResponse    lease response
 */
func (inst *EtcdInstance) keepKeyAlive(lease *clientv3.LeaseGrantResponse, successfullyUpdatedTTL *bool) {

	for {
		// Update the time-to-live
		err := inst.updateTTL(lease.ID)

		// If unable to update the TTL, pass back the error message.
		if err != nil {
			*successfullyUpdatedTTL = false
			break
		}

		// Then go back to sleep.
		time.Sleep(time.Duration(lease.TTL / 2))
		*successfullyUpdatedTTL = true
	}
}

//! Prime a given node to make it ready for jobs.
/*
 * @param    chan              channel for listening who is primed
 * @param    context.Context   current context
 *
 * @return   none
 */
func (inst *EtcdInstance) primedLock(primedNotificationChan chan bool,
	ctx context.Context) error {

	// Grant the lease.
	lease, err := inst.Client.Grant(ctx, lcfg.PrimedTTL)

	// if an error occurs, pass it back
	if err != nil {
		return err
	}

	// Setup a key-value store.
	kvc := clientv3.NewKV(inst.Client)

	// Check if the value exists, and insert it if it does not.
	response, err := kvc.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(lcfg.Primed), "=", "primed")).
		Then().
		Else(clientv3.OpPut(lcfg.Primed, "primed", clientv3.WithLease(lease.ID))).
		Commit()

	// If an error occurred, pass it back.
	if err != nil {
		return err
	}

	// If no other node response, this probably means that we are primed
	// node, so keep the lease entry alive.
	if !response.Succeeded {

		// Go ahead and attempt to update the TTL of the key.
		successfullyUpdatedTTL := true
		go inst.keepKeyAlive(lease, &successfullyUpdatedTTL)

		// As long as this can continue to properly update the TTL,
		// then the channel can stay open.
		primedNotificationChan <- successfullyUpdatedTTL
	}

	// if a node *did* respond, then this node is not the primed node; go
	// ahead and clean up the lease ref whilst waiting for a turn...
	if response.Succeeded {
		inst.Client.Revoke(ctx, lease.ID)
		go inst.watchUntilPrimed(primedNotificationChan)
	}

	// if the node got this far, close the notification channel
	primedNotificationChan <- false

	// success
	return nil
}

//! Listens on the notification channel until the current primed node
//! has completed it's job.
/*
 * @param    chan    notification channel
 *
 * @return   none
 */
func (inst *EtcdInstance) watchUntilPrimed(notificationChan chan bool) {

	// Check for a channel response.
	responsingChan := inst.Client.Watch(context.Background(), lcfg.Primed)

	// keep running until this node gets to be primed
	for {

		// Grab the response from the channel.
		wresponse := <-responsingChan

		// Cycle through the events.
		for _, ev := range wresponse.Events {

			// if not a deleted prime event, skip it...
			if ev.Type != mvccpb.DELETE {
				continue
			}

			// Grab the context, as it is required to manage the nodes.
			ctx, cancel := context.WithTimeout(context.Background(),
				lcfg.PrimedTTL*time.Second)

			// Attempt to prime the node.
			err := inst.primedLock(notificationChan, ctx)

			// Cancel the current context.
			cancel()

			// if an error occurs, print it our
			if err != nil {
				log.Println(err.Error() + "\n")
			}

			// end the loop since this is completed
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
			log.Println("primeThisNode() --> node gathering failed\n " +
				err.Error())
			continue
		}

		// Generate the job queue
		err = inst.initializeJobQueue()

		// if an error occurs, print it
		if err != nil {
			log.Println("primeThisNode() --> unable to create job queue\n" +
				err.Error())
			continue
		}

		// reserve a global id for jobs
		err = inst.initializeGlobalJobID()

		// if an error occurs, print it
		if err != nil {
			log.Println("primeThisNode() --> unable to set global id\n" +
				err.Error())
			continue
		}

		// setup the directory for processes to be ran on nodes
		err = inst.initializeProcessStorage()

		// if an error occurred, print it out
		if err != nil {
			log.Println("primeThisNode() --> unable to set aside storage for " +
				"processes\n" + err.Error())
			continue
		}

		// with all of the pieces in place, this syncs with the scheduler,
		// and the entire setup is primed for new jobs
		err = inst.syncScheduler(inst.node.HostID, nodes)

		// if an error occurs, print it out and skip to the next loop
		if err != nil {
			log.Println(err.Error())
			continue
		}

		// else watch the job queue
		go inst.watchGeneralJobQueue()
	}
}

//! Setup a job queue so that the end-user can add jobs to it.
/*
 * @return   error   error message, if any
 */
func (inst *EtcdInstance) initializeJobQueue() error {

	// Setup a key-value store for the queue elements.
	kvc := clientv3.NewKV(inst.Client)

	// Grab the context.
	ctx, cancel := context.WithTimeout(context.Background(),
		lcfg.EtcdGracePeriodSec*time.Second)

	// Ensure the queue dirs exist, and if not, create them.
	_, err := kvc.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(lcfg.QueueDir), "=", "queue")).
		Then().
		Else(clientv3.OpPut(lcfg.QueueDir, "queue")).
		Commit()

	// cancel the context since it is no longer needed
	cancel()

	// pass back the error, if any
	return err
}

//! Setup a global id for potential jobs.
/*
 * @return    error    error message, if any
 */
func (inst *EtcdInstance) initializeGlobalJobID() error {

	// Setup a key value storage.
	kvc := clientv3.NewKV(inst.Client)

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		lcfg.EtcdGracePeriodSec*time.Second)

	// Check if a jobs dir exists, else create one.
	_, err := kvc.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(lcfg.JobsDir), "=", "0")).
		Then().
		Else(clientv3.OpPut(lcfg.JobsDir, "0")).
		Commit()

	// cancel the current context if we no longer need it
	cancel()

	// pass back the error, if any
	return err
}

//! Setup the processes storage.
/*
 * @param    error
 */
func (inst *EtcdInstance) initializeProcessStorage() error {

	// setup the key-value store
	kvc := clientv3.NewKV(inst.Client)

	// grab the current context
	ctx, cancel := context.WithTimeout(context.Background(),
		lcfg.EtcdGracePeriodSec*time.Second)

	// check if the processes location exists, and if not, creates it
	_, err := kvc.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(lcfg.ProcessesDir), "=", "processes")).
		Then().
		Else(clientv3.OpPut(lcfg.ProcessesDir, "processes")).
		Commit()

	// cancel the current context as this no longer needs it
	cancel()

	// pass back the error, if any
	return err
}

//! Initialize the nodes.
/*
 * @return    chan    newly initialized node comm channel
 * @return    error   error message, if any
 */
func (inst *EtcdInstance) InitNode() (chan bool, error) {

	// Assign a chunk of memory for any potential processes
	inst.ProcessesList = make(map[int64]*Process)

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

	// Grab the context, as it is required to manage the nodes.
	ctx, cancel := context.WithTimeout(context.Background(),
		lcfg.PrimedTTL*time.Second)

	// Attempt to prime the node.
	err = inst.primedLock(notify, ctx)

	// Cancel the current context.
	cancel()

	// If any error, pass it back.
	if err != nil {
		return nil, err
	}

	// Otherwise hand a ref to the notification channel.
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
		log.Println(err.Error())
		return err
	}

	// Grant a lease.
	lease, err := inst.Client.Grant(context.TODO(), lcfg.NlistTTL)

	// if an error occurs, pass it back
	if err != nil {
		return err
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		lcfg.EtcdGracePeriodSec*time.Second)

	// insert the key into etcd
	_, err = inst.Client.Put(ctx, path.Join(lcfg.NodesDir, inst.node.HostID),
		string(mresult), clientv3.WithLease(lease.ID))

	// cancel the current context as this no longer needs it
	cancel()

	// revoke the lease if an error occurred
	if err != nil {
		inst.Client.Revoke(context.TODO(), lease.ID)
		return err
	}

	// Setup a local jobs queue for the node.
	jobQueue := path.Join(lcfg.NodesDir, inst.node.HostID, "jobs")

	// grab the current context
	ctx, cancel = context.WithTimeout(context.Background(),
		lcfg.EtcdGracePeriodSec*time.Second)

	// add the context into the k-v storage as well
	_, err = inst.Client.Put(ctx, jobQueue, "")

	// cancel the current context as it is no long needed
	cancel()

	// if an error occurs, pass it back
	if err != nil {
		return err
	}

	// keep the key entry alive
	successfullyUpdatedTTL := true
	go inst.keepKeyAlive(lease, &successfullyUpdatedTTL)

	// keep tabs on the job queue
	go inst.watchClientJobQueue()

	// everything works, so pass back nil
	return nil
}

//! Append a process to the list of processes.
/*
 * @param    Process    requested process
 *
 * @result   error      error message, if any
 */
func (inst *EtcdInstance) storeProcess(p Process) error {

	// utilize the Uuid as the hash key; add the process to the list of
	// processes
	inst.ProcessesList[p.Uuid] = &p

	// Attempt to marshell a node.
	mresult, err := json.Marshal(inst.node)

	// if an error occurs, pass it back
	if err != nil {
		return err
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		lcfg.EtcdGracePeriodSec*time.Second)

	// insert a key-value into etcd
	_, err = inst.Client.Put(ctx, path.Join(lcfg.ProcessesDir,
		strconv.FormatInt(p.Uuid, 10)), string(mresult))

	// cancel the current context, as it is no longer needed
	cancel()

	// pass back the resulting error, if any
	return err
}

//! Grab the process from the global list of processes.
/*
 * @param    int64    process Uuid
 *
 * @return   Process  the process information
 * @return   error    error message, if any
 */
func (inst *EtcdInstance) ObtainProcess(Uuid int64) (p *Process, err error) {

	// ensure the Uuid is something reasonable
	if Uuid < 0 {
		return nil, fmt.Errorf("obtainProcess() --> invalid input")
	}

	// otherwise return the process in question
	return inst.ProcessesList[Uuid], nil
}

//! Allows the node to watch Client queue, in case there are more jobs.
/*
 * @return    none
 */
func (inst *EtcdInstance) watchClientJobQueue() {

	// Assemble the job queue path.
	jobQueue := path.Join(lcfg.NodesDir, inst.node.HostID, "jobs")

	// Watch the channel for a response.
	rchan := inst.Client.Watch(context.Background(), jobQueue,
		clientv3.WithPrefix())

	for {

		// Grab responses from the channel.
		wresponse := <-rchan

		// Cycle through all given events.
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
				log.Println("watchClientJobQueue() --> failed to read job data")
				log.Println(err.Error())
				return
			}

			// Using the job data, attempt to start a process
			// on the node.
			p, err := inst.StartProcess(j)

			// print out the error message, if any
			if err != nil {
				log.Println("watchClientJobQueue() --> unable to start process")
				log.Println(err.Error())
				return
			}

			// store a ref to the process into etcd
			err = inst.storeProcess(p)

			// print out the error message, if any
			if err != nil {
				log.Println("watchClientJobQueue() --> could not store process " +
					"data into etcd")
				log.Println(err.Error())
				return
			}

			// otherwise the job started successful, so go ahead and print
			// out a helpful message
			log.Println("New job started (" + j.Path + ") in a node on cluster " +
				"host: " + inst.node.HostName)
			debugf("Primed node that received job has Uuid: " +
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
	rch := inst.Client.Watch(context.Background(), lcfg.QueueDir,
		clientv3.WithPrefix())

	for {

		// Look for responses in the channel.
		wresponse := <-rch

		// Cycle through all of the events.
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
				log.Println("watchGeneralJobQueue() --> unable to read job data")
				log.Println(err.Error())
			}

			// attempt to schedule the job
			err = inst.scheduleJob(&j)

			// if an error occurred, print it out
			if err != nil {
				log.Println("watchGeneralJobQueue() --> unable to schedule job")
				log.Println(err.Error())
			}

			// print out a helpful message
			log.Println("Node manager has detected a scheduled job.")
		}
	}
}

//! Add a job to the scheduler queue.
/*
 * @param    Job      given job
 *
 * @return   int64    Uuid
 * @return   error    error message, if any
 */
func (inst *EtcdInstance) AddToGlobalQueue(j Job) (int64, error) {

	// further check, ensure the instance is safe
	if inst.Client == nil {
		return -1, fmt.Errorf("addToGlobalQueue() --> malformed etcd instance")
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		lcfg.EtcdGracePeriodSec*time.Second)

	// Grab the list of queued jobs.
	response, err := inst.Client.Get(ctx, lcfg.QueueDir)

	// Debug; if there are no jobs, do this.
	if lcfg.DebugMode && len(response.Kvs) < 1 {
		debugf("No jobs are currently in the queue.")

		// Debug, if there is at least 1 job, print it out.
	} else if lcfg.DebugMode {
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
		log.Println(err.Error())
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
		log.Println(err.Error())
		return -1, err
	}

	// set the job process id to the recovered value
	j.Pid = nextUuid

	// Attempt to marshal the job.
	mresult, err := json.Marshal(j)

	// if an error occurs, pass it back
	if err != nil {
		log.Println(err.Error())
		return -1, err
	}

	// Add task to the general queue
	ctx, cancel = context.WithTimeout(context.Background(),
		lcfg.EtcdGracePeriodSec*time.Second)

	// attempt to insert it into etcd
	_, err = inst.Client.Put(ctx, path.Join(lcfg.QueueDir,
		nextUuidAsStr), string(mresult))

	// if debug mode...
	if lcfg.DebugMode {

		// Grab the newly inserted job entry.
		debugResponse, err := inst.Client.Get(ctx,
			path.Join(lcfg.QueueDir, nextUuidAsStr))

		// if an error occurs, this failed to insert the new job
		if err != nil {
			cancel()
			log.Println(err.Error())
			return -1, err
		}

		// Cycle through the values of the newly queued job.
		debugf("The newly queued job was as follows:")
		for _, ent := range debugResponse.Kvs {
			debugf(string(ent.Value))
		}
	}

	// cancel the current context as it is no longer needed
	cancel()

	// if an error occurred, pass it back
	if err != nil {
		log.Println(err.Error())
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
func (inst *EtcdInstance) QueueJobOnNode(hostID string, j *Job) error {

	// input validation
	if j == nil {
		return fmt.Errorf("QueueJobOnNode() --> invalid input")
	}

	// define a node husk; it'll get used later on when this hands the job off to it
	var node Node

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		lcfg.EtcdGracePeriodSec*time.Second)

	// Get a response from the primed node.
	response, err := inst.Client.Get(ctx, path.Join(lcfg.NodesDir, hostID),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	// if debug mode...
	if lcfg.DebugMode && err == nil {

		// cycle through all of the current nodes for the benefit of the developer
		debugf("The primed node contains the following:")
		for _, ent := range response.Kvs {
			debugf(string(ent.Key) + " => " + string(ent.Value))
		}
	}

	// cancel the current context as it is no longer needed
	cancel()

	// if an error occurs, pass it back
	if err != nil {
		return err
	}

	// further safety check, ensure the primed node is not null
	if len(response.Kvs) < 1 {
		return fmt.Errorf("QueueJobOnNode() --> primed node appears to be null")
	}

	// unmarshal the recovered response data
	err = json.Unmarshal(response.Kvs[0].Value, &node)

	// if an error occurs, pass it back
	if err != nil {
		return err
	}

	// Job ID is the next index in the array.
	Uuid := strconv.FormatInt(j.Pid, 10)

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
	// key => /nodes/jobs/Uuid
	//
	// Developer note: the queued job of Uuid is waiting, and its details
	// can be found in the global 'ProcessesList'.
	//
	jobQueue := path.Join(lcfg.NodesDir, hostID, "jobs")

	// grab the current context, need it to hand off the job to the node
	ctx, cancel = context.WithTimeout(context.Background(),
		lcfg.EtcdGracePeriodSec*time.Second)

	// Insert the job into the job queue of the node.
	responseToJobAddToNode, err := inst.Client.Put(ctx,
		path.Join(jobQueue, Uuid), string(mresult))

	// if debug mode...
	if lcfg.DebugMode && err == nil && responseToJobAddToNode != nil {

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
		log.Println(err.Error())
		return err
	}

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
func (inst *EtcdInstance) getNode(hostID string) (*Node, error) {

	// input validation
	if len(hostID) < 1 {
		return nil, fmt.Errorf("getNode() --> invalid input")
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		lcfg.EtcdGracePeriodSec*time.Second)

	// Use the host id to determine which node.
	response, err := inst.Client.Get(ctx, path.Join(lcfg.NodesDir, hostID),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	// cancel the current context
	cancel()

	// if an error occurs, pass it back
	if err != nil {
		return nil, err
	}

	// safety check, ensure the value is something proper
	if len(response.Kvs) < 1 {
		return nil, fmt.Errorf("No node with a Host id " + hostID +
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
		return fmt.Errorf("putNode() --> invalid input")
	}

	// Marshal the node data.
	mresult, err := json.Marshal(node)

	// if an error occurs, print it out
	if err != nil {
		return err
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		lcfg.EtcdGracePeriodSec*time.Second)

	// attempt to insert the value into etcd
	_, err = inst.Client.Put(ctx, path.Join(lcfg.NodesDir,
		node.HostID), string(mresult))

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
		lcfg.EtcdGracePeriodSec*time.Second)

	// Obtain the nodes dir contents.
	response, err := inst.Client.Get(ctx, lcfg.NodesDir,
		clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey,
			clientv3.SortAscend))

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
