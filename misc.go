/*
 * File: misc.go
 *
 * Description: Contains a number of helpful util functions.
 *
 */

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"time"
	"./lcfg"
)

//! Wrapper to make golang error funcs seem more C-like.
/*
 * @param     string    ASCII to dump to stdout
 *
 * @return    error     error message, if any
 */
func errorf(ascii string) error {

	// Input validation
	if len(ascii) < 1 {
		return nil
	}

	// Attempt to print the content to stdout
	return fmt.Errorf(ascii)
}

//! Wrapper to make golang print funcs seem more C-like.
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

//! Wrapper to give a log-like appearance to stdout.
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

//! Function to print out debug messages
/*
 * @param     string    ASCII to dump to stdout
 *
 * @return    none
 */
func debugf(ascii string) {

	// Input validation
	if len(ascii) < 1 {
		return
	}

	// ensure debug mode is actually on
	if !lcfg.DebugMode {
		return
	}

	// Grab the current time in seconds from epoch.
	currentTime := time.Now().String()

	// Append the timestamp to the string message.
	printf("[" + currentTime + "] DEBUG - " + ascii)
}

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

//! Function to grab the hostname at a given context.
/*
 * @return    string    hostname as a string
 */
func getHostname() (string, error) {

	// Attempt to grab the hostname from the OS.
	hostname, err := os.Hostname()

	// Go ahead and pass it back.
	return hostname, err
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
	stdlog("SIGINT detected, terminating program...\n")

	// Send the exit.
	os.Exit(0)
}

//! Spawns a pseudo-random string based on /dev/random.
/*
 * @param    int       number of bytes
 *
 * @return   string    pseudo-random string
 */
func spawnPseudorandomString(num int) string {

	// handle the case where an end user might enter 0 or less
	if num < 1 {
		return ""
	}

	// Assign a chunk of memory for holding the bytes.
	byteArray := make([]byte, num)

	// Populate the byte array with cryptographically secure pseudo-random
	// numbers, up to a max of `num` as per the param to this function.
	_, err := rand.Read(byteArray)

	// safety check, ensure no error occurred
	if err != nil {
		stdlog("spawnPseudorandomString() --> unable to spawn crypto num!")
		return ""
	}

	// Base64 encode the resulting pseudo-random bytes.
	pseudoRandStr := base64.URLEncoding.EncodeToString(byteArray)

	// safety check, ensure no error occurred
	if len(pseudoRandStr) < 1 {
		stdlog("spawnPseudorandomString() --> unable to base64 encode!")
		return ""
	}

	// trim away any = chars since they are not needed
	pseudoRandStr = strings.Trim(pseudoRandStr, "=")

	// replace certain non-alpha chars with alphas, if any
	pseudoRandStr = strings.Replace(pseudoRandStr, "-", "ww", -1)
	pseudoRandStr = strings.Replace(pseudoRandStr, "+", "vv", -1)
	pseudoRandStr = strings.Replace(pseudoRandStr, "_", "uu", -1)

	// otherwise return the (sufficiently?) random base64 string
	return pseudoRandStr
}
