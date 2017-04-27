/*
 * File: process.go
 *
 * Description: functions for handling libcontainer processes
 */
package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/opencontainers/runc/libcontainer"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
)

// The process definition.
type Process struct {

	// Id generated via lcluster
	uuid int64

	// Refs to the libcontainer object
	container libcontainer.BaseContainer
	proc      *libcontainer.Process
}

//! Function to initialize the libcontiner process handling.
/*
 * @return    none
 */
func init() {

	//
	// NOTE: this function was adapted from the etcd docs, ergo whence
	//       the original informational blurb has been copied from...
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
			stdlog("process.init() --> failure to initialize")
			stdlog(err.Error())
			return
		}
		panic("--this line should have never been executed, congratulations--")
	}
}

//! Function to start a process, as per the global rootfs.
/*
 * @param    Job    given job
 *
 * @return   Process    newly generated process
 * @return   error      error message, if any
 */
func StartProcess(j Job) (p *Process, err error) {

	// Spawn a containerizer.
	containerizer, err := libcontainer.New(rootfs, libcontainer.Cgroupfs)

	// check if an error occurred
	if err != nil {
		stdlog("StartProcess() --> libcontainer could not make a new instance!")
		return nil, err
	}

	// Make a crypto safe pseudo random string as the id.
	containerID := spawnPseudorandomString(16)

	// Set the location of the container.
	containerName := filepath.Base(j.Path) + "_" + containerID

	// Generate a new libcontainer config.
	config := generateLibcontainerConfig(containerName, j.Hostname)

	// Use the containerizer to make a container.
	container, err := containerizer.Create(containerID, config)

	// check if any error occurred
	if err != nil {
		stdlog("StartProcess() --> container creation failed!")
		return nil, err
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
		stdlog("StartProcess() --> container seems to have failed to start...")
		container.Destroy()
		return nil, err
	}

	// Assemble a reference to the newly generated process.
	return &Process{uuid: j.Pid, container: container, proc: process}, nil
}

//! Function to stop a process, as per the global rootfs.
/*
 * @param    Process    given process to stop
 *
 * @return   error      error message, if any
 */
func StopProcess(p *Process) error {

	// input validation
	if p == nil {
		return errorf("StopProcess() --> invalid input\n")
	}

	// safety check, ensure this is actually a process
	if p.proc == nil {
		debugf("StopProcess() --> no such process, ignoring")
		return nil
	}

	// Have the OS send the kill signal to the process.
	err := p.proc.Signal(os.Kill)

	// safety check, ensure this didn't fail
	if err != nil {

		// as the container is still running, attempt to free memory
		p.container.Destroy()

		// pass back the error
		return err
	}

	// attempt to free memory from the container
	p.container.Destroy()

	// attempt to remove the process from the global list of processes, in
	// the event it is still present
	if processesList[p.uuid] != nil {
		delete(processesList, p.uuid)
	}

	// everything turned out good, so pass back nil
	return nil
}
