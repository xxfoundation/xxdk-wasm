////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/logging"
	"gitlab.com/elixxir/xxdk-wasm/worker"
	"syscall/js"
)

func init() {
	// Set up Javascript console listener set at level INFO
	ll := logging.NewJsConsoleLogListener(jww.LevelDebug)
	jww.SetLogListeners(ll.Listen)
	jww.SetStdoutThreshold(jww.LevelFatal + 1)
}

// workerLogFile manages communication with the main thread and writing incoming
// logging messages to the log file.
type workerLogFile struct {
	wtm *worker.ThreadManager
	lf  *logging.LogFile
}

func main() {
	fmt.Println("[WW] Starting xxDK WebAssembly Log Worker.")
	jww.INFO.Print("[WW] Starting xxDK WebAssembly Log Worker.")

	js.Global().Set("LogLevel", js.FuncOf(logging.LogLevelJS))

	wlf := workerLogFile{
		wtm: worker.NewThreadManager("ChannelsIndexedDbWorker", false)}

	wlf.registerCallbacks()

	wlf.wtm.SignalReady()
	<-make(chan bool)
	fmt.Println("[WW] Closing xxDK WebAssembly Log Worker.")
}

// registerCallbacks registers all the necessary callbacks for the main thread
// to get the file and file metadata.
func (wlf *workerLogFile) registerCallbacks() {
	// Callback for logging.LogToFileWorker
	wlf.wtm.RegisterCallback(logging.NewLogFileTag,
		func(data []byte) ([]byte, error) {
			var msg logging.NewLogFileMessage
			err := json.Unmarshal(data, &msg)
			if err != nil {
				return []byte(err.Error()), err
			}

			wlf.lf, err = logging.NewLogFile(
				msg.LogFileName, msg.Threshold, msg.MaxLogFileSize)
			if err != nil {
				return []byte(err.Error()), err
			}

			jww.DEBUG.Printf(
				"[LOG] Created new worker log file %q of size %d with level %s",
				msg.LogFileName, msg.MaxLogFileSize, msg.Threshold)

			return []byte{}, nil
		})

	// Callback for LogFileWorker.Name
	wlf.wtm.RegisterCallback(logging.NameTag, func([]byte) ([]byte, error) {
		return []byte(wlf.lf.Name()), nil
	})

	// Callback for LogFileWorker.Threshold
	wlf.wtm.RegisterCallback(logging.ThresholdTag, func([]byte) ([]byte, error) {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(wlf.lf.Threshold()))
		return b, nil
	})

	// Callback for LogFileWorker.GetFile
	wlf.wtm.RegisterCallback(logging.GetFileTag, func([]byte) ([]byte, error) {
		return wlf.lf.GetFile(), nil
	})

	// Callback for LogFileWorker.MaxSize
	wlf.wtm.RegisterCallback(logging.MaxSizeTag, func([]byte) ([]byte, error) {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(wlf.lf.MaxSize()))
		return b, nil
	})

	// Callback for LogFileWorker.Size
	wlf.wtm.RegisterCallback(logging.SizeTag, func([]byte) ([]byte, error) {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(wlf.lf.Size()))
		return b, nil
	})
}
