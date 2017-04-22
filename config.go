/*
 * File: config.go
 *
 * Description: stores various hardcoded essential variables
 */

package main

// According to the etcd documentation, the clients listen on 2379
const etcdSocket      = "localhost:2379"

// IPv4 address of the gRPC server
const grpcServerAddr  = "localhost"

// High number port for use by the gRPC server
const grpcPort        = ":64051"

// Etcd grace period, in seconds
const etcdGracePeriod = 5

//
// Process state defines
//
const STOPPED = 0
const RUNNING = 1
const ABORTED = 2
const ERROR   = 3

/* --------------------------------------------------
 * Capabilities the libcontainer instances will need:
 * --------------------------------------------------
 *
 * chown
 * dac_override
 * fsetid
 * fowner
 * mknod
 * net_raw
 * setgid
 * setuid
 * setfcap
 * setpcap
 * net_bind_service
 * sys_chroot
 * kill
 * audit_write
 *
 * --------------------------------------------------
 */
var lclusterc_caps = []string {
                               "CAP_CHOWN",
                               "CAP_DAC_OVERRIDE",
                               "CAP_FSETID",
                               "CAP_FOWNER",
                               "CAP_MKNOD",
                               "CAP_NET_RAW",
                               "CAP_SETGID",
                               "CAP_SETUID",
                               "CAP_SETFCAP",
                               "CAP_SETPCAP",
                               "CAP_NET_BIND_SERVICE",
                               "CAP_SYS_CHROOT",
                               "CAP_KILL",
                               "CAP_AUDIT_WRITE",
                              }
