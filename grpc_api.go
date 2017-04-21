/*
 * File: grpc_api.go
 *
 * Description: contains a number of gRPC functions needed for the API
 */

package main

import (
    "fmt"
    "golang.org/x/net/context"
    pb "./lclusterpb"
)

// Definition of the LclusterdServer, useful with grpc proto.
type LclusterdServer struct {
}

//! Starts a job and adds it to the queue
/*
 * @param    Context            current process context
 * @param    StartJobRequest    job end-user wants to start
 *
 * @return   StartJobResponse   result from attempting to start the job 
 * @return   error              error message, if any
 */
func (lcdsv *LclusterdServer) StartJob(ctx context.Context,
  sjr *pb.StartJobRequest) (*pb.StartJobResponse, error) {

    // Input validation, make sure this actually got a proper request.
    if sjr == nil {
        return nil, fmt.Errorf("StartJob() --> invalid input\n")
    }

    // Add the new job to the queue.
    pid, err := etcdServer.addJobToQueue((*Job)(sjr))

    // Create a new StartJobResponse object
    response := &pb.StartJobResponse{}

    // Safety check, make sure an error didn't occur.
    if err != nil {

        // Set the pid to -1 and pass along the error.
        response.Pid = -1
        response.Error = err.Error()

        // Pass back the failed job creation.
        return response, err
    }

    // Assign the pid of the new job.
    response.Pid = pid

    // Have successfully started the job, go ahead 
    return response, err
}

//! Stops a job currently being ran
/*
 * @param    Context            current process context
 * @param    StopJobRequest     job end-user wants to stop
 *
 * @return   StopJobResponse    result from attempting to start the job 
 * @return   error              error message, if any
 */
func (lcdsv *LclusterdServer) StopJob(ctx context.Context,
  sjr *pb.StopJobRequest) (*pb.StopJobResponse, error) {

    // Input validation, make sure this actually got a proper stop request.
    if sjr == nil {
        return nil, fmt.Errorf("StartJob() --> invalid input\n")
    }

    // Grab a reference to the desired process.
    process_ref, err := etcdServer.getProcess(sjr.Pid)

    // Safety check, ensure that errors have not occurred.
    if err != nil {

        // Send back an response stating that the job stopping has has a
        // failed result, and the error explaining why.
        return &pb.StopJobResponse{Result: false, Error: err.Error()}, err
    }

    // Attempt to halt the process.
    err = etcdServer.haltJob(process_ref)

    // Safety check, ensure no error has occurred.
    if err != nil {

        // Send back an response stating that the job stopping has has a
        // failed result, and the error explaining why.
        return &pb.StopJobResponse{Result: false, Error: err.Error()}, err
    }

    // Otherwise return success.
    return &pb.StopJobResponse{Result: true}, nil
}

//! Returns the status of a job
/*
 * TODO: complete this function
 *
 * @param
 *
 * @return    
 */
