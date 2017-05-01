/*
 * File: misc.go
 *
 * Description: Contains a number of helpful util functions.
 *
 */

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
)

//! Determine if a given directory location actually exists.
/*
 * @param     string    location of the form: /path/to/directory/
 *
 * @return    bool      whether or not this location is a POSIX dir
 */
func directoryExists(path string) bool {

	// input validation
	if len(path) < 1 {
		return false
	}

	// attempt to read the directory and its contents, if any
	_, err := ioutil.ReadDir(path)

	// if an error occurred, then assume this is probably not a directory
	if err != nil {
		return false
	}

	// otherwise this succeeded, in which case return true
	return true
}

//! Determine if a SIGINT was thrown, and if so, handle it.
/*
 * @return    none
 */
func loopUtilSIGINT() {

	// Define a new variable for dealing with OS signals.
	checker := make(chan os.Signal, 1)

	// Make it notify the end-user upon receiving a signal.
	signal.Notify(checker, os.Interrupt)

	// Activate the checker
	<-checker

	// Tell stderr what occurred.
	fmt.Printf("SIGINT detected, terminating program...\n")

	// Send the exit.
	os.Exit(0)
}
