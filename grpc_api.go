/*
 * File: grpc_api.go
 *
 * Description: contains a number of gRPC functions needed for the API
 */

package main

import (
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
        return nil, errorf("StartJob() --> invalid input\n")
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

//! Checks the status of a job
/*
 * @param    Context            current process context
 * @param    CheckJobRequest    job end-user wants to status
 *
 * @return   CheckJobResponse   result of whether the job is active
 * @return   error              error message, if any
 */
func (lcdsv *LclusterdServer) CheckJob(ctx context.Context,
  cjr *pb.CheckJobRequest) (*pb.CheckJobResponse, error) {

    // Input validation, make sure this actually got a proper request.
    if cjr == nil {
        return nil, errorf("CheckJob() --> invalid input\n")
    }

    // Check if the etcd server contains the job in question.
    job, err := etcdServer.getProcess(cjr.Pid)

    // Create a new CheckJobResponse object
    response := &pb.CheckJobResponse{}

    // Safety check, make sure the job is well formed and that an error
    // didn't occur during the call to the etcd server.
    if job == nil || job.Uuid == "" || job.Command == "" ||
      job.Machine == "" || err != nil {

        // Set the result to be false.
        response.Result = false
        response.Error = err.Error()

        // Pass back a response stating that the job has yet to occur.
        return response, err
    }

    // Since the job was well formed and no error occurred, go ahead and
    // state that the job is still working as intended.
    response.Result = true

    // Have successfully started the job, go ahead
    return response, nil
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
        return nil, errorf("StartJob() --> invalid input\n")
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
