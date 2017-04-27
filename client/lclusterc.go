/*
 * File: client/main.go
 *
 * Description: Contains the lcluster client code.
 *
 * TODO: removejob is broken, others funcs are unimplemented
 */

package main

import (
    "fmt"
    "flag"
    "time"
    "strconv"
    "golang.org/x/net/context"
    grpc "google.golang.org/grpc"
    pb "../lclusterpb"
)

// Variables to hold the result of argument flags.
var addjob string
var checkjob string
var removejob string

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

    // Dial a connection to the grpc server.
    connection, err := grpc.Dial(grpcServerAddr + grpcPort,
      grpc.WithInsecure())

    // Safety check, ensure no errors have occurred.
    if err != nil {
        fmt.Printf("addJobToServer() --> unable to connect to grpc " +
                   "server at address")
        return
    }

    // Defer the connection for the time being, but eventually it will be
    // closed once we've finished with it.
    defer connection.Close()

    // Create an lcluster client
    lcluster_client := pb.NewLclusterdClient(connection)

    // Start a new job request using the data obtained above.
    request := &pb.StartJobRequest{
		Path:     cmd,
		Args:     []string{sh, cmd},
		Env:      []string{"PATH=/bin"},
		Hostname: grpcServerAddr,
    }

    // Grab the current background context.
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)

    // Using the new job request defined above, go ahead and 
    response, err := lcluster_client.StartJob(ctx, request)

    // Cancel the current context since this has either generated a
    // response or an error.
    cancel()

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

    // attempt to cast the pid into an int since that is CheckJobRequest
    // uses as the Pid var type
    pid, err := strconv.ParseInt(uuid, 10, 64)

    // if an error occurs, print it out
    if err != nil || pid < 1 {
        fmt.Printf("Please enter a valid int64 value.\n")
        return
    }

    // Dial a connection to the grpc server.
    connection, err := grpc.Dial(grpcServerAddr + grpcPort,
      grpc.WithInsecure())

    // Safety check, ensure no errors have occurred.
    if err != nil {
        fmt.Printf("checkJobOnServer() --> unable to connect to grpc " +
                   "server at address")
        return
    }

    // Defer the connection for the time being, but eventually it will be
    // closed once we've finished with it.
    defer connection.Close()

    // Create an lcluster client
    lcluster_client := pb.NewLclusterdClient(connection)

    // Assemble a check job request object
    request := &pb.CheckJobRequest{ Pid: pid }

    // Grab the current background context.
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)

    // Using the check job request defined above, go ahead and attempt to
    // stop the job
    response, err := lcluster_client.CheckJob(ctx, request)

    // Cancel the current context since this has either generated a
    // response or an error.
    cancel()

    // Safety check, ensure that no error has occurred.
    if err != nil {
        fmt.Printf("checkJobOnServer() --> the following error has " +
                   "occurred\n" + err.Error())
    }

    // If an internal server-side error has occurred...
    if response.Rc == CJR_CORRUPTED_SERVER_INPUT {
        fmt.Printf("An internal server-side error has occurred.\n")
        return

    // Unknown job state
    } else if response.Rc == CJR_UNKNOWN {
        fmt.Printf("The job '" + uuid + "' appears to have an unknown " +
                   "status.\nConsider contacting the server operator.\n")
        return

    // Job does not exist
    } else if response.Rc == CJR_PROCESS_NOT_EXIST {
        fmt.Printf("The job '" + uuid + "' is not present.\n")
        return

    // Job is queued
    } else if response.Rc == CJR_PROCESS_QUEUED {
        fmt.Printf("The job '" + uuid + "' is queued.\n")
        return

    // Job is currently running
    } else if response.Rc == CJR_PROCESS_ACTIVE {
        fmt.Printf("The job '" + uuid + "' is currently active.\n")
        return
    }

    // otherwise if none of the above happened, print out a message
    // stating this is unknown, with the response code
    fmt.Printf("The following undefined response code was given back: " +
      strconv.FormatInt(response.Rc,10) + "\n")
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

    // attempt to cast the pid into an int since that is StopJobRequest
    // uses as the Pid var type
    pid, err := strconv.ParseInt(uuid, 10, 64)

    // if an error occurs, print it out
    if err != nil || pid < 1 {
        fmt.Printf("Please enter a valid int64 value.\n")
        return
    }

    // Dial a connection to the grpc server.
    connection, err := grpc.Dial(grpcServerAddr + grpcPort,
      grpc.WithInsecure())

    // Safety check, ensure no errors have occurred.
    if err != nil {
        fmt.Printf("removeJobFromServer() --> unable to connect to grpc server at address")
        return
    }

    // Defer the connection for the time being, but eventually it will be
    // closed once we've finished with it.
    defer connection.Close()

    // Create an lcluster client
    lcluster_client := pb.NewLclusterdClient(connection)

    // Assemble a stop job request object
    request := &pb.StopJobRequest{ Pid: pid }

    // Grab the current background context.
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)

    // Using the stop job request defined above, go ahead and attempt to
    // stop the job
    response, err := lcluster_client.StopJob(ctx, request)

    // Cancel the current context since this has either generated a
    // response or an error.
    cancel()

    // Safety check, ensure that no error has occurred.
    if err != nil {
        fmt.Printf("removeJobFromServer() --> the following error has " +
                   "occurred...\n")
        fmt.Printf(err.Error())
        return
    }

    // check the return code from the StopJobResponse object, if it -1 then
    // an error has occurred.
    if response.Rc == SJR_FAILURE {
        fmt.Printf("Server has responded with the following error:\n")
        fmt.Printf(response.Error + "\n")
        return
    }

    // with a return code of `1` then the process could not be found, but
    // the server otherwise experienced no error while processing the
    // request
    if response.Rc == SJR_DOES_NOT_EXIST {
        fmt.Printf("No such process exists with Uuid: " +
               strconv.FormatInt(pid, 10) + "\n")
        return
    }

    // Since this has obtained a response, go ahead and return the result
    if response.Rc == SJR_SUCCESS {
        fmt.Printf("Successfully removed job of pid: " +
                   strconv.FormatInt(pid, 10) + "\n")
        return
    }

    // otherwise if none of the above happened, print out a message
    // stating this is unknown, with the response code
    fmt.Printf("The following undefined response code was given back: " +
      strconv.FormatInt(response.Rc,10) + "\n")
    fmt.Printf("Consider contacting the server operator.\n")
}

//
// Main
//
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
