/*
 * File: node.go
 *
 * Description: definition of the Node structure 
 */

package main

//
// Node structure
//
type Node struct {

    // String representation of an IPv4 address or Hostname
    Host string

    // Job uuid
    Job_uuid string

    // Whether or not the node is available for handling jobs
    Locked bool
}
