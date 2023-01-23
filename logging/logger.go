////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package logging

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/smallnest/ringbuffer"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"gitlab.com/elixxir/xxdk-wasm/worker"
	"io"
	"strconv"
	"sync/atomic"
	"syscall/js"
	"time"
)

const (
	// defaultInitThreshold is the log threshold used for the initial log before
	// any logging options is set.
	defaultInitThreshold = jww.LevelTrace

	// logListenerChanSize is the size of the listener channel that stores log
	// messages before they are written.
	logListenerChanSize = 250
)

// logger is the global that all jwalterweatherman logging is sent to.
var logger *Logger

func init() {
	logger = InitLogFile()
}

// GetLogger returns the Logger object, used to manager where logging is
// recorded.
func GetLogger() *Logger {
	return logger
}

// List of tags that can be used when sending a message or registering a handler
// to receive a message.
const (
	NewLogFileTag worker.Tag = "NewLogFile"
	WriteLogTag   worker.Tag = "WriteLog"
	GetFileTag    worker.Tag = "GetFile"
	GetFileExtTag worker.Tag = "GetFileExt"
	SizeTag       worker.Tag = "Size"
)

// Logger manages the recording of jwalterweatherman logs. It can write logs to
// a local, in-memory buffer or to an external worker.
type Logger struct {
	threshold      jww.Threshold
	maxLogFileSize int
	logListenerID  uint64

	listenChan  chan []byte
	mode        atomic.Uint32
	processQuit chan struct{}

	rb *ringbuffer.RingBuffer
	wm *worker.Manager
}

// InitLogFile creates a new Logger that begins storing the first 250 log
// entries. If either the log file or log worker is enabled, then these logs are
// redirected to the set destination. If the channel fills up with no log
// recorder enabled, then the listener is disabled.
func InitLogFile() *Logger {
	lf := &Logger{
		threshold:   defaultInitThreshold,
		listenChan:  make(chan []byte, logListenerChanSize),
		mode:        atomic.Uint32{},
		processQuit: make(chan struct{}),
	}
	lf.setMode(initMode)

	// Add the log listener
	lf.logListenerID = AddLogListener(lf.Listen)

	jww.INFO.Printf("Enabled initial log file listener at threshold %s that "+
		"can store %d entries", lf.threshold, len(lf.listenChan))

	return lf
}

// LogToFile starts logging to a local, in-memory log file.
func (l *Logger) LogToFile(threshold jww.Threshold, maxLogFileSize int) error {
	err := l.prepare(threshold, maxLogFileSize, fileMode)
	if err != nil {
		return err
	}
	l.rb = ringbuffer.New(maxLogFileSize)

	go l.processFileLog(l.processQuit)

	return nil
}

// processFileLog processes the log messages sent to the listener channel and
// writes them to the local ring buffer.
func (l *Logger) processFileLog(quit chan struct{}) {
	jww.INFO.Printf("Starting log file processing thread.")

	for {
		select {
		case <-quit:
			jww.INFO.Printf("Stopping log file processing thread.")
			return
		case p := <-l.listenChan:
			_, err := l.rb.Write(p)
			if err != nil {
				jww.ERROR.Printf(
					"Failed to write log entry to ring buffer: %+v", err)
			}
		}
	}
}

// LogToFileWorker starts a new worker that begins listening for logs and
// writing them to file. This function blocks until the worker has started.
func (l *Logger) LogToFileWorker(threshold jww.Threshold, maxLogFileSize int,
	wasmJsPath, workerName string) error {
	err := l.prepare(threshold, maxLogFileSize, workerMode)
	if err != nil {
		return err
	}

	// Create new worker manager, which will start the worker and wait until
	// communication has been established
	wm, err := worker.NewManager(wasmJsPath, workerName, false)
	if err != nil {
		return err
	}
	l.wm = wm

	// Register the callback used by the Javascript to request the log file.
	// This prevents an error print when GetFileExtTag is not registered.
	l.wm.RegisterCallback(GetFileExtTag, func([]byte) {
		jww.DEBUG.Print("[WW] Received file requested from external " +
			"Javascript. Ignoring file.")
	})

	data, err := json.Marshal(l.maxLogFileSize)
	if err != nil {
		return err
	}

	// Send message to initialize the log file listener
	errChan := make(chan error)
	l.wm.SendMessage(NewLogFileTag, data, func(data []byte) {
		if len(data) > 0 {
			errChan <- errors.New(string(data))
		} else {
			errChan <- nil
		}
	})

	// Wait for worker to respond
	select {
	case err = <-errChan:
		if err != nil {
			return err
		}
	case <-time.After(worker.ResponseTimeout):
		return errors.Errorf("timed out after %s waiting for new log "+
			"file in worker to initialize", worker.ResponseTimeout)
	}

	jww.INFO.Printf("Initialized log to file web worker %s.", workerName)

	go l.processWorkerLog(l.processQuit)

	return nil
}

