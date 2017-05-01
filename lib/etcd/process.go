/*
 * File: process.go
 *
 * Description: functions for handling libcontainer processes
 */
package libetcd

import (
	"github.com/opencontainers/runc/libcontainer"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	"os"
	"log"
	"path/filepath"
	"runtime"
)

// The process definition.
type Process struct {

	// Id generated via lcluster
	Uuid int64

	// Refs to the libcontainer object
	container libcontainer.BaseContainer
	Proc      *libcontainer.Process
}

//! Function to initialize the libcontiner process handling.
/*
 * @return    none
 */
func init() {

	//
	// NOTE: this function was adapted from the libcontainer docs:
	//
	// github.com/opencontainers/runc/blob/master/libcontainer/README.md
	//

	// Because containers are spawned in a two step process you will need
	// a binary that will be executed as the init process for the
	// container. In libcontainer, we use the current binary,
	// in /proc/self/exe, to be executed as the init process, and use arg
	// "init", we call the first step process "bootstrap", so you always
	// need a "init" function as the entry of "bootstrap"
	//
	if len(os.Args) > 1 && os.Args[1] == "init" {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()
		factory, _ := libcontainer.New("")
		if err := factory.StartInitialization(); err != nil {
			log.Println("process.init() --> failure to initialize")
			log.Println(err.Error())
			return
		}
		panic("--this line should have never been executed, congratulations--")
	}
}

//! Function to start a process, as per the global rootfs.
/*
 * @param    Job        given job
 * @param    string     rootfs
 *
 * @return   Process    newly generated process
 * @return   error      error message, if any
 */
func (inst *EtcdInstance) StartProcess(j Job) (p Process, err error) {

	// Spawn a containerizer.
	containerizer, err := libcontainer.New(inst.rootfs,
		libcontainer.Cgroupfs)

	// check if an error occurred
	if err != nil {
		return Process{}, err
	}

	// Make a crypto safe pseudo random string as the id.
	containerID := spawnUuid(16)

	// Set the location of the container.
	containerName := filepath.Base(j.Path) + "_" + containerID

	// Generate a new libcontainer config.
	config := GenerateRuncConfig(containerName, j.Hostname, inst.rootfs)

	// Use the containerizer to make a container.
	container, err := containerizer.Create(containerID, config)

	// check if any error occurred
	if err != nil {
		return Process{}, err
	}

	// Assemble a new process.
	process := &libcontainer.Process{
		Args:   append([]string{j.Path}, j.Args...),
		Env:    j.Env,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	// get the container to start the given process
	err = container.Start(process)

	// safety check, make sure this actually worked; also clean mem in the
	// event that this is no longer able to continue
	if err != nil {
		container.Destroy()
		return Process{}, err
	}

	// Assemble a reference to the newly generated process.
	return Process{Uuid: j.Pid, container: container, Proc: process}, nil
}

//! Function to stop a process, as per the global rootfs.
/*
 * @param    Process    given process to stop
 *
 * @return   error      error message, if any
 */
func (inst *EtcdInstance) StopProcess(p Process) error {

	// safety check, ensure this is actually a process
	if p.Proc == nil {
		debugf("StopProcess() --> no such process, ignoring")
		return nil
	}

	// Have the OS send an interrupt signal to the process.
	err := p.Proc.Signal(os.Interrupt)

	// safety check, ensure this didn't fail
	if err != nil {

		// Since attempting to interrupt the process failed, have
		// the OS kill it instead.
		p.Proc.Signal(os.Kill)

		// as the container is still running, attempt to free memory
		p.container.Destroy()

		// pass back the error
		return err
	}

	// attempt to free memory from the container
	p.container.Destroy()

	// attempt to remove the process from the global list of processes, in
	// the event it is still present
	if inst.ProcessesList[p.Uuid] != nil {
		delete(inst.ProcessesList, p.Uuid)
	}

	// everything turned out good, so pass back nil
	return nil
}
