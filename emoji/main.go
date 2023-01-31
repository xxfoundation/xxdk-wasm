////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

// package main is its own utility that is compiled separate from xxdk-WASM. It
// is used only to produce a compatible emoji file to be used by the frontend
// and is not a WASM module itself.

package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/xx_network/primitives/utils"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

// emojiMartUrl is the URL pointing to the native.JSON from emoji-mart that is
// used by front end.
//
// NOTE: This points specifically to set version 14 of the emoji-mart data. This
// URL should be updated if new sets become available.
const emojiMartUrl = "https://raw.githubusercontent.com/missive/emoji-mart/main/packages/emoji-mart-data/sets/14/native.json"

// Flag variables.
var (
	requestURL, outputPath, logFile string
	logLevel                        int
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Downloads the emoji file (from emoji-mart) and sanitizes that
// list. Sanitization removes all emojis not supported by the backend. The
// sanitized JSON is returned via a file specified by the user. Refer to the
// flags for details.
var cmd = &cobra.Command{
	Use: "sanitizeEmojis",
	Short: "Downloads the emoji file (from emoji-mart) and sanitizes that " +
		"list. Sanitization removes all emojis not supported by the backend. " +
		"The sanitized JSON is returned via a file specified by the user." +
		"Refer to the flags for details.",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		// Initialize the logging
		initLog(jww.Threshold(logLevel), logFile)

		// Retrieve emoji-mart file from URL
		jww.INFO.Printf("Requesting file %s", requestURL)
		resp, err := http.Get(requestURL)
		if err != nil {
			jww.FATAL.Panicf(
				"Failed to retrieve emoji-mart JSON from URL: %+v", err)
		} else if resp.StatusCode != http.StatusOK {
			jww.FATAL.Panicf("Bad status: %s", resp.Status)
		}

		jww.DEBUG.Printf("Received HTTP response: %+v", resp)

		// Read HTTP response into byte slice
		var buf bytes.Buffer
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			jww.FATAL.Panicf("Failed to read from HTTP response: %+v", err)
		}
		if err = resp.Body.Close(); err != nil {
			jww.FATAL.Panicf("Failed to close HTTP response: %+v", err)
		}
		emojiMartJson := buf.Bytes()

		jww.DEBUG.Printf("Read %d bytes of JSON file", len(emojiMartJson))

		// Sanitize the JSON file
		backendSet := NewSet()
		sanitizedJSON, err := backendSet.SanitizeEmojiMartSet(emojiMartJson)
		if err != nil {
			jww.FATAL.Panicf("Failed to sanitize emoji-mart list: %+v", err)
		}

		jww.DEBUG.Printf("Sanitised JSON file.")

		// Write sanitized JSON to file
		err = utils.WriteFileDef(outputPath, sanitizedJSON)
		if err != nil {
			jww.FATAL.Panicf(
				"Failed to write sanitized emojis to filepath %s: %+v",
				outputPath, err)
		}

		jww.INFO.Printf("Wrote sanitised JSON file to %s", outputPath)
	},
}

// init is the initialization function for Cobra which defines flags.
func init() {
	cmd.Flags().StringVarP(&requestURL, "url", "u", emojiMartUrl,
		"URL to download emoji-mart JSON file.")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "output.json",
		"Output JSON file path.")
	cmd.Flags().StringVarP(&logFile, "log", "l", "-",
		"Log output path. By default, logs are printed to stdout. "+
			"To disable logging, set this to empty (\"\").")
	cmd.Flags().IntVarP(&logLevel, "logLevel", "v", 4,
		"Verbosity level of logging. 0 = TRACE, 1 = DEBUG, 2 = INFO, "+
			"3 = WARN, 4 = ERROR, 5 = CRITICAL, 6 = FATAL")
}

// initLog will enable JWW logging to the given log path with the given
// threshold. If log path is empty, then logging is not enabled. Panics if the
// log file cannot be opened or if the threshold is invalid.
func initLog(threshold jww.Threshold, logPath string) {
	if logPath == "" {
		// Do not enable logging if no log file is set
		return
	} else if logPath != "-" {
		// Set the log file if stdout is not selected

		// Disable stdout output
		jww.SetStdoutOutput(io.Discard)

		// Use log file
		logOutput, err :=
			os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		jww.SetLogOutput(logOutput)
	}

	if threshold < jww.LevelTrace || threshold > jww.LevelFatal {
		panic("Invalid log threshold: " + strconv.Itoa(int(threshold)))
	}

	// Display microseconds if the threshold is set to TRACE or DEBUG
	if threshold == jww.LevelTrace || threshold == jww.LevelDebug {
		jww.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}

	// Enable logging
	jww.SetStdoutThreshold(threshold)
	jww.SetLogThreshold(threshold)
	jww.INFO.Printf("Log level set to: %s", threshold)
}
