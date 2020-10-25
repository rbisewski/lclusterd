package main

import (
	"flag"
	"fmt"
	"os"

	"../config"
)

var (
	addJob   string
	checkJob int64
	killJob  int64
)

func init() {
	flag.StringVar(&addJob, "add", "",
		"Commandline program to execute; e.g. 'grep abc /path/to/file' ")

	flag.Int64Var(&checkJob, "check", 0,
		"Uuid of the job to query status.")

	flag.Int64Var(&killJob, "kill", 0,
		"Uuid of the job to be killed.")
}

func main() {

	flag.Parse()

	if len(addJob) > 0 {
		os.Exit(addJobToServer(addJob))
	}

	if checkJob > 0 {
		os.Exit(checkJobOnServer(checkJob))
	}

	if killJob > 0 {
		os.Exit(killJobFromServer(killJob))
	}

	// if no flags were given, print the usage
	flag.Usage()
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

	response, err := HaveClientAddJobToServer(cmd)
	if err != nil || response.GetPid() < 0 {
		fmt.Printf("\nError: Unable to add job to queue!\n%s\n", err.Error())
		return 1
	}

	fmt.Printf("The requested job has been added to queue, has pid of %d\n", response.GetPid())
	return 0
}

//! Status of a job from the server.
/*
 * @param    int64    uuid of job to stat
 *
 * @return   int      error code
 */
func checkJobOnServer(uuid int64) int {

	response, err := HaveClientCheckJobOnServer(uuid)
	if err != nil {
		fmt.Printf(err.Error())
		return 1
	}

	// If an internal server-side error has occurred...
	switch response.Rc {

	case config.CjrCorruptedServerInput:
		fmt.Printf("An internal server-side error has occurred.\n")
		return 1

	case config.CjrUnknown:
		fmt.Printf("The job %d appears to have an unknown "+
			"status.\nConsider contacting the server "+
			"operator.\n", uuid)
		return 0

	case config.CjrProcessNotExist:
		fmt.Printf("The job %d is not present.\n", uuid)
		return 0

	case config.CjrProcessQueued:
		fmt.Printf("The job %d is queued.\n", uuid)
		return 0

	case config.CjrProcessActive:
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
 * @param    int64    uuid of job to kill
 *
 * @return   none
 */
func killJobFromServer(uuid int64) int {

	if uuid < 1 {
		fmt.Printf("killJobFromServer() --> invalid input")
		return 1
	}

	response, _ := HaveClientStopJobOnServer(uuid)

	switch response.Rc {

	case config.SjrFailure:
		fmt.Printf("Server has responded with the following error: \n%s\n", response.Error)
		return 1

	case config.SjrDoesNotExist:
		fmt.Printf("No such process exists with given uuid: %d\n", uuid)
		return 0

	case config.SjrSuccess:
		fmt.Printf("Successfully killed job %d\n", uuid)
		return 0
	}

	fmt.Printf("The following undefined response code was given "+
		"back: %d\nConsider contacting the server operator.\n",
		response.Rc)

	return 0
}
