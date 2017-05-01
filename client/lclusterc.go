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
	"os"
)

// Variables to hold the result of argument flags.
var addJob string
var checkJob int64
var removeJob int64

//! The client main function.
/*
 * @return    none
 */
func main() {

	// Exit return code.
	errCode := 0

	// Parse the given argument flags.
	flag.Parse()

	// If the add job flag was passed...
	if len(addJob) > 0 {
		errCode = addJobToServer(addJob)

		// If the check job flag was passed...
	} else if checkJob > 0 {
		errCode = checkJobOnServer(checkJob)

		// Safety check, ensure that remove job was given a value of 1 or
		// higher.
	} else if removeJob > 0 {
		errCode = removeJobFromServer(removeJob)

		// Otherwise just print the usage information
	} else {
		flag.Usage()
	}

	// Send out the error code once the program is over.
	os.Exit(errCode)
}

// Initialize the flags beforehand.
func init() {

	// Argument flag for when end-user requests to add a job
	flag.StringVar(&addJob, "addjob", "",
		"Commandline program to execute; e.g. 'grep abc /path/to/file' ")

	// Argument flag for when the end-user wants to check on a job
	flag.Int64Var(&checkJob, "checkjob", 0,
		"Uuid of the job to query status.")

	// Argument flag for when the end-user wants to remove a job
	flag.Int64Var(&removeJob, "removejob", 0,
		"Uuid of the job to be removed.")
}

//! Add a job to the server.
/*
 * @param    string    commandline to execute
 *
 * @return   int       error code
 */
func addJobToServer(cmd string) int {

	// Input validation
	if len(cmd) < 1 {
		fmt.Printf("addJobToServer() --> invalid input")
		return 1
	}

	// Have the client connect to the server
	response, err := libclient.HaveClientAddJobToServer(cmd)

	// Safety check, ensure that no error has occurred.
	if err != nil || response.GetPid() < 0 {
		fmt.Printf("\nError: Unable to add job to queue!\n%s\n",
			err.Error())
		return 1
	}

	// Since the job was started properly, go ahead and print out a message
	// telling the end-user the uuid of the new job.
	fmt.Printf("The requested job has been added to queue, has pid "+
		"of %d\n", response.GetPid())
	return 0
}

//! Status of a job from the server.
/*
 * @param    int64    uuid of job to stat
 *
 * @return   int      error code
 */
func checkJobOnServer(uuid int64) int {

	// Connect to the server and check if the job is still running.
	response, err := libclient.HaveClientCheckJobOnServer(uuid)

	// Safety check, ensure that no error has occurred.
	if err != nil {
		fmt.Printf(err.Error())
		return 1
	}

	// If an internal server-side error has occurred...
	switch response.Rc {

	// Corrupted server / input.
	case lcfg.CjrCorruptedServerInput:
		fmt.Printf("An internal server-side error has occurred.\n")
		return 1

	// Unknown job status.
	case lcfg.CjrUnknown:
		fmt.Printf("The job %d appears to have an unknown "+
			"status.\nConsider contacting the server "+
			"operator.\n", uuid)
		return 0

	// Process does not exist.
	case lcfg.CjrProcessNotExist:
		fmt.Printf("The job %d is not present.\n", uuid)
		return 0

	// Process is queued.
	case lcfg.CjrProcessQueued:
		fmt.Printf("The job %d is queued.\n", uuid)
		return 0

		// Job is currently running
	case lcfg.CjrProcessActive:
		fmt.Printf("The job %d is currently active.\n", uuid)
		return 0
	}

	// otherwise if none of the above happened, print out a message
	// stating this is unknown, with the response code
	fmt.Printf("The following undefined response code was "+
		"given back: %d\nConsider contacting the server operator.\n",
		response.Rc)
	return 1
}

//! Remove a job from the server.
/*
 * @param    int64    uuid of job to remove
 *
 * @return   none
 */
func removeJobFromServer(uuid int64) int {

	// Input validation
	if uuid < 1 {
		fmt.Printf("removeJobFromServer() --> invalid input")
		return 1
	}

	// Tell the server the Uuid of the job to stop.
	response, _ := libclient.HaveClientStopJobOnServer(uuid)

	// Handle the different response codes.
	switch response.Rc {

	case lcfg.SjrFailure:
		fmt.Printf("Server has responded with the following error: \n%s\n",
			response.Error)
		return 1

	case lcfg.SjrDoesNotExist:
		fmt.Printf("No such process exists with given uuid: %d\n", uuid)
		return 0

	case lcfg.SjrSuccess:
		fmt.Printf("Successfully removed job %d\n", uuid)
		return 0
	}

	// otherwise if none of the above happened, print out a message
	// stating this is unknown, with the response code
	fmt.Printf("The following undefined response code was given "+
		"back: %d\nConsider contacting the server operator.\n",
		response.Rc)
	return 0
}
