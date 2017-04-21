/*
 * File: job.go
 *
 * Description: definition of the Job structure 
 */

package main

//
// Job structure
//
type Job struct {

    // Uuid
    Uuid string

    // Brief description of the job
    Name string

    // String representation of a POSIX command.
    Command string

    // Job starting-time in seconds from epoch
    StartTime uint64

    // Job ending-time in seconds from epoch
    EndTime uint64

    // String representation of an IPv4 address or Hostname
    Machine string
}
