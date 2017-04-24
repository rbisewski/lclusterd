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
var removejob int64

// Initialize the flags beforehand.
func init() {

    // Argument flag for when end-user requests to add a job
    flag.StringVar(&addjob, "addjob", "",
      "Commandline program to execute; e.g. 'grep abc /path/to/file' ")

    // Argument flag for when the end-user wants to check on a job
    flag.StringVar(&checkjob, "checkjob", "",
      "Uuid of the job to query status.")

    // Argument flag for when the end-user wants to remove a job
    flag.Int64Var(&removejob, "removejob", -1,
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
    if err != nil {
        fmt.Printf("addJobToServer() --> the following error has occurred\n" +
          err.Error())
        return
    }

    // Since the job was started properly, go ahead and print out a message
    // telling the end-user the uuid of the new job.
    fmt.Printf("The requested job has been added to queue, has pid of " +
      strconv.FormatInt(response.GetPid(), 10) + "\n")
}

//! Print out the list of jobs currently on the server.
/*
 * @return   none
 *
 * SIDE EFFECT: prints out a list of jobs to the client's stdout
 */
func listJobsOnServer() {
    /*
    // Dial a connection to the grpc server.
    connection, err := grpc.Dial(grpcServerAddr + grpcPort,
      grpc.WithInsecure())

    // Safety check, ensure no errors have occurred.
    if err != nil {
        fmt.Printf("listJobsOnServer() --> unable to connect to grpc server at address")
        return
    }

    // Defer the connection for the time being, but eventually it will be
    // closed once we've finished with it.
    defer connection.Close()

    // Create an lcluster client
    lcluster_client := pb.NewLclusterdClient(connection)

    // Assemble a check job request object
    request := &pb.ListJobsRequest{}

    // Grab the current background context.
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)

    // Using the check job request defined above, go ahead and attempt to
    // stop the job
    response, err := lcluster_client.ListJobs(ctx, request)

    // Cancel the current context since this has either generated a
    // response or an error.
    cancel()

    // Safety check, ensure that no error has occurred.
    if err != nil {
        fmt.Printf("listJobsOnServer() --> the following error has " +
                   "occurred\n" + err.Error())
    }

    // the response queue is empty, print out a short message stating this
    if len(response.Qcontents) < 1 {
        fmt.Printf("The queue is currently empty.\n")
        return
    }

    // print out a list of jobs currently in the queue
    fmt.Printf("The following jobs are queued.\n")
    fmt.Printf(response.Qcontents)
    */
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
/*
    // Dial a connection to the grpc server.
    connection, err := grpc.Dial(grpcServerAddr + grpcPort,
      grpc.WithInsecure())

    // Safety check, ensure no errors have occurred.
    if err != nil {
        fmt.Printf("checkJobOnServer() --> unable to connect to grpc server at address")
        return
    }

    // Defer the connection for the time being, but eventually it will be
    // closed once we've finished with it.
    defer connection.Close()

    // Create an lcluster client
    lcluster_client := pb.NewLclusterdClient(connection)

    // Assemble a check job request object
    request := &pb.CheckJobRequest{ Uuid: uuid }

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

    // Unknown job state
    if response.Result == 0 {
        fmt.Printf("An unknown server-side error has occurred.")

    // Job does not exist
    } else if response.Result == 1 {
        fmt.Printf("The job '" + uuid + "' is not present.")

    // Job is queued
    } else if response.Result == 2 {
        fmt.Printf("The job '" + uuid + "' is queued.")

    // Job is currently running
    } else if response.Result == 3 {
        fmt.Printf("The job '" + uuid + "' is currently active.")
    }
    */
}

//! Remove a job from the server.
/*
 * @param    string    uuid of job to remove
 *
 * @return   none
 */
func removeJobFromServer(pid int64) {

    // Input validation
    if pid < 1 {
        fmt.Printf("removeJobFromServer() --> invalid input")
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
    if response.Rc == -1 {
        fmt.Printf("removeJobFromServer() --> the following error has " +
                   "occurred...\n")
        fmt.Printf(response.Error)
        return
    }

    // Since this has obtained a response, go ahead and return the result
    fmt.Printf("Successfully removed job of pid: " +
               strconv.FormatInt(pid, 10) + "\n")
    return
}

//
// Main
//
func main() {

    // Parse the given argument flags.
    flag.Parse()

    // In order for the client to actually work, only one argument must
    // be given; to add / check a job present on the server, etc.
    if len(addjob) < 1 && len(checkjob) < 1 {
        flag.Usage()
        return
    }

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
    if removejob > 0 {
        removeJobFromServer(removejob)
        return
    }

    // Otherwise just print the usage information
    flag.Usage()
}