// processWorkerLog processes the log messages sent to the listener channel and
// sends them to the worker to be logged.
func (l *Logger) processWorkerLog(quit chan struct{}) {
	jww.INFO.Printf("Starting worker log file processing thread.")

	for {
		select {
		case <-quit:
			jww.INFO.Printf("Stopping worker log file processing thread.")
			return
		case p := <-l.listenChan:
			l.wm.SendMessage(WriteLogTag, p, nil)
		}
	}
}

// prepare sets the threshold, maxLogFileSize, and mode of the logger and
// prints a log message indicating this information.
func (l *Logger) prepare(
	threshold jww.Threshold, maxLogFileSize int, m mode) error {
	if m := l.getMode(); m != initMode {
		return errors.Errorf("log already set to %s", m)
	} else if threshold < jww.LevelTrace || threshold > jww.LevelFatal {
		return errors.Errorf("log level of %d is invalid", threshold)
	}

	l.threshold = threshold
	l.maxLogFileSize = maxLogFileSize
	l.setMode(m)

	msg := fmt.Sprintf("Outputting log to file in %s of max size %d with "+
		"level %s", m, l.MaxSize(), l.Threshold())
	switch l.Threshold() {
	case jww.LevelTrace:
		fallthrough
	case jww.LevelDebug:
		fallthrough
	case jww.LevelInfo:
		jww.INFO.Print(msg)
	case jww.LevelWarn:
		jww.WARN.Print(msg)
	case jww.LevelError:
		jww.ERROR.Print(msg)
	case jww.LevelCritical:
		jww.CRITICAL.Print(msg)
	case jww.LevelFatal:
		jww.FATAL.Print(msg)
	}

	return nil
}

// StopLogging stops the logging of log messages and disables the log listener.
// If the log worker is running, it is terminated. Once logging is stopped, it
// cannot be resumed the log file cannot be recovered.
func (l *Logger) StopLogging() {
	RemoveLogListener(l.logListenerID)

	switch l.getMode() {
	case workerMode:
		l.wm.Terminate()
	case fileMode:
		l.rb.Reset()
	}

	select {
	case l.processQuit <- struct{}{}:
	default:
		jww.ERROR.Printf("Failed to stop quit processes.")
	}
}

// Listen is called for every logging event. This function adheres to the
// [jwalterweatherman.LogListener] type.
func (l *Logger) Listen(t jww.Threshold) io.Writer {
	if t < l.threshold {
		return nil
	}

	return l
}

// Write sends the bytes to the listener channel. It always returns the length
// of p and a nil error.
func (l *Logger) Write(p []byte) (n int, err error) {
	select {
	case l.listenChan <- p:
	default:
		jww.ERROR.Printf("Logger channel filled. Log file recording stopping.")
		l.StopLogging()
		return 0, errors.Errorf(
			"Logger channel filled. Log file recording stopping.")
	}
	return len(p), nil
}

// GetFile returns the entire log file.
//
// If the log file is listening locally, it returns it from the local buffer. If
// it is listening from the worker, it blocks until the file is returned.
func (l *Logger) GetFile() []byte {
	switch l.getMode() {
	case fileMode:
		return l.rb.Bytes()
	case workerMode:
		fileChan := make(chan []byte)
		l.wm.SendMessage(GetFileTag, nil, func(data []byte) { fileChan <- data })

		select {
		case file := <-fileChan:
			return file
		case <-time.After(worker.ResponseTimeout):
			jww.FATAL.Panicf("Timed out after %s waiting for log file from "+
				"worker", worker.ResponseTimeout)
			return nil
		}
	default:
		return nil
	}
}

// Threshold returns the log level threshold used in the file.
func (l *Logger) Threshold() jww.Threshold {
	return l.threshold
}

// MaxSize returns the max size, in bytes, that the log file is allowed to be.
func (l *Logger) MaxSize() int {
	return l.maxLogFileSize
}

// Size returns the current size, in bytes, written to the log file.
//
// If the log file is listening locally, it returns it from the local buffer. If
// it is listening from the worker, it blocks until the size is returned.
func (l *Logger) Size() int {
	switch l.getMode() {
	case fileMode:
		return l.rb.Length()
	case workerMode:
		sizeChan := make(chan []byte)
		l.wm.SendMessage(SizeTag, nil, func(data []byte) { sizeChan <- data })

		select {
		case data := <-sizeChan:
			return int(jww.Threshold(binary.LittleEndian.Uint64(data)))
		case <-time.After(worker.ResponseTimeout):
			jww.FATAL.Panicf("Timed out after %s waiting for log file size "+
				"from worker", worker.ResponseTimeout)
			return 0
		}
	default:
		return 0
	}
}

////////////////////////////////////////////////////////////////////////////////
// Log File Mode                                                              //
////////////////////////////////////////////////////////////////////////////////

// mode represents the state of the Logger.
type mode uint32

const (
	initMode mode = iota
	fileMode
	workerMode
)

// String returns a human-readable representation of the mode for logging and
// debugging. This function adheres to the fmt.Stringer interface.
func (m mode) String() string {
	switch m {
	case initMode:
		return "uninitialized"
	case fileMode:
		return "file mode"
	case workerMode:
		return "worker mode"
	default:
		return "invalid mode: " + strconv.Itoa(int(m))
	}
}

