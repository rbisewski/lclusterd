/*
 * File: client/main.go
 *
 * Description: Contains the lcluster client code.
 */

package main

import (
    "fmt"
    "flag"
    "time"

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
      "Uuid of the job to be removed.")
}

//! Add a job to the server.
/*
 * @param    string    commandline to execute
 *
 * @return   int64     pid
 */
func addJobToServer(cmd string) int64 {

    // Input validation
    if len(cmd) < 1 {
        fmt.Printf("addJobToServer() --> invalid input")
        return -1
    }

    // Dial a connection to the grpc server.
    connection, err := grpc.Dial(grpcServerAddr + grpcPort,
      grpc.WithInsecure())

    // Safety check, ensure no errors have occurred.
    if err != nil {
        fmt.Printf("addJobToServer() --> unable to connect to grpc server at address")
        return -1
    }

    // Defer the connection for the time being, but eventually it will be
    // closed once we've finished with it.
    defer connection.Close()

    // Create an lcluster client
    lcluster_client := pb.NewLclusterdClient(connection)

    // Start a new job request using the data obtained above.
    request := &pb.StartJobRequest{
        Command: addjob,
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
        fmt.Printf("addJobToServer() --> the following error has occurred\n")
        fmt.Printf(err.Error())
    }

    // Since the job was started properly, go ahead and print out a message
    // telling the end-user the uuid of the new job.
    fmt.Printf("The requested job has been added to queue, has pid of " +
      response.GetUuid() + "\n")

    // Return the resulting start job response uuid.
    //return response.GetUuid()
    return -1
}

//! Print out the list of jobs currently on the server.
/*
 * @param    string    uuid of job to remove
 *
 * @return   none
 */
func listJobsOnServer() {
    // TODO: implement this
}

//! Status of a job from the server.
/*
 * @param    string    uuid of job to remove
 *
 * @return   none
 */
func checkJobOnServer(uuid string) {
    // TODO: implement this
}

//! Remove a job from the server.
/*
 * @param    string    uuid of job to remove
 *
 * @return   none
 */
func removeJobFromServer(uuid string) bool {

    // Input validation
    if len(uuid) < 1 {
        fmt.Printf("addJobToServer() --> invalid input")
        return false
    }

    // Dial a connection to the grpc server.
    connection, err := grpc.Dial(grpcServerAddr + grpcPort,
      grpc.WithInsecure())

    // Safety check, ensure no errors have occurred.
    if err != nil {
        fmt.Printf("addJobToServer() --> unable to connect to grpc server at address")
        return false
    }

    // Defer the connection for the time being, but eventually it will be
    // closed once we've finished with it.
    defer connection.Close()

    // Create an lcluster client
    lcluster_client := pb.NewLclusterdClient(connection)

    // Assemble a stop job request object
    request := & pb.StopJobRequest{ Uuid: uuid }

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
                   "occurred\n")
        fmt.Printf(err.Error())
    }

    // Since this has obtained a response, go ahead and return the result
    return response.Result
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
    if len(removejob) > 0 {
        removeJobFromServer(removejob)
        return
    }

    // Otherwise just print the usage information
    flag.Usage()
}
