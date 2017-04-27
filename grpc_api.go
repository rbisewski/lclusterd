/*
 * File: grpc_api.go
 *
 * Description: various funcs needed for the gRPC API
 */

package main

import (
	pb "./lclusterpb"
	"golang.org/x/net/context"
	"strconv"
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

	// create a start job response husk for use later on...
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
 * @param     CheckJobRequest    object to hold the check job request
 *
 * @return    CheckJobResponse   the server's response to the remote call
 * @return    error              error message, if any
 */
func (s *LclusterdServer) CheckJob(ctx context.Context,
	cjr *pb.CheckJobRequest) (*pb.CheckJobResponse, error) {

	// input validation
	if cjr == nil || cjr.Pid < 1 {
		return &pb.CheckJobResponse{Rc: -1}, errorf("CheckJob() --> " +
			"invalid input")
	}

	// grab the list of queued jobs
	response, err := etcdServer.internal.Get(ctx, queue_dir)

	// if an error occurs here, pass back a return code of 0, since for
	// whatever reason, the server is unable to query jobs at this time
	if err != nil {
		return &pb.CheckJobResponse{Rc: 0}, err
	}

	// cycle thru all of the currently queued jobs
	for _, job := range response.Kvs {

		// cast the job key to a string; it should be the uuid
		job_uuid := string(job.Key)

		// cast the CheckJobRequest pid to a string
		cjr_uuid := strconv.FormatInt(cjr.Pid, 10)

		// if a job exists with the given pid
		if job_uuid == cjr_uuid {

			// pass back a return code of 2, stating that the job is
			// present and currently queued.
			return &pb.CheckJobResponse{Rc: 2}, nil
		}
	}

	// since the job was not scheduled, perhaps it is active, so go ahead
	// and cycle thru all of the process refs
	for _, p := range processesList {

		// if a job exists with the given pid
		if p.uuid == cjr.Pid {

			// pass back a return code of 3, stating that the process is
			// present and actively running on a node.
			return &pb.CheckJobResponse{Rc: 3}, nil
		}
	}

	// since this was unable to find the job on the server, assume the
	// job does not exist
	return &pb.CheckJobResponse{Rc: 1}, nil
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

	// create a response husk for use later on...
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
