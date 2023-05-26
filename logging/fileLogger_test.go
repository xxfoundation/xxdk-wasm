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
		threshold: jww.LevelError,
	}
	expected.cb, _ = circbuf.NewBuffer(512)
	fl, err := newFileLogger(expected.threshold, int(expected.cb.Size()))
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

	expected := make([]byte, fl.MaxSize())
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

// Tests that fileLogger.Listen only returns an io.Writer for valid thresholds.
func Test_fileLogger_Listen(t *testing.T) {
	th := jww.LevelError
	fl, err := newFileLogger(th, 512)
	if err != nil {
		t.Fatalf("Failed to make new fileLogger: %+v", err)
	}

	thresholds := []jww.Threshold{-1, jww.LevelTrace, jww.LevelDebug,
		jww.LevelFatal, jww.LevelWarn, jww.LevelError, jww.LevelCritical,
		jww.LevelFatal}

	for _, threshold := range thresholds {
		w := fl.Listen(threshold)
		if threshold < th {
			if w != nil {
				t.Errorf("Did not receive nil io.Writer for level %s: %+v",
					threshold, w)
			}
		} else if w == nil {
			t.Errorf("Received nil io.Writer for level %s", threshold)
		}
	}
}

// Tests that fileLogger.Listen always returns nil after fileLogger.StopLogging
// is called.
func Test_fileLogger_StopLogging(t *testing.T) {
	fl, err := newFileLogger(jww.LevelError, 512)
	if err != nil {
		t.Fatalf("Failed to make new fileLogger: %+v", err)
	}

	fl.StopLogging()

	if w := fl.Listen(jww.LevelFatal); w != nil {
		t.Errorf("Listen returned non-nil io.Writer when logging should have "+
			"been stopped: %+v", w)
	}

	file := fl.GetFile()
	if !bytes.Equal([]byte{}, file) {
		t.Errorf("Did not receice empty file: %+v", file)
	}
}

// Tests that fileLogger.GetFile returns the expected file.
func Test_fileLogger_GetFile(t *testing.T) {
	rng := rand.New(rand.NewSource(9863))
	fl, err := newFileLogger(jww.LevelError, 512)
	if err != nil {
		t.Fatalf("Failed to make new fileLogger: %+v", err)
	}

	var expected []byte
	for i := 0; i < 5; i++ {
		p := make([]byte, rng.Intn(64))
		rng.Read(p)
		expected = append(expected, p...)

		if _, err = fl.Write(p); err != nil {
			t.Errorf("Write %d failed: %+v", i, err)
		}
	}

	file := fl.GetFile()
	if !bytes.Equal(expected, file) {
		t.Errorf("Unexpected file.\nexpected: %v\nreceived: %v", expected, file)
	}
}

// Tests that fileLogger.Threshold returns the expected threshold.
func Test_fileLogger_Threshold(t *testing.T) {
	thresholds := []jww.Threshold{-1, jww.LevelTrace, jww.LevelDebug,
		jww.LevelFatal, jww.LevelWarn, jww.LevelError, jww.LevelCritical,
		jww.LevelFatal}

	for _, threshold := range thresholds {
		fl, err := newFileLogger(threshold, 512)
		if err != nil {
			t.Fatalf("Failed to make new fileLogger: %+v", err)
		}

		if fl.Threshold() != threshold {
			t.Errorf("Incorrect threshold.\nexpected: %s (%d)\nreceived: %s (%d)",
				threshold, threshold, fl.Threshold(), fl.Threshold())
		}
	}
}

// Unit test of fileLogger.MaxSize.
func Test_fileLogger_MaxSize(t *testing.T) {
	maxSize := 512
	fl, err := newFileLogger(jww.LevelError, maxSize)
	if err != nil {
		t.Fatalf("Failed to make new fileLogger: %+v", err)
	}

	if fl.MaxSize() != maxSize {
		t.Errorf("Incorrect max size.\nexpected: %d\nreceived: %d",
			maxSize, fl.MaxSize())
	}
}

// Unit test of fileLogger.Size.
func Test_fileLogger_Size(t *testing.T) {
	rng := rand.New(rand.NewSource(9863))
	fl, err := newFileLogger(jww.LevelError, 512)
	if err != nil {
		t.Fatalf("Failed to make new fileLogger: %+v", err)
	}

	var expected []byte
	for i := 0; i < 5; i++ {
		p := make([]byte, rng.Intn(64))
		rng.Read(p)
		expected = append(expected, p...)

		if _, err = fl.Write(p); err != nil {
			t.Errorf("Write %d failed: %+v", i, err)
		}

		size := fl.Size()
		if size != len(expected) {
			t.Errorf("Incorrect size (%d).\nexpected: %d\nreceived: %d",
				i, len(expected), size)
		}
	}

	file := fl.GetFile()
	if !bytes.Equal(expected, file) {
		t.Errorf("Unexpected file.\nexpected: %v\nreceived: %v", expected, file)
	}
}

// Tests that fileLogger.Worker always returns nil.
func Test_fileLogger_Worker(t *testing.T) {
	fl, err := newFileLogger(jww.LevelError, 512)
	if err != nil {
		t.Fatalf("Failed to make new fileLogger: %+v", err)
	}

	w := fl.Worker()
	if w != nil {
		t.Errorf("Did not get nil worker: %+v", w)
	}
}
