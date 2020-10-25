package libetcd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"../config"
	pb "../lclusterpb"
	clientv3 "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
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

	if len(namespace) < 1 {
		return fmt.Errorf("Error: Improper network namespace length!")
	}

	protocol := "http://"

	// Assemble a command based on the client / server ports and given
	// global etcd server address.
	//
	// Note: this appends a 32 digit number to etcd data dir ensure the
	// given etcd session is unique.
	initialAdvertisePeerUrls := protocol + namespace + config.EtcdServerPort
	listenPeerUrls := protocol + namespace + config.EtcdServerPort
	listenClientUrls := protocol + namespace + config.EtcdClientPort
	advertiseClientUrls := protocol + namespace + config.EtcdClientPort
	dataDirWithUniqueCryptonum := config.EtcdDataDir + spawnUuid(32)

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

	return exec.Command(config.EtcdBinaryPath, etcdArgs...).Start()
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

	etcdClientConfiguration := clientv3.Config{
		Endpoints:   []string{namespace + config.EtcdClientPort},
		DialTimeout: 5 * time.Second,
	}

	newlyGeneratedClient, err := clientv3.New(etcdClientConfiguration)
	if err != nil {
		return nil, fmt.Errorf("CreateEtcdInstance() --> improperly " +
			"generated client due to...\n" +
			err.Error())
	}

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

	lease, err := inst.Client.Grant(ctx, config.PrimedTTL)
	if err != nil {
		return err
	}

	// Setup a key-value store.
	kvc := clientv3.NewKV(inst.Client)

	response, err := kvc.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(config.Primed), "=", "primed")).
		Then().
		Else(clientv3.OpPut(config.Primed, "primed", clientv3.WithLease(lease.ID))).
		Commit()

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
	responsingChan := inst.Client.Watch(context.Background(), config.Primed)

	// keep running until this node gets to be primed
	for {
		wresponse := <-responsingChan

		for _, ev := range wresponse.Events {

			// if not a deleted prime event, skip it...
			if ev.Type != mvccpb.DELETE {
				continue
			}

			// Grab the context, as it is required to manage the nodes.
			ctx, cancel := context.WithTimeout(context.Background(),
				config.PrimedTTL*time.Second)

			// Attempt to prime the node.
			err := inst.primedLock(notificationChan, ctx)

			// Cancel the current context.
			cancel()

			if err != nil {
				log.Println(err.Error() + "\n")
			}

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
		isPrimed := <-notify

		// not yet primed? do nothing...
		if !isPrimed {
			continue
		}

		nodes, err := inst.obtainListOfNodes()
		if err != nil {
			log.Println("primeThisNode() --> node gathering failed\n " +
				err.Error())
			continue
		}

		err = inst.initializeJobQueue()
		if err != nil {
			log.Println("primeThisNode() --> unable to create job queue\n" +
				err.Error())
			continue
		}

		err = inst.initializeGlobalJobID()
		if err != nil {
			log.Println("primeThisNode() --> unable to set global id\n" +
				err.Error())
			continue
		}

		err = inst.initializeProcessStorage()
		if err != nil {
			log.Println("primeThisNode() --> unable to set aside storage for " +
				"processes\n" + err.Error())
			continue
		}

		// with all of the pieces in place, this syncs with the scheduler,
		// and the entire setup is primed for new jobs
		err = inst.syncScheduler(inst.node.HostID, nodes)
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

	ctx, cancel := context.WithTimeout(context.Background(),
		config.EtcdGracePeriodSec*time.Second)

	// Ensure the queue dirs exist, and if not, create them.
	_, err := kvc.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(config.QueueDir), "=", "queue")).
		Then().
		Else(clientv3.OpPut(config.QueueDir, "queue")).
		Commit()

	cancel()
	return err
}

//! Setup a global id for potential jobs.
/*
 * @return    error    error message, if any
 */
func (inst *EtcdInstance) initializeGlobalJobID() error {

	// Setup a key value storage.
	kvc := clientv3.NewKV(inst.Client)

	ctx, cancel := context.WithTimeout(context.Background(),
		config.EtcdGracePeriodSec*time.Second)

	// Check if a jobs dir exists, else create one.
	_, err := kvc.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(config.JobsDir), "=", "0")).
		Then().
		Else(clientv3.OpPut(config.JobsDir, "0")).
		Commit()

	cancel()
	return err
}

//! Setup the processes storage.
/*
 * @param    error
 */
func (inst *EtcdInstance) initializeProcessStorage() error {

	// setup the key-value store
	kvc := clientv3.NewKV(inst.Client)

	ctx, cancel := context.WithTimeout(context.Background(),
		config.EtcdGracePeriodSec*time.Second)

	// check if the processes location exists, and if not, creates it
	_, err := kvc.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(config.ProcessesDir), "=", "processes")).
		Then().
		Else(clientv3.OpPut(config.ProcessesDir, "processes")).
		Commit()

	cancel()
	return err
}

//! Initialize the nodes.
/*
 * @return    chan    newly initialized node comm channel
 * @return    error   error message, if any
 */
