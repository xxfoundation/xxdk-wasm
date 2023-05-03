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
	"github.com/armon/circbuf"
	jww "github.com/spf13/jwalterweatherman"
	"math/rand"
	"reflect"
	"testing"
)

func Test_newFileLogger(t *testing.T) {
	expected := &fileLogger{
		threshold:      jww.LevelError,
		maxLogFileSize: 512,
	}
	expected.cb, _ = circbuf.NewBuffer(int64(expected.maxLogFileSize))
	fl, err := newFileLogger(expected.threshold, expected.maxLogFileSize)
	if err != nil {
		t.Fatalf("Failed to make new fileLogger: %+v", err)
	}

	if !reflect.DeepEqual(expected, fl) {
		t.Errorf("Unexpected new fileLogger.\nexpected: %+v\nreceived: %+v",
			expected, fl)
	}
	if !reflect.DeepEqual(logger, fl) {
		t.Errorf("Failed to set logger.\nexpected: %+v\nreceived: %+v",
			logger, fl)
	}

}

// Tests that fileLogger.Write writes the expected data to the buffer and that
// when the max file size is reached, old data is replaced.
func Test_fileLogger_Write(t *testing.T) {
	rng := rand.New(rand.NewSource(3424))
	fl, err := newFileLogger(jww.LevelError, 512)
	if err != nil {
		t.Fatalf("Failed to make new fileLogger: %+v", err)
	}

	expected := make([]byte, fl.maxLogFileSize)
	rng.Read(expected)
	n, err := fl.Write(expected)
	if err != nil {
		t.Fatalf("Failed to write: %+v", err)
	} else if n != len(expected) {
		t.Fatalf("Did not write expected length.\nexpected: %d\nreceived: %d",
			len(expected), n)
	}

	if !bytes.Equal(fl.cb.Bytes(), expected) {
		t.Fatalf("Incorrect bytes in buffer.\nexpected: %v\nreceived: %v",
			expected, fl.cb.Bytes())
	}

	// Check that the data is overwritten
	rng.Read(expected)
	n, err = fl.Write(expected)
	if err != nil {
		t.Fatalf("Failed to write: %+v", err)
	} else if n != len(expected) {
		t.Fatalf("Did not write expected length.\nexpected: %d\nreceived: %d",
			len(expected), n)
	}

	if !bytes.Equal(fl.cb.Bytes(), expected) {
		t.Fatalf("Incorrect bytes in buffer.\nexpected: %v\nreceived: %v",
			expected, fl.cb.Bytes())
	}
}

func Test_fileLogger_Listen(t *testing.T) {
}

func Test_fileLogger_StopLogging(t *testing.T) {
}

func Test_fileLogger_GetFile(t *testing.T) {
}

func Test_fileLogger_Threshold(t *testing.T) {
}

func Test_fileLogger_MaxSize(t *testing.T) {
}

func Test_fileLogger_Size(t *testing.T) {
}

func Test_fileLogger_Worker(t *testing.T) {
}
