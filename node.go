/*
 * File: node.go
 *
 * Description: definition of the Node structure 
 */

package main

import (
    "os"
    "runtime"
    "strconv"
    "syscall"
    "time"

    "github.com/opencontainers/runc/libcontainer"
    "github.com/opencontainers/runc/libcontainer/configs"
    _ "github.com/opencontainers/runc/libcontainer/nsenter"
    pb "./lclusterpb"
)

//
// Node structure
//
type Node struct {

    // String representation of an IPv4 address or Hostname
    Host string

    // Job uuid
    Job_uuid int64

    // Whether or not the node is available for handling jobs
    Locked bool

    // Reference to the container and its process
    Container  libcontainer.BaseContainer
    Proc      *libcontainer.Process
}

//
// NodeManager structure
//
type NodeManager struct {

    // Array of pointers to the currently available
    Nodelist []*Node
}


// ------------------------------
// Node Manager Functions Section
// ------------------------------


//! Grabs the first available, unlocked node from Nodelist.
/*
 * @return     Node*    pointer to the first unlocked node
 */
func (nm *NodeManager) grabNode() *Node {

    // If the node list is empty, send back nil
    if len(nm.Nodelist) < 1 {
        return nil
    }

    // Cycle thru all of the nodes...
    for _, n := range nm.Nodelist {

        // If a node that is unlocked is discovered...
        if n.Locked == false {

            // Return a pointer to this node.
            return n
        }
    }

    // If all of the nodes are locked, send back nil.
    return nil
}


//! Stringifies the Nodelist, useful for listing or debugging, in a format
//! as detailed below:
//!
//! ip addr      | job uuid | is locked?
//! ------------------------------------
//! 192.168.1.77 | 34F77AB2 | no
//! 192.168.1.82 | C4F49A92 | no
//!
/*
 * @return   string    list of nodes as string, in the above format
 */
func (nm *NodeManager) listNodes() string {

    // If the node list is empty, send back a message that the node list
    // appears to contain no nodes...
    if len(nm.Nodelist) < 1 {
        return "There are no nodes currently present in the list..."
    }

    // Variable declaration
    var stringified_list_of_nodes string = ""

    // Append the table titles and separator
    stringified_list_of_nodes += "ip addr      | job uuid | is locked?\n"
    stringified_list_of_nodes += "------------------------------------\n"

    // Cycle thru all of the nodes...
    for _, n := range nm.Nodelist {

        // Assemble the given line...
        line := n.Host + " | " + strconv.FormatInt(n.Job_uuid, 10) + " | "

        // Set string based on whether the node is locked.
        if n.Locked == false {
            line += "no"
        } else {
            line += "yes"
        }

        // Append the line w/ newline char
        stringified_list_of_nodes += line + "\n"
    }

    // Finally return the assembled stringified output.
    return stringified_list_of_nodes
}

//! Returns all of the available nodes from the Nodelist
/*
 * @return    Node[]    array of all unlocked nodes
 */
func (nm *NodeManager) getNodes() []*Node {

    // ensure the node manager actually has nodes
    if len(nm.Nodelist) < 1 {
        stdlog("NOTE: Node manager currently has zero nodes.")
        return nil
    }

    // Make a node array to hold the unlocked nodes
    active_nodes := make([]*Node, 0)

    // Cycle thru all of the nodes...
    for _, n := range nm.Nodelist {

        // Set string based on whether the node is locked.
        if n.Locked == false {
            active_nodes = append(active_nodes, n)
        }
    }

    // if active nodes is zero, send back nil
    if len(active_nodes) < 1 {
        stdlog("NOTE: Node manager has zero unlocked nodes.")
        return nil
    }

    // send back the filled array
    return active_nodes
}


//! Remove a node from the Nodelist
/*
 * @return     error    resultant error message, if any
 */
