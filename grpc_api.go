/*
 * File: grpc_api.go
 *
 * Description: various funcs needed for the gRPC API
 */

package main

import (
    "strconv"
    pb "./lclusterpb"
    "golang.org/x/net/context"
)

//! API function to start a job
/*
 * @param     Context            given host context
 * @param     StartJobRequest    object to hold the start job request
 *
 * @return    StartJobResponse   the server's response to the remote call
 * @return    error              error message, if any
 */
func (s *LclusterdServer) StartJob(ctx context.Context,
  r *pb.StartJobRequest) (*pb.StartJobResponse, error) {

    // Variable declaration
    response := &pb.StartJobResponse{}

    // Cast the job request into a Job, then attempt to add it to the
    // queue.
    pid, err := etcdServer.addToGlobalQueue((*Job)(r))

    // if any errors occur...
    if err != nil || pid < 0 {

        // set the pid to -1 and grab the error as a string
        response.Pid = -1
        response.Error = err.Error()

        // then return the failed job creation via response
        return response, err
    }

    // otherwise use the given pid and pass it back
    response.Pid = pid
    return response, nil
}

//! API function to check a job
/*
 * @param     Context            given host context
 * @param     StartJobRequest    object to hold the start job request
 *
 * @return    StartJobResponse   the server's response to the remote call
 * @return    error              error message, if any
 */
func (s *LclusterdServer) CheckJob(ctx context.Context,
  r *pb.CheckJobRequest) (*pb.CheckJobResponse, error) {

      // TODO: just a stub until this can be merged in
      return &pb.CheckJobResponse{Rc: 0}, nil
}

//! API function to stop a job
/*
 * @param     Context            given host context
 * @param     StartJobRequest    object to hold the start job request
 *
 * @return    StartJobResponse   the server's response to the remote call
 * @return    error              error message, if any
 */
func (s *LclusterdServer) StopJob(ctx context.Context,
  request *pb.StopJobRequest) (*pb.StopJobResponse, error) {

    // variable declaration
    response := &pb.StopJobResponse{}

    // request that the etcd server hand back the process
    process, err := etcdServer.obtainProcess(request.Pid)

    // if any error occurred, pass it back
    if err != nil {
        stdlog(err.Error())
        response.Rc = -1
        response.Error = err.Error()
        return response, err
    }

    // safety check, ensure this actually got a process
    if process == nil || process.proc == nil {
        stdlog("No such process exists with Uuid: " +
               strconv.FormatInt(request.Pid, 10))
        response.Rc = 1
        return response, nil
    }

    // attempt to stop the given process
    err = StopProcess(process)

    // if any error occurred while halting the process, pass it back
    if err != nil {
        stdlog(err.Error())
        response.Rc = -1
        response.Error = err.Error()
        return response, err
    }

    // otherwise the process was successfully halted, go ahead and pass
    // back a success
    response.Rc = 0
    return response, nil
}
