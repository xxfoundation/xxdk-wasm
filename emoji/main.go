////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"gitlab.com/xx_network/primitives/utils"
	"net/http"
	"os"
)

// emojiMartUrl is the URL pointing to the native.JSON from emoji-mart that is
// used by front end.
//
// NOTE: This points specifically to set version 14 of the emoji-mart data. This
// URL should be updated if new sets become available.
const emojiMartUrl = "https://raw.githubusercontent.com/missive/emoji-mart/main/packages/emoji-mart-data/sets/14/native.json"

// Flag constants.
const (
	sanitizedOutputFlag = "output"
	logLevelFlag        = "logLevel"
	logFileFlag         = "logFile"
)

func main() {
	if err := sanitizeEmojis.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// sanitizeEmojis Downloads the emoji file (from emoji-mart) and sanitizes that
// list. Sanitization removes all emojis not supported by the backend. The
// sanitized JSON is returned via a file specified by the user. Refer to the flags
// for details.
var sanitizeEmojis = &cobra.Command{
	Use: "sanitizeEmojis",
	Short: "Downloads the emoji file (from emoji-mart) and sanitizes that " +
		"list. Sanitization removes all emojis not supported by the backend. " +
		"The sanitized JSON is returned via a file specified by the user." +
		"Refer to the flags for details.",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		// Initialize the logging if set
		if logFile := viper.GetString(logFileFlag); logFile != "" {
			initLog(viper.GetInt(logFileFlag), logFile)
		}

		jww.INFO.Printf("Retrieving emoji-mart JSON file...")

		// Retrieve emoji-mart file from URL
		resp, err := http.Get(emojiMartUrl)
		if err != nil {
			jww.FATAL.Panicf(
				"Failed to retrieve emoji-mart JSON from URL: %+v", err)
		}

		jww.INFO.Printf("Reading emoji-mart JSON file into bytes...")

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

		jww.INFO.Printf("Sanitizing emoji-mart JSON...")

		// Sanitize the JSON file
		backendSet := NewSet()
		sanitizedJSON, err := backendSet.SanitizeEmojiMartSet(emojiMartJson)
		if err != nil {
			jww.FATAL.Panicf("Failed to sanitize emoji-mart list: %+v", err)
		}

		jww.INFO.Printf("Outputting sanitized emoji JSON to file...")

		// Write sanitized JSON to file
		sanitizedOutputFilePath := viper.GetString(sanitizedOutputFlag)
		err = utils.WriteFileDef(sanitizedOutputFilePath, sanitizedJSON)
		if err != nil {
			jww.FATAL.Panicf(
				"Failed to write sanitized emojis to filepath %s: %+v",
				sanitizedOutputFilePath, err)
		}
	},
}

// init is the initialization function for Cobra which defines commands
// and flags.
func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	sanitizeEmojis.PersistentFlags().StringP(sanitizedOutputFlag, "o",
		"output.json",
		"File path that the sanitized JSON file will be outputted to.")
	err := viper.BindPFlag(sanitizedOutputFlag, sanitizeEmojis.PersistentFlags().
		Lookup(sanitizedOutputFlag))
	if err != nil {
		jww.FATAL.Panicf(
			"Failed to bind pf flag to %q: %+v", sanitizedOutputFlag, err)
	}

	sanitizeEmojis.PersistentFlags().StringP(logFileFlag, "l", "",
		"Path to the log output path. By default, this flag is not set "+
			"so a log will not be created unless specified.")
	err = viper.BindPFlag(logFileFlag, sanitizeEmojis.PersistentFlags().
		Lookup(logFileFlag))
	if err != nil {
		jww.FATAL.Panicf(
			"Failed to bind pf flag to %q: %+v", logFileFlag, err)
	}

	sanitizeEmojis.PersistentFlags().IntP(logLevelFlag, "v", 0,
		"Verbosity level of logging. This defaults to 0. ")
	err = viper.BindPFlag(logLevelFlag, sanitizeEmojis.PersistentFlags().
		Lookup(logLevelFlag))
	if err != nil {
		jww.FATAL.Panicf(
			"Failed to bind pf flag to %q: %+v", logLevelFlag, err)
	}

}