func (nm *NodeManager) removeNode(n *Node) error {

    // input validation
    if n == nil {
        return errorf("removeNode() --> invalid input")
    }

    // safety check, ensure the node manager *actually* has nodes
    if len(nm.Nodelist) < 1 {
        stdlog("---------------------------------------------------------")
        stdlog("WARNING: Node manager currently has zero nodes, so it is ")
        stdlog("         unable to remove Node w/ Uuid:")
        stdlog("                                                         ")
        stdlog(strconv.FormatInt(n.Job_uuid, 10))
        stdlog("---------------------------------------------------------")
        return errorf("removeNode() --> cannot remove node from empty list")
    }

    // Cycle thru all of the nodes...
    for i, node := range nm.Nodelist {

        // Node w/ requested job was found?
        if node.Job_uuid == n.Job_uuid {

            // mention to the end-user that is node was found and removed
            stdlog("Succesfully removed node w/ uuid of:")
            stdlog(strconv.FormatInt(n.Job_uuid, 10))
            stdlog("                                    ")

            // attempt to remove it from the manager's Nodelist array
            nm.Nodelist = append(nm.Nodelist[:i], nm.Nodelist[i+1:]...)

            // return nil here since everything worked as expected
            return nil
        }
    }

    // Since the node list did not contain the node in question, throw an
    // error mentioning that the node does not exist.
    return errorf("removeNode() --> attempted to remove node that does not exist")
}


// ------------------------------
// Libcontainer Functions Section
// ------------------------------


//! Init basic OS elements of libcontainer
/*
 * @return    none
 *
 * SIDE EFFECT: provides OS exec information to a given container instance
 */
func init() {

    // OS validation, make sure the OS isn't going to start something NULL
    if len(os.Args) <= 1 || len(os.Args[0]) < 1 {
        return
    }

    // further OS validation, ensure the OS is *actually* trying to init
    // something...
    if os.Args[1] != "init" {
        return
    }

    // set the runtime thread limit for golang to 1, for safety
    runtime.GOMAXPROCS(1)

    // place a lock on this thread; the goal being process isolation
    runtime.LockOSThread()

    // spawn an instance of libcontainer for the purpose of containerizing
    containerizer, _ := libcontainer.New("")

    // initialize the instance
    err := containerizer.StartInitialization()

    // if the program gets to this line, then either the process is done or
    // the process has ended prematurely
    if err != nil {
        stdlog(err.Error())
        panic("Libcontainer terminated unexpectedly...")
    }
}

//! Function to start a new process
/*
 * @param     Node    given node
 *
 * @return    error   error message, if any
 */
