/*
 * File: misc.go
 *
 * Description: Contains a number of helpful util functions.
 *
 */

package main

import (
    "fmt"
    "os"
    "os/signal"
    "time"
)

//! Wrapper to make golang print funcs seem more C-like
/*
 * @param     string    ASCII to dump to stdout
 *
 * @return    none
 */
func printf(ascii string) {

    // Input validation
    if len(ascii) < 1 {
        return
    }

    // Attempt to print the content to stdout
    fmt.Printf(ascii + "\n")
}

//! Wrapper to give a log-like appearance to stdout
/*
 * @param     string    ASCII to dump to stdout
 *
 * @return    none
 */
func stdlog(ascii string) {

    // Input validation
    if len(ascii) < 1 {
        return
    }

    // Grab the current time in seconds from epoch.
    currentTime := time.Now().String()

    // Append the timestamp to the string message.
    printf("[" + currentTime + "] " + ascii)
}

//! Determine if a SIGINT was thrown, and if so, handle it.
/*
 * @return    none
 */
func loopUtilSIGINT() {

    // Define a new variable for dealing with OS signals
    checker := make(chan os.Signal, 1)

    // Make it notify the end-user upon receiving a signal.
    signal.Notify(checker, os.Interrupt)

    // Activate the checker
    <-checker

    // Tell stderr what occurred.
    stdlog("SIGINT detected, terminating program...\n")

    // Send the exit.
    os.Exit(0)
}