func (l *Logger) setMode(m mode) { l.mode.Store(uint32(m)) }
func (l *Logger) getMode() mode  { return mode(l.mode.Load()) }

////////////////////////////////////////////////////////////////////////////////
// Javascript Bindings                                                        //
////////////////////////////////////////////////////////////////////////////////

// GetLoggerJS returns the Logger object, used to manager where logging is
// recorded and accessing the log file.
//
// Returns:
//   - A Javascript representation of the [Logger] object.
func GetLoggerJS(js.Value, []js.Value) any {
	return newLoggerJS(GetLogger())
}

// newLoggerJS creates a new Javascript compatible object (map[string]any) that
// matches the [Logger] structure.
func newLoggerJS(lfw *Logger) map[string]any {
	logFileWorker := map[string]any{
		"LogToFile":       js.FuncOf(lfw.LogToFileJS),
		"LogToFileWorker": js.FuncOf(lfw.LogToFileWorkerJS),
		"StopLogging":     js.FuncOf(lfw.StopLoggingJS),
		"GetFile":         js.FuncOf(lfw.GetFileJS),
		"Threshold":       js.FuncOf(lfw.ThresholdJS),
		"MaxSize":         js.FuncOf(lfw.MaxSizeJS),
		"Size":            js.FuncOf(lfw.SizeJS),
		"Worker":          js.FuncOf(lfw.WorkerJS),
	}

	return logFileWorker
}

// LogToFileJS starts logging to a local, in-memory log file.
//
// Parameters:
//   - args[0] - Log level (int).
//   - args[1] - Max log file size, in bytes (int).
//
// Returns:
//   - Throws a TypeError if starting the log file fails.
func (l *Logger) LogToFileJS(_ js.Value, args []js.Value) any {
	threshold := jww.Threshold(args[0].Int())
	maxLogFileSize := args[1].Int()

	err := l.LogToFile(threshold, maxLogFileSize)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// LogToFileWorkerJS starts a new worker that begins listening for logs and
// writing them to file. This function blocks until the worker has started.
//
// Parameters:
//   - args[0] - Log level (int).
//   - args[1] - Max log file size, in bytes (int).
//   - args[2] - Path to Javascript start file for the worker WASM (string).
//   - args[3] - Name of the worker (used in logs) (string).
//
// Returns a promise:
//   - Resolves to nothing on success (void).
//   - Rejected with an error if starting the worker fails.
func (l *Logger) LogToFileWorkerJS(_ js.Value, args []js.Value) any {
	threshold := jww.Threshold(args[0].Int())
	maxLogFileSize := args[1].Int()
	wasmJsPath := args[2].String()
	workerName := args[3].String()

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		err := l.LogToFileWorker(
			threshold, maxLogFileSize, wasmJsPath, workerName)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve()
		}
	}

	return utils.CreatePromise(promiseFn)
}

// StopLoggingJS stops the logging of log messages and disables the log
// listener. If the log worker is running, it is terminated. Once logging is
// stopped, it cannot be resumed the log file cannot be recovered.
func (l *Logger) StopLoggingJS(js.Value, []js.Value) any {
	go l.StopLogging()

	return nil
}

// GetFileJS returns the entire log file.
//
// If the log file is listening locally, it returns it from the local buffer. If
// it is listening from the worker, it blocks until the file is returned.
//
// Returns a promise:
//   - Resolves to the log file contents (string).
func (l *Logger) GetFileJS(js.Value, []js.Value) any {
	promiseFn := func(resolve, _ func(args ...any) js.Value) {
		resolve(string(l.GetFile()))
	}

	return utils.CreatePromise(promiseFn)
}

// ThresholdJS returns the log level threshold used in the file.
//
// Returns:
//   - Log level (int).
func (l *Logger) ThresholdJS(js.Value, []js.Value) any {
	return int(l.Threshold())
}

// MaxSizeJS returns the max size, in bytes, that the log file is allowed to be.
//
// Returns:
//   - Max file size (int).
func (l *Logger) MaxSizeJS(js.Value, []js.Value) any {
	return l.MaxSize()
}

// SizeJS returns the current size, in bytes, written to the log file.
//
// If the log file is listening locally, it returns it from the local buffer. If
// it is listening from the worker, it blocks until the size is returned.
//
// Returns a promise:
//   - Resolves to the current file size (int).
func (l *Logger) SizeJS(js.Value, []js.Value) any {
	promiseFn := func(resolve, _ func(args ...any) js.Value) {
		resolve(l.Size())
	}

	return utils.CreatePromise(promiseFn)
}

// WorkerJS returns the web worker object.
//
// Returns:
//   - Javascript worker object. If the worker has not been initialized, it
//     returns null.
func (l *Logger) WorkerJS(js.Value, []js.Value) any {
	if l.getMode() == workerMode {
		return l.wm.GetWorker()
	}

	return js.Null()
}
