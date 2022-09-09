////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package utils

import (
	"github.com/pkg/errors"
	"syscall/js"
	"testing"
)

func TestJsError(t *testing.T) {
	err := errors.Errorf("test error")

	jsError := JsError(err)

	t.Logf("%+v", jsError)
	t.Logf("%+v", jsError.String())
	t.Logf("%+v", js.Error{Value: jsError})
}

func TestJsTrace(t *testing.T) {
}

func TestThrow(t *testing.T) {
}

func Test_throw(t *testing.T) {
}
