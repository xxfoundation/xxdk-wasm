////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package logging

import (
	"bytes"
	"fmt"
	jww "github.com/spf13/jwalterweatherman"
	"testing"
)

// Tests InitLogger
func TestInitLogger(t *testing.T) {
}

// Tests GetLogger
func TestGetLogger(t *testing.T) {
}

// Tests NewLogger
func TestNewLogger(t *testing.T) {
}

// Tests Logger.LogToFile
func TestLogger_LogToFile(t *testing.T) {
	jww.SetStdoutThreshold(jww.LevelTrace)
	l := NewLogger()

	err := l.LogToFile(jww.LevelTrace, 50000000)
	if err != nil {
		t.Fatalf("Failed to LogToFile: %+v", err)
	}

	jww.INFO.Printf("test")

	file := l.cb.Bytes()
	fmt.Printf("file:----------------------------\n%s\n---------------------------------\n", file)
}

// Tests Logger.LogToFileWorker
func TestLogger_LogToFileWorker(t *testing.T) {
}

// Tests Logger.processLog
func TestLogger_processLog(t *testing.T) {
}

// Tests Logger.prepare
func TestLogger_prepare(t *testing.T) {
}

// Tests Logger.StopLogging
func TestLogger_StopLogging(t *testing.T) {
}

// Tests Logger.GetFile
func TestLogger_GetFile(t *testing.T) {
}

// Tests Logger.Threshold
func TestLogger_Threshold(t *testing.T) {
}

// Tests Logger.MaxSize
func TestLogger_MaxSize(t *testing.T) {
}

// Tests Logger.Size
func TestLogger_Size(t *testing.T) {
}

// Tests Logger.Listen
func TestLogger_Listen(t *testing.T) {

	// l := newLogger()

}

// Tests that Logger.Write can fill the listenChan channel completely and that
// all messages are received in the order they were added.
func TestLogger_Write(t *testing.T) {
	l := newLogger()
	expectedLogs := make([][]byte, logListenerChanSize)

	for i := range expectedLogs {
		p := []byte(
			fmt.Sprintf("Log message %d of %d.", i+1, logListenerChanSize))
		expectedLogs[i] = p
		n, err := l.Listen(jww.LevelError).Write(p)
		if err != nil {
			t.Errorf("Received impossible error (%d): %+v", i, err)
		} else if n != len(p) {
			t.Errorf("Received incorrect bytes written (%d)."+
				"\nexpected: %d\nreceived: %d", i, len(p), n)
		}
	}

	for i, expected := range expectedLogs {
		select {
		case received := <-l.listenChan:
			if !bytes.Equal(expected, received) {
				t.Errorf("Received unexpected meessage (%d)."+
					"\nexpected: %q\nreceived: %q", i, expected, received)
			}
		default:
			t.Errorf("Failed to read from channel.")
		}
	}
}

// Error path: Tests that Logger.Write returns an error when the listener
// channel is full.
func TestLogger_Write_ChannelFilledError(t *testing.T) {
	l := newLogger()
	expectedLogs := make([][]byte, logListenerChanSize)

	for i := range expectedLogs {
		p := []byte(
			fmt.Sprintf("Log message %d of %d.", i+1, logListenerChanSize))
		expectedLogs[i] = p
		n, err := l.Listen(jww.LevelError).Write(p)
		if err != nil {
			t.Errorf("Received impossible error (%d): %+v", i, err)
		} else if n != len(p) {
			t.Errorf("Received incorrect bytes written (%d)."+
				"\nexpected: %d\nreceived: %d", i, len(p), n)
		}
	}

	_, err := l.Write([]byte("test"))
	if err == nil {
		t.Error("Failed to receive error when the chanel should be full.")
	}
}

// Tests that Logger.getMode gets the same value set with Logger.setMode.
func TestLogger_setMode_getMode(t *testing.T) {
	l := newLogger()

	for i, m := range []mode{initMode, fileMode, workerMode, 12} {
		l.setMode(m)
		received := l.getMode()
		if m != received {
			t.Errorf("Received wrong mode (%d).\nexpected: %s\nreceived: %s",
				i, m, received)
		}
	}

}

// Unit test of mode.String.
func Test_mode_String(t *testing.T) {
	for m, expected := range map[mode]string{
		initMode:   "uninitialized mode",
		fileMode:   "file mode",
		workerMode: "worker mode",
		12:         "invalid mode: 12",
	} {
		s := m.String()
		if s != expected {
			t.Errorf("Wrong string for mode %d.\nexpected: %s\nreceived: %s",
				m, expected, s)
		}
	}
}