func (nm *NodeManager) startProcess(sjr pb.StartJobRequest, rootfs string) (*Node, error) {

    // input validation
    if len(rootfs) < 1 {
        return nil, errorf("startProcess() --> unable to start job")
    }

    // containerize the given rootfs location
    containerizer, err := libcontainer.New(rootfs, libcontainer.Cgroupfs)

    // safety check, ensure no error occurred
    if err != nil {
        return nil, errorf("startProcess() --> unable to make new " +
                           "libcontainter instance")
    }

    // generate a container uuid
    containerUuid := time.Now().UnixNano()

    // generate container name
    containerName := "[" + rootfs + "]-" + strconv.FormatInt(containerUuid, 10)

    // grab a host for the container
    //
    // TODO: implement this
    containerHost := "127.0.0.2"

    // use syscall to define the mounting flags
    defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

    // assemble a libcontainer config
    config := &configs.Config{
        Rootfs: rootfs,
        Capabilities: &configs.Capabilities {
            Bounding: lclusterc_caps,
            Permitted: lclusterc_caps,
            Inheritable: lclusterc_caps,
            Ambient: lclusterc_caps,
            Effective: lclusterc_caps,
        },
        Namespaces: configs.Namespaces([]configs.Namespace{
         {Type: configs.NEWNS},
         {Type: configs.NEWUTS},
         {Type: configs.NEWIPC},
         {Type: configs.NEWPID},
         {Type: configs.NEWNET},
        }),
        Cgroups: &configs.Cgroup{
         Name:   containerName,
         Parent: "system",
         Resources: &configs.Resources{
          MemorySwappiness: nil,
          AllowAllDevices: nil,
          AllowedDevices:  configs.DefaultAllowedDevices,
         },
        },
        MaskPaths: []string{
         "/proc/kcore",
        },
        ReadonlyPaths: []string{
         "/proc/sys", "/proc/sysrq-trigger", "/proc/irq", "/proc/bus",
        },
        Devices:  configs.DefaultAutoCreatedDevices,
        Hostname: containerHost,
        Mounts: []*configs.Mount{
         {
          Source:      "proc",
          Destination: "/proc",
          Device:      "proc",
          Flags:       defaultMountFlags,
         },
         {
          Source:      "tmpfs",
          Destination: "/dev",
          Device:      "tmpfs",
          Flags:       syscall.MS_NOSUID | syscall.MS_STRICTATIME,
          Data:        "mode=755",
         },
         {
          Source:      "devpts",
          Destination: "/dev/pts",
          Device:      "devpts",
          Flags:       syscall.MS_NOSUID | syscall.MS_NOEXEC,
          Data:        "newinstance,ptmxmode=0666,mode=0620,gid=5",
         },
         {
          Device:      "tmpfs",
          Source:      "shm",
          Destination: "/dev/shm",
          Data:        "mode=1777,size=65536k",
          Flags:       defaultMountFlags,
         },
         {
          Source:      "mqueue",
          Destination: "/dev/mqueue",
          Device:      "mqueue",
          Flags:       defaultMountFlags,
         },
         {
          Source:      "sysfs",
          Destination: "/sys",
          Device:      "sysfs",
          Flags:       defaultMountFlags | syscall.MS_RDONLY,
         },
        },
        Networks: []*configs.Network{
         {
          Type:    "loopback",
          Address: "127.0.0.1/0",
          Gateway: "localhost",
         },
        },
        Rlimits: []configs.Rlimit{
         {
          Type: syscall.RLIMIT_NOFILE,
          Hard: uint64(1025),
          Soft: uint64(1025),
         },
        },
    }

    // attempt to create a container instance
    container, err := containerizer.Create(containerName, config)

    // safety check, ensure this could actually make a container instance
    if err != nil {
        return nil, errorf("startJob() --> containerizer.Create() failed")
    }

    // assemble a process for the node
    process := &libcontainer.Process{
        Args:   []string{"/bin/bash"},
        Env:    []string{"PATH=/bin"},
        Stdin:  os.Stdin,
        Stdout: os.Stdout,
        Stderr: os.Stderr,
    }

    // attempt to start the given process inside of the container
    err = container.Start(process)

    // safety check, ensure that no error occurred...
    if err != nil {

        // also, if it did, attempt to clean up any memory
        container.Destroy()
        return nil, errorf("startJob() --> containerizer.Start() failed")
    }

    // Assemble the node object using the piece above
    node           := &Node{}
    node.Job_uuid   = containerUuid
    node.Container  = container
    node.Proc       = process

    // Return the new instance of node
    return node, nil
}

//! Function to stop a given process
/*
 * @param     Node    given node
 *
 * @return    error   error message, if any
 */
func (nm *NodeManager) stopProcess(node *Node) error {

    // input validation
    if node == nil {
        return errorf("stopProcess() --> Error: node does not exist!")
    }

    // ensure the node actually has a running process
    if node.Proc == nil {
        return nil
    }

    // tell the OS to send the kill signal, since this needs to stop
    err := node.Proc.Signal(os.Kill)

    // a one-job-process-per-container model means this needs to clean up
    // the associated container since this is now complete
    node.Container.Destroy()

    // if any error occurred, send 'em back
    return err
}