func (inst *EtcdInstance) InitNode() (chan bool, error) {

	// initialize an instance
	inst.ProcessesList = make(map[int64]*Process)
	err := inst.addToNodesList()
	if err != nil {
		return nil, err
	}

	// wait around until this node can be primed
	notify := make(chan bool, 1)
	go inst.primeThisNode(notify)

	// Grab the context, as it is required to manage the nodes.
	ctx, cancel := context.WithTimeout(context.Background(), config.PrimedTTL*time.Second)

	// Attempt to prime the node.
	err = inst.primedLock(notify, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	return notify, nil
}

//! Setup this node and attach it to the list.
/*
 * @return    error   error message, if any
 */
func (inst *EtcdInstance) addToNodesList() error {

	mresult, err := json.Marshal(inst.node)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		config.EtcdGracePeriodSec*time.Second)

	// Grant a lease.
	lease, err := inst.Client.Grant(ctx, config.NlistTTL)
	if err != nil {
		return err
	}

	// insert the key into etcd
	_, err = inst.Client.Put(ctx, path.Join(config.NodesDir, inst.node.HostID),
		string(mresult), clientv3.WithLease(lease.ID))
	if err != nil {
		inst.Client.Revoke(ctx, lease.ID)
		return err
	}

	cancel()

	jobQueue := path.Join(config.NodesDir, inst.node.HostID, "jobs")

	// grab the current context
	ctx, cancel = context.WithTimeout(context.Background(),
		config.EtcdGracePeriodSec*time.Second)

	_, err = inst.Client.Put(ctx, jobQueue, "")
	cancel()
	if err != nil {
		return err
	}

	// keep the key entry alive
	successfullyUpdatedTTL := true
	go inst.keepKeyAlive(lease, &successfullyUpdatedTTL)

	// keep tabs on the job queue
	go inst.watchClientJobQueue()

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

	mresult, err := json.Marshal(inst.node)
	if err != nil {
		return err
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		config.EtcdGracePeriodSec*time.Second)

	// insert a key-value into etcd
	_, err = inst.Client.Put(ctx, path.Join(config.ProcessesDir,
		strconv.FormatInt(p.Uuid, 10)), string(mresult))

	cancel()
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

	if Uuid < 0 {
		return nil, fmt.Errorf("obtainProcess() --> invalid input")
	}

	return inst.ProcessesList[Uuid], nil
}

//! Allows the node to watch Client queue, in case there are more jobs.
/*
 * @return    none
 */
