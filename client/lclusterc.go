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
)

// Variables to hold the result of argument flags.
var addjob string
var checkjob int64
var removejob int64

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
	if checkjob > 0 {
		checkJobOnServer(checkjob)
		return
	}

	// Safety check, ensure that remove job was given a value of 1 or
	// higher.
	if removejob > 0 {
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
	flag.Int64Var(&checkjob, "checkjob", 0,
		"Uuid of the job to query status.")

	// Argument flag for when the end-user wants to remove a job
	flag.Int64Var(&removejob, "removejob", 0,
		"Uuid of the job to be removed.")
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
		fmt.Printf("\nError: Unable to add job to queue!\n" +
                  err.Error() + "\n")
		return
	}

	// Since the job was started properly, go ahead and print out a message
	// telling the end-user the uuid of the new job.
	fmt.Printf("The requested job has been added to queue, has pid " +
          "of %d\n", response.GetPid())
}

//! Status of a job from the server.
/*
 * @param    int64    uuid of job to stat
 *
 * @return   none
 */
func checkJobOnServer(uuid int64) {

        // Connect to the server and check if the job is still running.
        response, err := libclient.HaveClientCheckJobOnServer(uuid)

	// Safety check, ensure that no error has occurred.
	if err != nil {
            fmt.Printf(err.Error())
            return
	}

	// If an internal server-side error has occurred...
	if response.Rc == lcfg.CjrCorruptedServerInput {
		fmt.Printf("An internal server-side error has occurred.\n")
		return

		// Unknown job state
	} else if response.Rc == lcfg.CjrUnknown {
		fmt.Printf("The job %d appears to have an unknown " +
			   "status.\nConsider contacting the server " +
                           "operator.\n", uuid)
		return

		// Job does not exist
	} else if response.Rc == lcfg.CjrProcessNotExist {
		fmt.Printf("The job %d is not present.\n", uuid)
		return

		// Job is queued
	} else if response.Rc == lcfg.CjrProcessQueued {
		fmt.Printf("The job %d is queued.\n", uuid)
		return

		// Job is currently running
	} else if response.Rc == lcfg.CjrProcessActive {
		fmt.Printf("The job %d is currently active.\n", uuid)
		return
	}

	// otherwise if none of the above happened, print out a message
	// stating this is unknown, with the response code
	fmt.Printf("The following undefined response code was " +
          "given back: %d\nConsider contacting the server operator.\n",
          response.Rc)
}

//! Remove a job from the server.
/*
 * @param    int64    uuid of job to remove
 *
 * @return   none
 */
func removeJobFromServer(uuid int64) {

	// Input validation
	if uuid < 1 {
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
                fmt.Printf("No such process exists with given uuid: %d\n", uuid)
		return
	}

	// Since this has obtained a response, go ahead and return the result
	if response.Rc == lcfg.SjrSuccess {
		fmt.Printf("Successfully removed job %d\n", uuid)
		return
	}

	// otherwise if none of the above happened, print out a message
	// stating this is unknown, with the response code
	fmt.Printf("The following undefined response code was given " +
          "back: %d\nConsider contacting the server operator.\n",
          response.Rc)
}
