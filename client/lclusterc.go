/*
 * File: client/main.go
 *
 * Description: Contains the lcluster client code.
 */

package main

import (
	"../lcfg"
        libclient "../lib/client"
	"flag"
	"fmt"
	"strconv"
)

// Variables to hold the result of argument flags.
var addjob string
var checkjob string
var removejob string

//! The client main function.
/*
 * @return    none
 */
func main() {

	// Parse the given argument flags.
	flag.Parse()

	// If the add job flag was passed...
	if len(addjob) > 0 {
		addJobToServer(addjob)
		return
	}

	// If the check job flag was passed...
	if len(checkjob) > 0 {
		checkJobOnServer(checkjob)
		return
	}

	// Safety check, ensure that remove job was given a value of 1 or
	// higher.
	if len(removejob) > 0 {
		removeJobFromServer(removejob)
		return
	}

	// Otherwise just print the usage information
	flag.Usage()
}

// Initialize the flags beforehand.
func init() {

	// Argument flag for when end-user requests to add a job
	flag.StringVar(&addjob, "addjob", "",
		"Commandline program to execute; e.g. 'grep abc /path/to/file' ")

	// Argument flag for when the end-user wants to check on a job
	flag.StringVar(&checkjob, "checkjob", "",
		"Uuid of the job to query status.")

	// Argument flag for when the end-user wants to remove a job
	flag.StringVar(&removejob, "removejob", "",
		"Pid of the job to be removed.")
}

//! Add a job to the server.
/*
 * @param    string    commandline to execute
 *
 * @return   none
 */
func addJobToServer(cmd string) {

	// Input validation
	if len(cmd) < 1 {
		fmt.Printf("addJobToServer() --> invalid input")
		return
	}

        // Have the client connect to the server
        response, err := libclient.HaveClientAddJobToServer(cmd)

	// Safety check, ensure that no error has occurred.
	if err != nil || response.GetPid() < 0 {
		fmt.Printf("\nError: Unable to add job to queue!\n")
		fmt.Printf(err.Error() + "\n")
		return
	}

	// Since the job was started properly, go ahead and print out a message
	// telling the end-user the uuid of the new job.
	fmt.Printf("The requested job has been added to queue, has pid of " +
		strconv.FormatInt(response.GetPid(), 10) + "\n")
}

//! Status of a job from the server.
/*
 * @param    string    uuid of job to stat
 *
 * @return   none
 */
func checkJobOnServer(uuid string) {

	// Input validation
	if len(uuid) < 1 {
		fmt.Printf("checkJobOnServer() --> invalid input")
		return
	}

        // Connect to the server and check if the job is still running.
        response, err := libclient.HaveClientCheckJobOnServer(uuid)

	// Safety check, ensure that no error has occurred.
	if err != nil {
		fmt.Printf("checkJobOnServer() --> the following error has " +
			"occurred\n" + err.Error())
	}

	// If an internal server-side error has occurred...
	if response.Rc == lcfg.CjrCorruptedServerInput {
		fmt.Printf("An internal server-side error has occurred.\n")
		return

		// Unknown job state
	} else if response.Rc == lcfg.CjrUnknown {
		fmt.Printf("The job '" + uuid + "' appears to have an unknown " +
			"status.\nConsider contacting the server operator.\n")
		return

		// Job does not exist
	} else if response.Rc == lcfg.CjrProcessNotExist {
		fmt.Printf("The job '" + uuid + "' is not present.\n")
		return

		// Job is queued
	} else if response.Rc == lcfg.CjrProcessQueued {
		fmt.Printf("The job '" + uuid + "' is queued.\n")
		return

		// Job is currently running
	} else if response.Rc == lcfg.CjrProcessActive {
		fmt.Printf("The job '" + uuid + "' is currently active.\n")
		return
	}

	// otherwise if none of the above happened, print out a message
	// stating this is unknown, with the response code
	fmt.Printf("The following undefined response code was given back: " +
		strconv.FormatInt(response.Rc, 10) + "\n")
	fmt.Printf("Consider contacting the server operator.\n")
}

//! Remove a job from the server.
/*
 * @param    string    uuid of job to remove
 *
 * @return   none
 */
func removeJobFromServer(uuid string) {

	// Input validation
	if len(uuid) < 1 {
		fmt.Printf("removeJobFromServer() --> invalid input")
		return
	}

        // Tell the server the Uuid of the job to stop.
        response, err := libclient.HaveClientStopJobOnServer(uuid)

	// check the return code from the StopJobResponse object, if it -1 then
	// an error has occurred.
	if err != nil || response.Rc == lcfg.SjrFailure {
		fmt.Printf("Server has responded with the following error:\n")
		fmt.Printf(response.Error + "\n")
		return
	}

	// with a return code of `1` then the process could not be found, but
	// the server otherwise experienced no error while processing the
	// request
	if response.Rc == lcfg.SjrDoesNotExist {
		fmt.Printf("No such process exists with Uuid: " + uuid + "\n")
		return
	}

	// Since this has obtained a response, go ahead and return the result
	if response.Rc == lcfg.SjrSuccess {
		fmt.Printf("Successfully removed job of pid: " + uuid + "\n")
		return
	}

	// otherwise if none of the above happened, print out a message
	// stating this is unknown, with the response code
	fmt.Printf("The following undefined response code was given back: " +
		strconv.FormatInt(response.Rc, 10) + "\n")
	fmt.Printf("Consider contacting the server operator.\n")
}

