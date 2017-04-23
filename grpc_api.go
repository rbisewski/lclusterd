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
    uuid, err := etcdServer.addJobToQueue(sjr.Command)

    // Safety check, make sure an error didn't occur.
    if err != nil {

        // Pass back the failed job creation.
        return nil, err
    }

    // Create a new StartJobResponse object
    response := &pb.StartJobResponse{Uuid: uuid}

    // Have successfully started the job, go ahead 
    return response, nil
}


//! Returns the status of a job
/*
 * @param    Context             current process context
 * @param    CheckJobRequest     job to be stat'd
 *
 * @return   uint                response code:
 *                               0 -> unknown
 *                               1 -> job does not exist
 *                               2 -> job is queued
 *                               3 -> job is active
 *
 * @return   error               error message, if any
 */
func (lcdsv *LclusterdServer) CheckJob(ctx context.Context,
  cjr *pb.CheckJobRequest) (*pb.CheckJobResponse, error) {

      // variable declaration
      response := &pb.CheckJobResponse{}

      // input validation
      if cjr == nil || len(cjr.Uuid) < 1 {
          response.Result = 0
          response.Error = "CheckJob() --> invalid input"
          return response, errorf(response.Error)
      }

      // have the scheduler cycle thru the queue
      thatJobIsQueued, err := scheduler.jobExists(cjr.Uuid)

      // if an error occurs, pass back an 'unknown'
      if err != nil {
          response.Result = 0
          response.Error = "CheckJob() --> scheduler was unable to query job"
          return response, errorf(response.Error)
      }

      // if the scheduler found the job, pass back an int telling that the
      // 'job is queued'
      if thatJobIsQueued {
          response.Result = 2
          response.Error = ""
          return response, nil
      }

      // check if that job is actively running on one of the nodes
      thatJobIsRunning, err := nodeManager.jobRunningOnNode(cjr.Uuid)

      // if an error occurs, pass back an 'unknown'
      if err != nil {
          response.Result = 0
          response.Error = "CheckJob() --> node manager unable to query job"
          return response, errorf(response.Error)
      }

      // if the job is running, pass back the relevant int
      if thatJobIsRunning {
          response.Result = 3
          response.Error = ""
          return response, nil
      }

      // finally, if the job is neither queued nor running, pass back the
      // int that states the 'job does not exist'
      response.Result = 1
      response.Error = ""
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
    process_ref, err := etcdServer.getProcess(sjr.Uuid)

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
