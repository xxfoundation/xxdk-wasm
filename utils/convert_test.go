////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package utils

import (
	"reflect"
	"syscall/js"
	"testing"
)

func TestCopyBytesToGo(t *testing.T) {
	type args struct {
		src js.Value
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CopyBytesToGo(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CopyBytesToGo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopyBytesToJS(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name string
		args args
		want js.Value
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CopyBytesToJS(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CopyBytesToJS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsToJson(t *testing.T) {
	type args struct {
		value js.Value
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JsToJson(tt.args.value); got != tt.want {
				t.Errorf("JsToJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonToJS(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name    string
		args    args
		want    js.Value
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JsonToJS(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("JsonToJS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonToJS() got = %v, want %v", got, tt.want)
			}
		})
	}
}
