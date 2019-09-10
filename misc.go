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

	if len(path) < 1 {
		return false
	}

	_, err := ioutil.ReadDir(path)
	if err != nil {
		return false
	}

	return true
}

//! Determine if a SIGINT was thrown, and if so, handle it.
/*
 * @return    none
 */
func loopUtilSIGINT() {

	checker := make(chan os.Signal, 1)
	signal.Notify(checker, os.Interrupt)
	<-checker

	fmt.Printf("SIGINT detected, terminating program...\n")
	os.Exit(0)
}
