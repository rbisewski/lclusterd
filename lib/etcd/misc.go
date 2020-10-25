package libetcd

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"strings"

	"../../lcfg"
)

//! Spawns a pseudo-random uuid based on /dev/random.
/*
 * @param    int       number of bytes
 *
 * @return   string    pseudo-random string
 */
func spawnUuid(num int) string {

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
		log.Println("spawnPseudorandomString() --> unable to spawn crypto num!")
		return ""
	}

	// Base64 encode the resulting pseudo-random bytes.
	pseudoRandStr := base64.URLEncoding.EncodeToString(byteArray)

	// safety check, ensure no error occurred
	if len(pseudoRandStr) < 1 {
		log.Println("spawnPseudorandomString() --> unable to base64 encode!")
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

	// Append the timestamp to the string message.
	log.Println("DEBUG - " + ascii + "\n")
}