func (inst *EtcdInstance) watchClientJobQueue() {

	jobQueue := path.Join(config.NodesDir, inst.node.HostID, "jobs")

	rchan := inst.Client.Watch(context.Background(), jobQueue,
		clientv3.WithPrefix())

	for {
		wresponse := <-rchan

		for _, ev := range wresponse.Events {

			// this function is only interested in put events
			if ev.Type != mvccpb.PUT {
				continue
			}

			// Unmarshal the data, if any.
			var j Job
			err := json.Unmarshal(ev.Kv.Value, &j)
			if err != nil {
				log.Println("watchClientJobQueue() --> failed to read job data")
				log.Println(err.Error())
				return
			}

			// Using the job data, attempt to start a process
			// on the node.
			p, err := inst.StartProcess(j)
			if err != nil {
				log.Println("watchClientJobQueue() --> unable to start process")
				log.Println(err.Error())
				return
			}

			// store a ref to the process into etcd
			err = inst.storeProcess(p)
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

	rch := inst.Client.Watch(context.Background(), config.QueueDir,
		clientv3.WithPrefix())

	for {
		wresponse := <-rch

		for _, ev := range wresponse.Events {

			// this function is only interested in put events
			if ev.Type != mvccpb.PUT {
				continue
			}

			var j Job

			err := json.Unmarshal(ev.Kv.Value, &j)
			if err != nil {
				log.Println("watchGeneralJobQueue() --> unable to read job data")
				log.Println(err.Error())
			}

			err = inst.scheduleJob(&j)
			if err != nil {
				log.Println("watchGeneralJobQueue() --> unable to schedule job")
				log.Println(err.Error())
			}

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

	if inst.Client == nil {
		return -1, fmt.Errorf("addToGlobalQueue() --> malformed etcd instance")
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		config.EtcdGracePeriodSec*time.Second)

	// Grab the list of queued jobs.
	response, err := inst.Client.Get(ctx, config.QueueDir)

	// Debug; if there are no jobs, do this.
	if config.DebugMode && len(response.Kvs) < 1 {
		debugf("No jobs are currently in the queue.")

		// Debug, if there is at least 1 job, print it out.
	} else if config.DebugMode {
		debugf("Current queued jobs are as follows:")
		for i, ent := range response.Kvs {
			debugf(strconv.Itoa(i+1) + ") " + string(ent.Value))
		}
	}

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
	if err != nil {
		log.Println(err.Error())
		return -1, err
	}

	// set the job process id to the recovered value
	j.Pid = nextUuid
	mresult, err := json.Marshal(j)
	if err != nil {
		log.Println(err.Error())
		return -1, err
	}

	// Add task to the general queue
	ctx, cancel = context.WithTimeout(context.Background(),
		config.EtcdGracePeriodSec*time.Second)

	_, err = inst.Client.Put(ctx, path.Join(config.QueueDir,
		nextUuidAsStr), string(mresult))

	if config.DebugMode {

		debugResponse, err := inst.Client.Get(ctx,
			path.Join(config.QueueDir, nextUuidAsStr))

		if err != nil {
			cancel()
			log.Println(err.Error())
			return -1, err
		}

		debugf("The newly queued job was as follows:")
		for _, ent := range debugResponse.Kvs {
			debugf(string(ent.Value))
		}
	}

	cancel()
	if err != nil {
		log.Println(err.Error())
		return -1, err
	}

	return nextUuid, nil
}

//! Hand the job off to the job list of a node.
/*
 * @param    string    host id
 * @param    Job       given job to add to node
 */
func (inst *EtcdInstance) QueueJobOnNode(hostID string, j *Job) error {

	if j == nil {
		return fmt.Errorf("QueueJobOnNode() --> invalid input")
	}

	var node Node

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		config.EtcdGracePeriodSec*time.Second)

	// Get a response from the primed node.
	response, err := inst.Client.Get(ctx, path.Join(config.NodesDir, hostID),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	if config.DebugMode && err == nil {

		debugf("The primed node contains the following:")
		for _, ent := range response.Kvs {
			debugf(string(ent.Key) + " => " + string(ent.Value))
		}
	}

	cancel()
	if err != nil {
		return err
	}
	if len(response.Kvs) < 1 {
		return fmt.Errorf("QueueJobOnNode() --> primed node appears to be null")
	}

	err = json.Unmarshal(response.Kvs[0].Value, &node)
	if err != nil {
		return err
	}

	// Job ID is the next index in the array.
	Uuid := strconv.FormatInt(j.Pid, 10)
	mresult, err := json.Marshal(j)
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
	jobQueue := path.Join(config.NodesDir, hostID, "jobs")

	// grab the current context, need it to hand off the job to the node
	ctx, cancel = context.WithTimeout(context.Background(),
		config.EtcdGracePeriodSec*time.Second)

	// Insert the job into the job queue of the node.
	responseToJobAddToNode, err := inst.Client.Put(ctx,
		path.Join(jobQueue, Uuid), string(mresult))

	if config.DebugMode && err == nil && responseToJobAddToNode != nil {
		debugf("Node manager has responsed with valid put response.")
	}

	cancel()
	if err != nil {
		return err
	}

	err = inst.putNode(&node)
	if err != nil {
		log.Println(err.Error())
		return err
	}

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

	if len(hostID) < 1 {
		return nil, fmt.Errorf("getNode() --> invalid input")
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		config.EtcdGracePeriodSec*time.Second)

	// Use the host id to determine which node.
	response, err := inst.Client.Get(ctx, path.Join(config.NodesDir, hostID),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	cancel()
	if err != nil {
		return nil, err
	}

	if len(response.Kvs) < 1 {
		return nil, fmt.Errorf("No node with a Host id " + hostID +
			" detected in the store")
	}

	node := &Node{}
	err = json.Unmarshal(response.Kvs[0].Value, node)
	if err != nil {
		return nil, err
	}

	return node, nil
}

//! Insert the node ref into etcd.
/*
 * @param    Node     pointer to given node
 *
 * @param    error    error message, if any
 */
func (inst *EtcdInstance) putNode(node *Node) error {

	if node == nil {
		return fmt.Errorf("putNode() --> invalid input")
	}

	mresult, err := json.Marshal(node)
	if err != nil {
		return err
	}

	// Grab the current context.
	ctx, cancel := context.WithTimeout(context.Background(),
		config.EtcdGracePeriodSec*time.Second)

	// attempt to insert the value into etcd
	_, err = inst.Client.Put(ctx, path.Join(config.NodesDir,
		node.HostID), string(mresult))

	cancel()
	if err != nil {
		return err
	}

	return nil
}

//! Grab the entire list of every node in the cluster.
/*
 * @return    Node*[]    array of node pointers
 * @return    error      error message
 */
func (inst *EtcdInstance) obtainListOfNodes() (nodes []*Node, err error) {

	ctx, cancel := context.WithTimeout(context.Background(),
		config.EtcdGracePeriodSec*time.Second)

	response, err := inst.Client.Get(ctx, config.NodesDir,
		clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey,
			clientv3.SortAscend))

	cancel()
	if err != nil {
		return nil, err
	}

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
		if err != nil {
			return nodes, err
		}

		// having obtained the data correctly, append it to the list of
		// nodes array that eventually gets passed back
		nodes = append(nodes, &node)
		debugf(string(entry.Key) + " => hostname: " + node.HostName +
			", host id: " + node.HostID)
	}

	return nodes, nil
}
